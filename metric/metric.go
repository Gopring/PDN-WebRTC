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

// Metrics contains the Prometheus metrics server and registered custom metrics.
type Metrics struct {
	httpServer           *http.Server
	config               Config
	webSocketConnections prometheus.Gauge
	webRTCConnections    prometheus.Gauge
	cpuUsage             prometheus.Gauge
	memoryUsage          prometheus.Gauge
	networkUsage         *prometheus.GaugeVec
}

// Config defines the configuration for the metrics server.
type Config struct {
	Port int    // Port for metrics server
	Path string // Path for metrics endpoint
}

// Default values for metrics configuration.
const (
	DefaultMetricsPort = 9090
	DefaultMetricsPath = "/metrics"
)

// New creates a new Metrics instance with the specified configuration.
func New(config Config) *Metrics {
	return &Metrics{
		config: config,
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
	}
}

// RegisterMetrics registers custom metrics with Prometheus.
func (m *Metrics) RegisterMetrics() {
	prometheus.MustRegister(m.webSocketConnections)
	prometheus.MustRegister(m.webRTCConnections)
	prometheus.MustRegister(m.cpuUsage)
	prometheus.MustRegister(m.memoryUsage)
	prometheus.MustRegister(m.networkUsage)
}

// Start initializes and starts the metrics HTTP server.
func (m *Metrics) Start(stop <-chan struct{}) {
	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.config.Port),
		Handler: promhttp.Handler(),
	}

	go m.UpdateSystemMetrics(stop)
	go func() {
		log.Printf("Starting metrics server on port %d at path %s", m.config.Port, m.config.Path)
		if err := m.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error starting metrics server: %v", err)
		}
	}()
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
func (m *Metrics) UpdateSystemMetrics(stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.collectMetrics()
			case <-stop:
				log.Println("Stopping metrics collection")
				return
			}
		}
	}()
}

// collectMetrics collects individual system metrics.
func (m *Metrics) collectMetrics() {
	// Collect CPU usage
	if percentages, err := cpu.Percent(1*time.Second, false); err == nil && len(percentages) > 0 {
		m.cpuUsage.Set(percentages[0])
		log.Printf("CPU usage updated: %.2f%%", percentages[0])
	} else {
		log.Printf("Error fetching CPU usage: %v", err)
	}

	// Collect memory usage
	if vmStats, err := mem.VirtualMemory(); err == nil {
		m.memoryUsage.Set(float64(vmStats.Used))
		log.Printf("Memory usage updated: %v bytes", vmStats.Used)
	} else {
		log.Printf("Error fetching memory usage: %v", err)
	}

	// Collect network usage
	if ioStats, err := net.IOCounters(false); err == nil && len(ioStats) > 0 {
		totalRecv, totalSent := 0.0, 0.0
		for _, stat := range ioStats {
			totalRecv += float64(stat.BytesRecv)
			totalSent += float64(stat.BytesSent)
		}
		m.UpdateNetworkUsage("inbound", totalRecv)
		m.UpdateNetworkUsage("outbound", totalSent)
		log.Printf("Network usage updated: Inbound=%.0f bytes, Outbound=%.0f bytes", totalRecv, totalSent)
	} else {
		log.Printf("Error fetching network usage: %v", err)
	}
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
