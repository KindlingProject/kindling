package cgoreceiver

type Config struct {
	SubscribeInfo     []SubEvent    `mapstructure:"subscribe"`
	ProcessFilterInfo ProcessFilter `mapstructure:"process_filter"`
}

type SubEvent struct {
	Category string            `mapstructure:"category"`
	Name     string            `mapstructure:"name"`
	Params   map[string]string `mapstructure:"params"`
}

type ProcessFilter struct {
	Comms []string `mapstructure:"comms"`
}

func NewDefaultConfig() *Config {
	return &Config{
		ProcessFilterInfo: ProcessFilter{
			Comms: []string{"kindling-collec", "containerd", "dockerd", "containerd-shim"},
		},
	}
}
