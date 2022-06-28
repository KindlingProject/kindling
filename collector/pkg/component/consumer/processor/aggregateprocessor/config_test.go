package aggregateprocessor

import (
	"testing"

	"github.com/spf13/viper"
)

func TestCreateConfig(t *testing.T) {
	v := viper.New()
	v.SetConfigFile("testdata/config.yaml")
	err := v.ReadInConfig()
	if err != nil {
		t.Fatalf("Error happened during reading config file: %v", err)
	}

	var config Config
	err = v.Unmarshal(&config)
	if err != nil {
		t.Fatalf("Error happened during unmarshaling config: %v", err)
	}

	t.Log(config)
}
