package cpuanalyzer

import "github.com/Kindling-project/kindling/collector/model"

type CpuAnalyzer struct {
	cpuPidEvents map[int]string
}

type timeEvents struct{
	startTime int64
	cpuEvent  CpuEvent
}

type CpuEvent struct {

}


func (ca *CpuAnalyzer) Start() error {
	es, _ := elasticsearch.NewDefaultClient()
	return nil
}


func (ca *CpuAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {

}
