package classifier

import "time"

// DefaultTimeoutDuration defines the default timeout duration for classification operations.
const (
	DefaultTimeoutDuration  = 10 * time.Second
	DefaultCronJobFrequency = 60 * time.Second
)

// Config holds configuration options for the Classifier.
type Config struct {
	TimeoutDuration  time.Duration
	CronJobFrequency time.Duration
}
