package tcpconnectanalyzer

type Config struct {
	ChannelSize     int `mapstructure:"channel_size"`
	WaitEventSecond int `mapstructure:"wait_event_second"`
}

func NewDefaultConfig() *Config {
	return &Config{
		ChannelSize:     2000,
		WaitEventSecond: 10,
	}
}
