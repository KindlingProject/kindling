package k8sinfoanalyzer

type Config struct {
	// send_datagroup_interval is the datagroup sending interval.
	// The unit is seconds.
	SendDataGroupInterval int `mapstructure:"send_datagroup_interval"`
}

func NewDefaultConfig() *Config {
	return &Config{
		SendDataGroupInterval: 15,
	}
}
