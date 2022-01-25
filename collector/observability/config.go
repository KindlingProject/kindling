package observability

type Config struct {
	Port string `mapstructure:"prometheus_port"`
}

var DefaultConfig = Config{
	Port: ":8081",
}
