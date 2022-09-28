package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
)

const (
	NoError        = 1
	StartWithError = 2
	StopWithError  = 3
	NoOperation    = 4
)

const ProfileModule = "profile"

type Profile struct {
	Module
}

type ProfileOption struct {
	// seconds
	Duration time.Duration

	// bytes
	FileWatch *FileWatchOption
}

type FileWatchOption struct {
	File     string
	DataSize int
}

type ExportSubModule func() (name string, start func() error, stop func() error)

func NewProfileController(tools *component.TelemetryTools) *Profile {
	profile := NewModule(ProfileModule, tools)
	return &Profile{
		Module: profile,
	}
}

func (p *Profile) RegistSubModules(subModules ...ExportSubModule) {
	for _, subModule := range subModules {
		p.Module.RegisterSubModule(subModule())
	}
}

func (o *ProfileOption) UnmarshalJSON(b []byte) error {
	var v struct {
		FileWatch *FileWatchOption
		Duration  interface{}
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	o.FileWatch = v.FileWatch
	switch value := v.Duration.(type) {
	case float64:
		o.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		o.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func (p *Profile) GetModuleKey() string {
	return p.Name()
}

func (p *Profile) HandRequest(req *ControlRequest) *ControlResponse {
	switch req.Operation {
	case "start":
		if err := p.Start(p.GetOptions(req.Options)...); err != nil {
			return &ControlResponse{
				Code: StartWithError,
				Msg:  err.Error(),
			}
		}
		return &ControlResponse{
			Code: NoError,
			Msg:  "start success",
		}
	case "stop":
		if err := p.Stop("module stop by manual"); err != nil {
			return &ControlResponse{
				Code: StopWithError,
				Msg:  err.Error(),
			}
		}
		return &ControlResponse{
			Code: NoError,
			Msg:  "stop success",
		}
	default:
		return &ControlResponse{
			Code: NoOperation,
			Msg:  fmt.Sprintf("unexpected operation:%s", req.Operation),
		}
	}
}

func (p *Profile) GetOptions(raw_opts *json.RawMessage) []Option {
	var pOption ProfileOption
	var opts []Option
	json.Unmarshal(*raw_opts, &pOption)
	if pOption.Duration > 0 {
		opts = append(opts, WithStopInterval(pOption.Duration))
	}
	if pOption.FileWatch != nil {
		opts = append(opts, WithStopSignal(fileSizeWatch(pOption.FileWatch)))
	}
	return opts
}

func fileSizeWatch(filewatch *FileWatchOption) <-chan struct{} {
	// TDDO
	panic("file size watch is not implemented yet")
}
