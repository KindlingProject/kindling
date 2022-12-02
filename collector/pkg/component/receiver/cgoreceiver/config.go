package cgoreceiver

type Config struct {
	SubscribeInfo []SubEvent `mapstructure:"subscribe"`
}

type SubEvent struct {
	Category string            `mapstructure:"category"`
	Name     string            `mapstructure:"name"`
	Params   map[string]string `mapstructure:"params"`
}
