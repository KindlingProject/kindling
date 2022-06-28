package cgoreceiver

type Config struct {
	SubscribeInfo []SubEvent `mapstructure:"subscribe"`
}

type SubEvent struct {
	Category string `mapstructure:"category"`
	Name     string `mapstructure:"name"`
}
