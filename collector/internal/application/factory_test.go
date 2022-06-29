package application

import (
	"reflect"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/processor/k8sprocessor"
	"github.com/spf13/viper"
)

func TestConstructConfig(t *testing.T) {
	factory := NewComponentsFactory()
	factory.RegisterProcessor(k8sprocessor.K8sMetadata, k8sprocessor.NewKubernetesProcessor, &k8sprocessor.DefaultConfig)
	v := viper.New()
	v.SetConfigFile("testdata/kindling-collector-config.yaml")
	v.ReadInConfig()
	factory.ConstructConfig(v)
	k8sprocessorFactory := factory.Processors[k8sprocessor.K8sMetadata]
	cfg := k8sprocessorFactory.Config.(*k8sprocessor.Config)
	expectedCfg := &k8sprocessor.Config{
		KubeAuthType:      "kubeConfig",
		KubeConfigDir:     "~/.kube/config",
		GraceDeletePeriod: 60,
	}
	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Errorf("Expected %v, but get %v", expectedCfg, cfg)
	}
}
