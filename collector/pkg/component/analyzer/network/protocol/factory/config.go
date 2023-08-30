package factory

type config struct {
	urlClusteringMethod string
	ignoreDnsRcode3Error bool
}

func newDefaultConfig() *config {
	return &config{
		urlClusteringMethod: "alphabet",
		ignoreDnsRcode3Error: false,
	}
}

type Option func(cfg *config)

func WithUrlClusteringMethod(urlClusteringMethod string) Option {
	return func(cfg *config) {
		cfg.urlClusteringMethod = urlClusteringMethod
	}
}

func WithIgnoreDnsRcode3Error(ignoreDnsRcode3Error bool) Option {
	return func(cfg *config) {
		cfg.ignoreDnsRcode3Error = ignoreDnsRcode3Error
	}
}
