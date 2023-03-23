package kubernetes

import "time"

// config contains optional settings for connecting to kubernetes.
type config struct {
	KubeAuthType  AuthType
	KubeConfigDir string
	// GraceDeletePeriod controls the delay interval after receiving delete event.
	// The unit is seconds, and the default value is 60 seconds.
	// Should not be lower than 30 seconds.
	GraceDeletePeriod time.Duration
	// EnableFetchReplicaSet controls whether to fetch ReplicaSet information.
	// The default value is false. It should be enabled if the ReplicaSet
	// is used to control pods in the third-party CRD except for Deployment.
	EnableFetchReplicaSet bool
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

// WithGraceDeletePeriod sets the graceful period of deleting Pod's metadata
// after receiving "delete" event from client-go.
func WithGraceDeletePeriod(interval int) Option {
	return func(cfg *config) {
		cfg.GraceDeletePeriod = time.Duration(interval) * time.Second
	}
}

// WithFetchReplicaSet sets whether to fetch ReplicaSet information.
func WithFetchReplicaSet(fetch bool) Option {
	return func(cfg *config) {
		cfg.EnableFetchReplicaSet = fetch
	}
}
