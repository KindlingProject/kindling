package kubernetes

// config contains optional settings for connecting to kubernetes.
type config struct {
	KubeAuthType  AuthType
	KubeConfigDir string
}

type Option func(cfg *config)

// WithAuthType sets way of authenticating kubernetes api-server
// Supported AuthTypeNone, AuthTypeServiceAccount, AuthTypeKubeConfig
func WithAuthType(authType AuthType) Option {
	return func(cfg *config) {
		cfg.KubeAuthType = authType
	}
}

// WithKubeConfigDir sets the directory where the file "kubeconfig" is stored
func WithKubeConfigDir(dir string) Option {
	return func(cfg *config) {
		cfg.KubeConfigDir = dir
	}
}
