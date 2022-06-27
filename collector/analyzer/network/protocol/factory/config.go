package factory

type config struct {
	urlClusteringMethod string
}

func newDefaultConfig() *config {
	return &config{
		urlClusteringMethod: "alphabet",
	}
}

type Option func(cfg *config)

func WithUrlClusteringMethod(urlClusteringMethod string) Option {
	return func(cfg *config) {
		cfg.urlClusteringMethod = urlClusteringMethod
	}
}
