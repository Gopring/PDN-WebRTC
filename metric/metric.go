// Package metric provides Prometheus metrics collection and monitoring.
package metric

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var (
	ErrNetworkNotFound = errors.New("no network stats found")
)

// Metrics contains the Prometheus metrics server and registered custom metrics.
type Metrics struct {
	httpServer           *http.Server
	config               Config
	webSocketConnections prometheus.Gauge
	webRTCConnections    prometheus.Gauge
	cpuUsage             prometheus.Gauge
	memoryUsage          prometheus.Gauge
	networkUsage         *prometheus.GaugeVec

	clientConnectionAttempts  prometheus.Counter
	clientConnectionSuccesses prometheus.Counter
	clientConnectionFailures  prometheus.Counter

	pushConnections prometheus.Gauge
	pullConnections prometheus.Gauge
	peerConnections prometheus.Gauge

	balancingOccurs prometheus.Counter
}

// New creates a new Metrics instance with the specified configuration.
func New(c Config) *Metrics {
	return &Metrics{
		config: c,
		webSocketConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "websocket_connections_total",
			Help: "Current number of WebSocket connections.",
		}),
		webRTCConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "webrtc_connections_total",
			Help: "Current number of WebRTC connections.",
		}),
		cpuUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cpu_usage_percentage",
			Help: "CPU usage percentage.",
		}),
		memoryUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Current memory usage in bytes.",
		}),
		networkUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "network_usage_bytes",
			Help: "Current network usage in bytes.",
		}, []string{"direction"}), // Direction: "inbound" or "outbound"
		clientConnectionAttempts: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "client_connection_attempts_total",
			Help: "Total number of client connection attempts.",
		}),
		clientConnectionSuccesses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "client_connection_successes_total",
			Help: "Total number of successful client connections.",
		}),
		clientConnectionFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "client_connection_failures_total",
			Help: "Total number of failed client connections.",
		}),
		pushConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "push_connections_total",
			Help: "Current number of push connections.",
		}),
		pullConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pull_connections_total",
			Help: "Current number of pull connections.",
		}),
		peerConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "peer_connections_total",
			Help: "Current number of peer connections.",
		}),
		balancingOccurs: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "balancing_occurs_total",
			Help: "Total number of load balancing occurrences.",
		}),
	}
}

// RegisterMetrics registers custom metrics with Prometheus.
func (m *Metrics) registerMetrics() {
	prometheus.MustRegister(m.webSocketConnections)
	prometheus.MustRegister(m.webRTCConnections)
	prometheus.MustRegister(m.cpuUsage)
	prometheus.MustRegister(m.memoryUsage)
	prometheus.MustRegister(m.networkUsage)
	prometheus.MustRegister(m.clientConnectionAttempts)
	prometheus.MustRegister(m.clientConnectionSuccesses)
	prometheus.MustRegister(m.clientConnectionFailures)
	prometheus.MustRegister(m.pushConnections)
	prometheus.MustRegister(m.pullConnections)
	prometheus.MustRegister(m.peerConnections)
	prometheus.MustRegister(m.balancingOccurs)
}

// Start initializes and starts the metrics HTTP server.
func (m *Metrics) Start() {
	m.registerMetrics()
	m.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", m.config.Port),
		Handler:           promhttp.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go m.UpdateSystemMetrics()
	log.Printf("Starting metrics server on port %d at path %s", m.config.Port, m.config.Path)
	if err := m.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Error starting metrics server: %v", err)
	}
}

// Stop gracefully shuts down the metrics server.
func (m *Metrics) Stop() error {
	if m.httpServer != nil {
		log.Printf("Stopping metrics server on port %d", m.config.Port)
		return m.httpServer.Close()
	}
	return nil
}

// UpdateSystemMetrics collects and updates system-level metrics (e.g., memory usage).
func (m *Metrics) UpdateSystemMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	netStats, err := net.IOCounters(false)
	prev := netStats[0]
	if err != nil {
		panic(err)
	}
	for {
		select {
		case <-ticker.C:
			prev, err = m.collectMetrics(prev)
			//case <-stop:
			//	log.Println("Stopping metrics collection")
			//	return
		}
	}
}

// collectMetrics collects individual system metrics.
func (m *Metrics) collectMetrics(prevNetStats net.IOCountersStat) (net.IOCountersStat, error) {
	// Collect CPU usage
	percentages, err := cpu.Percent(1*time.Second, false)
	if err != nil && len(percentages) == 0 {
		return net.IOCountersStat{}, fmt.Errorf("error fetching CPU usage: %v", err)
	}
	m.cpuUsage.Set(percentages[0])

	// Collect memory usage
	vmStats, err := mem.VirtualMemory()
	if err != nil {
		return net.IOCountersStat{}, fmt.Errorf("error fetching memory usage: %v", err)
	}
	shortMemory := float64(vmStats.Used) / (1 << 20) // Convert to MB
	m.memoryUsage.Set(shortMemory)

	// Collect network usage
	curNetStats, err := net.IOCounters(false)
	if err != nil {
		return net.IOCountersStat{}, fmt.Errorf("error fetching network usage: %v", err)
	} else if len(curNetStats) == 0 {
		return net.IOCountersStat{}, fmt.Errorf("error fetching network usage: %v", ErrNetworkNotFound)
	}

	receive := (float64(curNetStats[0].BytesRecv) - float64(prevNetStats.BytesRecv)) / (1 << 20) // Convert to MB
	sent := (float64(curNetStats[0].BytesSent) - float64(prevNetStats.BytesSent)) / (1 << 20)    // Convert to MB
	m.UpdateNetworkUsage("inbound", receive)
	m.UpdateNetworkUsage("outbound", sent)
	return curNetStats[0], nil
}

// IncrementWebSocketConnections increments the WebSocket connection count.
func (m *Metrics) IncrementWebSocketConnections() {
	m.webSocketConnections.Inc()
}

// DecrementWebSocketConnections decrements the WebSocket connection count.
func (m *Metrics) DecrementWebSocketConnections() {
	m.webSocketConnections.Dec()
}

// IncrementWebRTCConnections increments the WebRTC connection count.
func (m *Metrics) IncrementWebRTCConnections() {
	m.webRTCConnections.Inc()
}

// DecrementWebRTCConnections decrements the WebRTC connection count.
func (m *Metrics) DecrementWebRTCConnections() {
	m.webRTCConnections.Dec()
}

// UpdateNetworkUsage updates network usage metrics (e.g., inbound and outbound traffic).
func (m *Metrics) UpdateNetworkUsage(direction string, bytes float64) {
	m.networkUsage.WithLabelValues(direction).Set(bytes)
}

// IncrementClientConnectionAttempts increments the client connection attempts counter.
func (m *Metrics) IncrementClientConnectionAttempts() {
	m.clientConnectionAttempts.Inc()
}

// IncrementClientConnectionSuccesses increments the client connection successes counter.
func (m *Metrics) IncrementClientConnectionSuccesses() {
	m.clientConnectionSuccesses.Inc()
}

// IncrementClientConnectionFailures increments the client connection failures counter.
func (m *Metrics) IncrementClientConnectionFailures() {
	m.clientConnectionFailures.Inc()
}

// IncrementPeerConnections increments the number of forwarder connections by 1.
func (m *Metrics) IncrementPeerConnections() {
	m.peerConnections.Inc()
}

// DecrementPeerConnections decrements the number of forwarder connections by 1.
func (m *Metrics) DecrementPeerConnections() {
	m.peerConnections.Dec()
}

// IncrementPushConnections increments the number of push connections by 1.
func (m *Metrics) IncrementPushConnections() {
	m.pushConnections.Inc()
}

// DecrementPushConnections decrements the number of push connections by 1.
func (m *Metrics) DecrementPushConnections() {
	m.pushConnections.Dec()
}

// IncrementPullConnections increments the number of pull connections by 1.
func (m *Metrics) IncrementPullConnections() {
	m.pullConnections.Inc()
}

// DecrementPullConnections decrements the number of pull connections by 1.
func (m *Metrics) DecrementPullConnections() {
	m.pullConnections.Dec()
}

// IncrementBalancingOccurs increments the number of load balancing occurrences by 1.
func (m *Metrics) IncrementBalancingOccurs() {
	m.balancingOccurs.Inc()
}
