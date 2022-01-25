package uprobeanalyzer

type Config struct {
	// unit: millisecond
	ResponseSlowThreshold uint64 `mapstructure:"response_slow_threshold"`
}
