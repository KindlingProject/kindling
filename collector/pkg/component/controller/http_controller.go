package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"go.uber.org/zap"
)

type HttpAPI struct {
	*http.ServeMux

	controllerMap map[string]Controller
	tools         *component.TelemetryTools
}

type HttpControllerConfig struct {
	Enable bool
	Port   string
}

func NewHttpAPI(tools *component.TelemetryTools) *HttpAPI {
	return &HttpAPI{
		controllerMap: make(map[string]Controller),
		ServeMux:      http.NewServeMux(),
		tools:         tools,
	}
}

func (hc *HttpAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, _ := hc.Handler(r)
	h.ServeHTTP(w, r)
}

func (hc *HttpAPI) RegistModule(module string, subModules ...ExportSubModule) {
	if c, ok := hc.controllerMap[module]; ok {
		c.RegistSubModules(subModules...)
	}
}

func (hc *HttpAPI) RegistController(c Controller) {
	hc.controllerMap[c.GetModuleKey()] = c
	hc.HandleFunc(c.GetModuleKey(), func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			hc.tools.Logger.Error("read request failed", zap.Error(err))
			return
		}
		req, err := parseRequest(b)
		if err != nil {
			w.WriteHeader(500)
			hc.tools.Logger.Error("parse request failed", zap.Error(err))
			return
		}
		resp := c.HandRequest(req)
		if err != nil {
			w.WriteHeader(500)
			hc.tools.Logger.Error("handle request failed", zap.Error(err))
			return
		}

		msg, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(500)
			hc.tools.Logger.Error("write response failed", zap.Error(err))
			return
		}
		w.WriteHeader(200)
		w.Write(msg)
	})
}

func parseRequest(body []byte) (*ControlRequest, error) {
	var req *ControlRequest
	err := json.Unmarshal(body, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}
