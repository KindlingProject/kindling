package k8sprocessor

import (
	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	KubeAuthType  kubernetes.AuthType `mapstructure:"kube_auth_type"`
	KubeConfigDir string              `mapstructure:"kube_config_dir"`
	// GraceDeletePeriod controls the delay interval after receiving delete event.
	// The unit is seconds, and the default value is 60 seconds.
	// Should not be lower than 30 seconds.
	GraceDeletePeriod int `mapstructure:"grace_delete_period"`
	// EnableFetchReplicaSet controls whether to fetch ReplicaSet information.
	// The default value is false. It should be enabled if the ReplicaSet
	// is used to control pods in the third-party CRD except for Deployment.
	EnableFetchReplicaSet bool `mapstructure:"enable_fetch_replicaset"`
	// Set "Enable" false if you want to run the agent in the non-Kubernetes environment.
	// Otherwise, the agent will panic if it can't connect to the API-server.
	Enable bool `mapstructure:"enable"`

	LabelSelector *v1.LabelSelector `mapstructure:"label_selector"`
}

var DefaultConfig Config = Config{
	KubeAuthType:      "serviceAccount",
	KubeConfigDir:     "/root/.kube/config",
	GraceDeletePeriod: 60,
	Enable:            true,
}
