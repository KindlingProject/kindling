package network

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/cpuanalyzer"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
)

// signal is used to trigger to send CPU on/off events
func checkSendSignalToCpuAnalyzer(data *model.DataGroup) {
	if data.Labels.GetBoolValue(constlabels.IsError) ||
		data.Labels.GetBoolValue(constlabels.IsSlow) {
		duration, ok := data.GetMetric(constvalues.RequestTotalTime)
		if !ok {
			return
		}
		event := cpuanalyzer.SendTriggerEvent{
			Pid:       uint32(data.Labels.GetIntValue("pid")),
			StartTime: data.Timestamp,
			SpendTime: uint64(duration.GetInt().Value),
		}
		cpuanalyzer.ReceiveSendSignal(event)
	}
}