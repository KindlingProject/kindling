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
