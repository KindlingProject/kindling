package k8sprocessor

import (
	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
)

type Config struct {
	KubeAuthType  kubernetes.AuthType `mapstructure:"kube_auth_type"`
	KubeConfigDir string              `mapstructure:"kube_config_dir"`
	// GraceDeletePeriod controls the delay interval after receiving delete event.
	// The unit is seconds, and the default value is 60 seconds.
	// Should not be lower than 30 seconds.
	GraceDeletePeriod int `mapstructure:"grace_delete_period"`
	// Disable is used when the agent runs at the non-Kubernetes environment.
	Disable bool
}

var DefaultConfig Config = Config{
	KubeAuthType:      "serviceAccount",
	KubeConfigDir:     "~/.kube/config",
	GraceDeletePeriod: 60,
	Disable:           false,
}
