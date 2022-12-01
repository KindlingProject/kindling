package network

const (
	defaultRequestTimeout        = 15
	defaultNoResponseThreshold   = 120
	defaultConnectTimeout        = 1
	defaultResponseSlowThreshold = 500
)

type Config struct {
	ConnectTimeout int `mapstructure:"connect_timeout"`
	RequestTimeout int `mapstructure:"request_timeout"`
	// unit is ms
	ResponseSlowThreshold int `mapstructure:"response_slow_threshold"`

	EnableConntrack       bool   `mapstructure:"enable_conntrack"`
	ConntrackMaxStateSize int    `mapstructure:"conntrack_max_state_size"`
	ConntrackRateLimit    int    `mapstructure:"conntrack_rate_limit"`
	ProcRoot              string `mapstructure:"proc_root"`

	ProtocolParser      []string         `mapstructure:"protocol_parser"`
	ProtocolConfigs     []ProtocolConfig `mapstructure:"protocol_config,omitempty"`
	UrlClusteringMethod string           `mapstructure:"url_clustering_method"`
	NoResponseThreshold int              `mapstructure:"no_response_threshold"`
}

func NewDefaultConfig() *Config {
	return &Config{
		ConnectTimeout:        100,
		RequestTimeout:        60,
		ResponseSlowThreshold: 500,
		EnableConntrack:       true,
		ConntrackMaxStateSize: 131072,
		ConntrackRateLimit:    500,
		ProcRoot:              "/proc",
		ProtocolParser:        []string{"http", "mysql", "dns", "redis", "kafka", "dubbo"},
		ProtocolConfigs: []ProtocolConfig{
			{
				Key:           "http",
				PayloadLength: 200,
			},
			{
				Key:           "dubbo",
				PayloadLength: 200,
			},
			{
				Key:       "mysql",
				Ports:     []uint32{3306},
				Threshold: 100,
			},
			{
				Key:       "kafka",
				Ports:     []uint32{9092},
				Threshold: 100,
			},
			{
				Key:       "dns",
				Ports:     []uint32{53},
				Threshold: 100,
			},
			{
				Key:       "cassandra",
				Ports:     []uint32{9042},
				Threshold: 100,
			},
			{
				Key:       "s3",
				Ports:     []uint32{9190},
				Threshold: 100,
			},
		},
		UrlClusteringMethod: "alphabet",
		NoResponseThreshold: 120,
	}
}

type ProtocolConfig struct {
	Key            string   `mapstructure:"key,omitempty"`
	Ports          []uint32 `mapstructure:"ports,omitempty"`
	PayloadLength  int      `mapstructure:"payload_length"`
	DisableDiscern bool     `mapstructure:"disable_discern,omitempty"`
	Threshold      int      `mapstructure:"slow_threshold,omitempty"`
}

func (cfg *Config) GetConnectTimeout() int {
	if cfg.ConnectTimeout > 0 {
		return cfg.ConnectTimeout
	} else {
		return defaultConnectTimeout
	}
}

func (cfg *Config) GetRequestTimeout() int {
	if cfg.RequestTimeout > 0 {
		return cfg.RequestTimeout
	} else {
		return defaultRequestTimeout
	}
}

func (cfg *Config) getResponseSlowThreshold() int {
	if cfg.ResponseSlowThreshold > 0 {
		return cfg.ResponseSlowThreshold
	} else {
		return defaultResponseSlowThreshold
	}
}

func (cfg *Config) getNoResponseThreshold() int {
	if cfg.NoResponseThreshold > 0 {
		return cfg.NoResponseThreshold
	} else {
		return defaultNoResponseThreshold
	}
}
