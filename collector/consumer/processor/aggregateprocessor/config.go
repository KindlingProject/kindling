package aggregateprocessor

type Config struct {
	// TODO: Expose filters to configuration
	// Unable to work now.
	FilterLabels []string `mapstructure:"filter_labels"`
	// The unit is second.
	TickerInterval int `mapstructure:"ticker_interval"`

	AggregateKindMap map[string][]string `mapstructure:"aggregate_kind_map"`
	SamplingRate     *SampleConfig       `mapstructure:"sampling_rate"`
}

type SampleConfig struct {
	NormalData int `mapstructure:"normal_data"`
	SlowData   int `mapstructure:"slow_data"`
	ErrorData  int `mapstructure:"error_data"`
}

func NewDefaultConfig() *Config {
	ret := &Config{
		FilterLabels:   make([]string, 0),
		TickerInterval: 5,
		AggregateKindMap: map[string][]string{
			"request_count":      {"sum"},
			"request_total_time": {"sum", "avg"},
			"request_io":         {"sum"},
			"response_io":        {"sum"},
			// tcp
			"kindling_tcp_rtt_microseconds":  {"last"},
			"kindling_tcp_retransmit_total":  {"sum"},
			"kindling_tcp_packet_loss_total": {"sum"},
		},
		SamplingRate: &SampleConfig{
			NormalData: 0,
			SlowData:   100,
			ErrorData:  100,
		},
	}
	return ret
}
