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
	// Set "Enable" false if you want to run the agent in the non-Kubernetes environment.
	// Otherwise, the agent will panic if it can't connect to the API-server.
	Enable bool `mapstructure:"enable"`
}

var DefaultConfig Config = Config{
	KubeAuthType:      "serviceAccount",
	KubeConfigDir:     "/root/.kube/config",
	GraceDeletePeriod: 60,
	Enable:            true,
}
