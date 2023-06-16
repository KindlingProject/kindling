package k8sinfoanalyzer

type Config struct {
	// SendDataGroupInterval is the datagroup sending interval.
	// The unit is seconds.
	SendDataGroupInterval int `mapstructure:"send_datagroup_interval"`
}

func NewDefaultConfig() *Config {
	return &Config{
		SendDataGroupInterval: 15,
	}
}
