package cpuanalyzer

const (
	defaultSegmentSize = 40
)

type Config struct {
	SegmentSize       int `mapstructure:"segment_size"`
	javaTraceSlowTime int `mapstructure:"java_trace_slow_time"`
}

func (cfg *Config) GetSegmentSize() int {
	if cfg.SegmentSize > 0 {
		return cfg.SegmentSize
	} else {
		return defaultSegmentSize
	}
}
