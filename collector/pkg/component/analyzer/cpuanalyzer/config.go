package cpuanalyzer

const (
	defaultSegmentSize = 40
)

type Config struct {
	javaTraceSlowTime int `mapstructure:"java_trace_slow_time"`
	// SegmentSize defines how many segments(seconds) can be cached to wait for sending.
	// The elder segments will be overwritten by the newer ones, so don't set it too low.
	SegmentSize int `mapstructure:"segment_size"`
	// EdgeEventsWindowSize is the size of the duration window that seats the edge events.
	// The unit is seconds. The greater it is, the more data will be stored.
	EdgeEventsWindowSize int `mapstructure:"edge_events_window_size"`
}

func NewDefaultConfig() *Config {
	return &Config{
		SegmentSize:          40,
		EdgeEventsWindowSize: 2,
	}
}
