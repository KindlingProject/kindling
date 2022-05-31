package tcpconnectanalyzer

type Config struct {
	ChannelSize   int `mapstructure:"channel_size"`
	TimeoutSecond int `mapstructure:"timeout_second"`
}

func NewDefaultConfig() *Config {
	return &Config{
		ChannelSize:   2000,
		TimeoutSecond: 60,
	}
}
