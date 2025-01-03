package metric

// Default values for metrics configuration.
const (
	DefaultMetricsPort = 9090
	DefaultMetricsPath = "/metrics"
)

// Config defines the configuration for the metrics server.
type Config struct {
	Port int    // Port for metrics server
	Path string // Path for metrics endpoint
}
