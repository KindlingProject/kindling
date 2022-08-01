package cpuanalyzer

const (
	defaultSegmentSize = 40
)

type Config struct {
	SegmentSize int    `mapstructure:"segment_size"`
	EsHost      string `mapstructure:"es_host"`
}

func (cfg *Config) GetSegmentSize() int {
	if cfg.SegmentSize > 0 {
		return cfg.SegmentSize
	} else {
		return defaultSegmentSize
	}
}

func (cfg *Config) GetEsHost() string {
	if cfg.EsHost == "" {
		return "http://39.103.171.51:8080"
	} else {
		return cfg.EsHost
	}
}
