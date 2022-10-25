package controller

import (
	"encoding/json"
	"net/http"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/spf13/viper"
)

type ControllerAPI interface {
	RegistController(c Controller)
	RegistModule(module string, subModules ...ExportSubModule)
}

type Controller interface {
	GetModuleKey() string
	RegistSubModules(subModules ...ExportSubModule)
	HandRequest(*ControlRequest) *ControlResponse
	GetOptions(*json.RawMessage) []Option
}

const ControllerComponent string = "controller"

type ControlRequest struct {
	// start / stop
	Operation string

	// json Options
	Options *json.RawMessage
}

type ControlResponse struct {
	Code int
	Msg  string
}

type ControllerFactory struct {
	Controller ControllerAPI
}

type ControllerConfig struct {
	Http    *HttpControllerConfig
	Modules []string
}

func (cf *ControllerFactory) ConstructConfig(viper *viper.Viper, tools *component.TelemetryTools) error {
	var controllerConfig ControllerConfig
	key := ControllerComponent
	err := viper.UnmarshalKey(key, &controllerConfig)
	if err != nil {
		tools.Logger.Errorf("Error happened when reading controller config, will disable all controller: %v", err)
	}
	if controllerConfig.Http != nil {
		httpAPI := NewHttpAPI(tools)
		for _, module := range controllerConfig.Modules {
			switch module {
			case ProfileModule:
				profileController := NewProfileController(tools)
				httpAPI.RegistController(profileController)
			}
		}
		go http.ListenAndServe(controllerConfig.Http.Port, httpAPI)
		cf.Controller = httpAPI
	}
	return nil
}

func (cf *ControllerFactory) RegistModule(module string, subModules ...ExportSubModule) {
	cf.Controller.RegistModule(module, subModules...)
}
