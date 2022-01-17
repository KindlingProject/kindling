package analyzer

import "github.com/dxsup/kindling-collector/model"

type Type string

func (t Type) String() string {
	return string(t)
}

type Analyzer interface {
	// Start initializes the analyzer
	Start() error
	// ConsumeEvent gets the event from the previous component
	ConsumeEvent(event *model.KindlingEvent) error
	// Shutdown cleans all the resources used by the analyzer
	Shutdown() error
	// Type returns the type of the analyzer
	Type() Type
}
