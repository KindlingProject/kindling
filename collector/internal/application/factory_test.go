package application

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/processor/k8sprocessor"
)

func TestConstructConfig(t *testing.T) {
	factory := NewComponentsFactory()
	factory.RegisterProcessor(k8sprocessor.K8sMetadata, k8sprocessor.NewKubernetesProcessor, &k8sprocessor.DefaultConfig)
	factory.RegisterAnalyzer(network.Network.String(), network.NewNetworkAnalyzer, network.NewDefaultConfig())

	// Construct the config from the yaml file
	v := viper.New()
	v.SetConfigFile("./testdata/kindling-collector-config.yaml")
	err := v.ReadInConfig()
	assert.NoError(t, err)

	err = factory.ConstructConfig(v)
	assert.NoError(t, err)

	//// Assert the config is as expected
	k8sProcessorFactory := factory.Processors[k8sprocessor.K8sMetadata]
	k8sCfg := k8sProcessorFactory.Config.(*k8sprocessor.Config)
	// The expected config is exactly the opposite of the default config
	expectedCfg := &k8sprocessor.Config{
		Enable:                false,
		KubeAuthType:          "kubeConfig",
		KubeConfigDir:         "/opt/.kube/config",
		GraceDeletePeriod:     30,
		EnableFetchReplicaSet: true,
	}
	assert.Equal(t, expectedCfg, k8sCfg)

	networkAnalyzerFactory := factory.Analyzers[network.Network.String()]
	networkConfig := networkAnalyzerFactory.Config
	expectedNetworkConfig := &network.Config{
		EnableTimeoutCheck:    true,
		ConnectTimeout:        100,
		FdReuseTimeout:        15,
		NoResponseThreshold:   120,
		ResponseSlowThreshold: 500,
		EnableConntrack:       true,
		ConntrackMaxStateSize: 131072,
		ConntrackRateLimit:    500,
		ProcRoot:              "/proc",
		// Case: This slice is from the default config. The config file doesn't have this field.
		ProtocolParser: []string{"http", "mysql", "dns", "redis", "kafka", "dubbo"},
		// Case: This slice is overridden by the config file. The default config is different.
		ProtocolConfigs: []network.ProtocolConfig{
			{
				Key:           "http",
				Ports:         []uint32{80, 8080},
				PayloadLength: 100,
				Threshold:     200,
			},
		},
		UrlClusteringMethod: "blank",
		
		IgnoreDnsRcode3Error: false,
	}
	assert.Equal(t, expectedNetworkConfig, networkConfig)
}
