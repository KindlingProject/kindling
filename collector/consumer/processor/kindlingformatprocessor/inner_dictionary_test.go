package kindlingformatprocessor

import (
	"encoding/json"
	"fmt"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"reflect"
	"testing"
)

var values = map[string]int64{
	constvalues.RequestIo:           1,
	constvalues.ResponseIo:          2,
	constvalues.RequestTotalTime:    3,
	constvalues.RequestSentTime:     4,
	constvalues.WaitingTtfbTime:     5,
	constvalues.ContentDownloadTime: 6,
}

func Test_gauges_Process(t *testing.T) {
	type args struct {
		gauges   *gauges
		cfg      *Config
		relabels []Relabel
	}
	tests := []struct {
		name string
		args args
		want *model.GaugeGroup
	}{
		{
			name: "Trace",
			args: args{
				gauges:   newGauges(newInnerGauges(true)),
				cfg:      &Config{NeedTraceAsMetric: true, NeedPodDetail: true},
				relabels: []Relabel{TraceName, TopologyInstanceInfo, TopologyK8sInfo, ServiceProtocolInfo, TraceStatusInfo},
			},
			want: getTrace(),
		},
		{
			name: "ServiceMetric",
			args: args{
				gauges:   newGauges(newInnerGauges(true)),
				cfg:      &Config{NeedTraceAsMetric: true, NeedPodDetail: true},
				relabels: []Relabel{MetricName, ServiceInstanceInfo, ServiceK8sInfo, ServiceProtocolInfo},
			},
			want: getServiceMetric(),
		},
		{
			name: "TopologyMetric",
			args: args{
				gauges:   newGauges(newInnerGauges(false)),
				cfg:      &Config{NeedTraceAsMetric: true, NeedPodDetail: true},
				relabels: []Relabel{MetricName, TopologyInstanceInfo, TopologyK8sInfo, SrcDockerInfo, TopologyProtocolInfo},
			},
			want: getTopologyMetric(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.args.gauges
			if got := g.Process(tt.args.cfg, tt.args.relabels...); !reflect.DeepEqual(got, tt.want) {
				gotJson, err := json.Marshal(got.Values)
				if err != nil {
					fmt.Println("err = ", err)
					return
				}
				wantJson, err := json.Marshal(tt.want.Values)
				if err != nil {
					fmt.Println("err = ", err)
					return
				}
				t.Errorf("ProcessValues() = \n%s\nWant:\n%s\n", gotJson, wantJson)
				gotJson, err = json.Marshal(got.Labels.ToStringMap())
				if err != nil {
					fmt.Println("err = ", err)
					return
				}
				wantJson, err = json.Marshal(tt.want.Labels.ToStringMap())
				if err != nil {
					fmt.Println("err = ", err)
					return
				}
				t.Errorf("Process() = \n%s\nWant:\n%s\n", gotJson, wantJson)
			}
		})
	}
}

func newInnerGauges(isServer bool) *model.GaugeGroup {
	gaugesGroup := model.GaugeGroup{
		Name: "testGauge",
		Values: []*model.Gauge{
			{constvalues.RequestIo, values[constvalues.RequestIo]},
			{constvalues.ResponseIo, values[constvalues.ResponseIo]},
			{constvalues.RequestTotalTime, values[constvalues.RequestTotalTime]},
			{constvalues.RequestSentTime, values[constvalues.RequestSentTime]},
			{constvalues.WaitingTtfbTime, values[constvalues.WaitingTtfbTime]},
			{constvalues.ContentDownloadTime, values[constvalues.ContentDownloadTime]},
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 1156665465,
	}

	gaugesGroup.Labels.AddStringValue(constlabels.Pid, "test-Pid")
	gaugesGroup.Labels.AddStringValue(constlabels.Protocol, http)
	gaugesGroup.Labels.AddStringValue(constlabels.HttpUrl, "httpUrl")
	// For now, only http gauges will have the ContentKey label.
	gaugesGroup.Labels.AddStringValue(constlabels.ContentKey, "httpUrl")
	gaugesGroup.Labels.AddIntValue(constlabels.HttpStatusCode, 200)
	gaugesGroup.Labels.AddStringValue(constlabels.IsError, "test-IsError")
	gaugesGroup.Labels.AddStringValue(constlabels.ErrorType, "test-ErrorType")
	gaugesGroup.Labels.AddStringValue(constlabels.IsSlow, "test-IsSlow")
	gaugesGroup.Labels.AddBoolValue(constlabels.IsServer, isServer)
	gaugesGroup.Labels.AddStringValue(constlabels.ContainerId, "test-ContainerId")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcPort, "test-SrcPort")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService")
	gaugesGroup.Labels.AddStringValue(constlabels.DstIp, "test-DstIp")
	gaugesGroup.Labels.AddIntValue(constlabels.DstPort, 8080)
	gaugesGroup.Labels.AddStringValue(constlabels.DnatIp, "test-DnatIp")
	gaugesGroup.Labels.AddIntValue(constlabels.DnatPort, 8081)
	gaugesGroup.Labels.AddStringValue(constlabels.DstContainerId, "test-DstContainerId")
	gaugesGroup.Labels.AddStringValue(constlabels.DstContainer, "test-DstContainer")

	return &gaugesGroup
}

func getTrace() *model.GaugeGroup {
	gaugesGroup := model.GaugeGroup{
		Name: "testGauge",
		Values: []*model.Gauge{
			{"kindling_trace_request_duration_nanoseconds", values[constvalues.RequestTotalTime]},
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 1156665465,
	}

	gaugesGroup.Labels.AddStringValue(constlabels.Protocol, http)
	gaugesGroup.Labels.AddBoolValue(constlabels.IsServer, true)
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService")

	gaugesGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	gaugesGroup.Labels.AddIntValue(constlabels.DstPort, 8081)

	gaugesGroup.Labels.AddStringValue(constlabels.Protocol, http)
	gaugesGroup.Labels.AddStringValue(constlabels.RequestContent, "httpUrl")
	gaugesGroup.Labels.AddStringValue(constlabels.ResponseContent, "200")

	gaugesGroup.Labels.AddStringValue(constlabels.RequestDurationStatus, GreenStatus)
	gaugesGroup.Labels.AddStringValue(constlabels.RequestReqxferStatus, GreenStatus)
	gaugesGroup.Labels.AddStringValue(constlabels.RequestProcessingStatus, GreenStatus)
	gaugesGroup.Labels.AddStringValue(constlabels.ResponseRspxferStatus, GreenStatus)

	return &gaugesGroup
}

func getServiceMetric() *model.GaugeGroup {
	gaugesGroup := model.GaugeGroup{
		Name: "testGauge",
		Values: []*model.Gauge{
			{"kindling_entity_request_receive_bytes_total", values[constvalues.RequestIo]},
			{"kindling_entity_request_send_bytes_total", values[constvalues.ResponseIo]},
			{"kindling_entity_request_duration_nanoseconds", values[constvalues.RequestTotalTime]},
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 1156665465,
	}

	gaugesGroup.Labels.AddStringValue(constlabels.Protocol, http)
	gaugesGroup.Labels.AddStringValue(constlabels.Node, "test-DstNode")
	gaugesGroup.Labels.AddStringValue(constlabels.Namespace, "test-DstNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.Pod, "test-DstPod")
	gaugesGroup.Labels.AddStringValue(constlabels.WorkloadName, "test-DstWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.WorkloadKind, "test-DstWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.Service, "test-DstService")
	gaugesGroup.Labels.AddStringValue(constlabels.Ip, "test-DstIp")
	gaugesGroup.Labels.AddIntValue(constlabels.Port, 8080)
	gaugesGroup.Labels.AddStringValue(constlabels.Container, "test-DstContainer")
	gaugesGroup.Labels.AddStringValue(constlabels.ContainerId, "test-DstContainerId")

	// EntityDataDoesNotExistDNatIp
	gaugesGroup.Labels.AddStringValue(constlabels.Ip, "test-DstIp")
	gaugesGroup.Labels.AddIntValue(constlabels.Port, 8080)

	gaugesGroup.Labels.AddStringValue(constlabels.Protocol, http)
	gaugesGroup.Labels.AddStringValue(constlabels.RequestContent, "httpUrl")
	gaugesGroup.Labels.AddStringValue(constlabels.ResponseContent, "200")
	return &gaugesGroup
}

func getTopologyMetric() *model.GaugeGroup {
	gaugesGroup := model.GaugeGroup{
		Name: "testGauge",
		Values: []*model.Gauge{
			{"kindling_topology_request_request_bytes_total", values[constvalues.RequestIo]},
			{"kindling_topology_request_response_bytes_total", values[constvalues.ResponseIo]},
			{"kindling_topology_request_duration_nanoseconds", values[constvalues.RequestTotalTime]},
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 1156665465,
	}

	gaugesGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService")

	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId")

	gaugesGroup.Labels.AddStringValue(constlabels.Protocol, http)
	gaugesGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use DNat Ip and DNat Port
	gaugesGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	gaugesGroup.Labels.AddIntValue(constlabels.DstPort, 8081)

	return &gaugesGroup
}
