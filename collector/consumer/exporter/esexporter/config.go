package esexporter

type Config struct {
	EsHost string `mapstructure:"es_host"`
}

func (cfg *Config) GetEsHost() string {
	if cfg.EsHost == "" {
		return "http://39.103.171.51:8080"
	} else {
		return cfg.EsHost
	}
}
