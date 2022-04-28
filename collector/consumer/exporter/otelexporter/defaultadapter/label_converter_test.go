package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
	"testing"
)

var baseAdapter = createNetAdapterManager([]attribute.KeyValue{
	{"const-labels1", attribute.StringValue("const-values1")},
})

func TestAdapter_transform(t *testing.T) {
	type fields struct {
		labelsMap       map[extraLabelsKey]realAttributes
		updateKeys      []updateKey
		valueLabelsFunc valueToLabels
		adjustFunctions []adjustFunctions
	}
	type args struct {
		group *model.GaugeGroup
	}
	tests := []struct {
		name           string
		labelConverter *LabelConverter
		args           args
		want           *model.AttributeMap
		wantErr        bool
	}{
		{
			name:           "kindling_agg_net_topology",
			labelConverter: baseAdapter.aggTopologyAdapter[0],
			args: args{group: model.NewGaugeGroup(
				constnames.AggregatedNetRequestGaugeGroup,
				model.NewAttributeMapWithValues(
					map[string]model.AttributeValue{
						// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
						constlabels.SrcIp:   model.NewStringValue("src-ip"),
						constlabels.SrcPort: model.NewIntValue(33333),
						constlabels.DstIp:   model.NewStringValue("dst-ip"),
						constlabels.DstPort: model.NewIntValue(8080),

						// protocolInfo
						constlabels.Protocol:       model.NewStringValue("http"),
						constlabels.HttpUrl:        model.NewStringValue("/test"),
						constlabels.HttpStatusCode: model.NewIntValue(200),

						// k8sInfo
						constlabels.DstPod:          model.NewStringValue("dst-pod"),
						constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
						constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
						constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
						constlabels.SrcPod:          model.NewStringValue("src-pod"),
						constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
						constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
						constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
						constlabels.SrcService:      model.NewStringValue("src-service"),
						constlabels.DstService:      model.NewStringValue("dst-service"),
						constlabels.SrcNode:         model.NewStringValue("src-node"),
						constlabels.DstNode:         model.NewStringValue("dst-node"),

						// isSlow
						constlabels.IsSlow: model.NewBoolValue(false),
					}),
				123,
				[]*model.Gauge{
					{constvalues.RequestTotalTime, 123},
					{constvalues.RequestIo, 456},
				}...),
			},
			want: model.NewAttributeMapWithValues(map[string]model.AttributeValue{
				// instanceInfo is moved from agg
				// protocolInfo
				constlabels.Protocol:   model.NewStringValue("http"),
				constlabels.StatusCode: model.NewStringValue("200"),

				// k8sInfo
				constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
				constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
				constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
				constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
				constlabels.SrcService:      model.NewStringValue("src-service"),
				constlabels.DstService:      model.NewStringValue("dst-service"),

				// remove but exist
				constlabels.DstNode: model.NewStringValue(""),
				constlabels.DstPod:  model.NewStringValue(""),

				"const-labels1": model.NewStringValue("const-values1"),
			}),
		},
		{
			name:           "kindling_detail_net_topology",
			labelConverter: baseAdapter.detailTopologyAdapter[0],
			args: args{group: model.NewGaugeGroup(
				constnames.AggregatedNetRequestGaugeGroup,
				model.NewAttributeMapWithValues(
					map[string]model.AttributeValue{
						// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
						constlabels.SrcIp:   model.NewStringValue("src-ip"),
						constlabels.SrcPort: model.NewIntValue(33333),
						constlabels.DstIp:   model.NewStringValue("dst-ip"),
						constlabels.DstPort: model.NewIntValue(8080),

						// protocolInfo
						constlabels.Protocol:       model.NewStringValue("http"),
						constlabels.HttpUrl:        model.NewStringValue("/test"),
						constlabels.HttpStatusCode: model.NewIntValue(200),

						// k8sInfo
						constlabels.DstPod:          model.NewStringValue("dst-pod"),
						constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
						constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
						constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
						constlabels.SrcPod:          model.NewStringValue("src-pod"),
						constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
						constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
						constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
						constlabels.SrcService:      model.NewStringValue("src-service"),
						constlabels.DstService:      model.NewStringValue("dst-service"),
						constlabels.SrcNode:         model.NewStringValue("src-node"),
						constlabels.DstNode:         model.NewStringValue("dst-node"),
						constlabels.DnatIp:          model.NewStringValue("dnat-ip"),
						constlabels.DnatPort:        model.NewIntValue(80),

						// isSlow
						constlabels.IsSlow: model.NewBoolValue(false),
					}),
				123,
				[]*model.Gauge{
					{constvalues.RequestTotalTime, 123},
					{constvalues.RequestIo, 456},
				}...),
			},
			want: model.NewAttributeMapWithValues(map[string]model.AttributeValue{
				// instanceInfo is moved from agg
				constlabels.SrcIp:   model.NewStringValue("src-ip"),
				constlabels.DstIp:   model.NewStringValue("dnat-ip"),
				constlabels.DstPort: model.NewIntValue(80),
				// protocolInfo
				constlabels.Protocol:   model.NewStringValue("http"),
				constlabels.StatusCode: model.NewStringValue("200"),

				// k8sInfo
				constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
				constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
				constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
				constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
				constlabels.SrcService:      model.NewStringValue("src-service"),
				constlabels.DstService:      model.NewStringValue("dst-service"),
				constlabels.SrcNode:         model.NewStringValue("src-node"),
				constlabels.DstNode:         model.NewStringValue("dst-node"),
				constlabels.SrcPod:          model.NewStringValue("src-pod"),
				constlabels.DstPod:          model.NewStringValue("dst-pod"),

				"const-labels1": model.NewStringValue("const-values1"),
			}),
		},
		{
			name:           "kindling_detail_net_entity",
			labelConverter: baseAdapter.detailEntityAdapter[0],
			args: args{group: model.NewGaugeGroup(
				constnames.AggregatedNetRequestGaugeGroup,
				model.NewAttributeMapWithValues(
					map[string]model.AttributeValue{
						// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
						constlabels.SrcIp:   model.NewStringValue("src-ip"),
						constlabels.SrcPort: model.NewIntValue(33333),
						constlabels.DstIp:   model.NewStringValue("dst-ip"),
						constlabels.DstPort: model.NewIntValue(8080),

						// protocolInfo
						constlabels.Protocol:       model.NewStringValue("http"),
						constlabels.ContentKey:     model.NewStringValue("/test"),
						constlabels.HttpStatusCode: model.NewIntValue(200),

						// k8sInfo
						constlabels.DstPod:          model.NewStringValue("dst-pod"),
						constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
						constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
						constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
						constlabels.SrcPod:          model.NewStringValue("src-pod"),
						constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
						constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
						constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
						constlabels.SrcService:      model.NewStringValue("src-service"),
						constlabels.DstService:      model.NewStringValue("dst-service"),
						constlabels.SrcNode:         model.NewStringValue("src-node"),
						constlabels.DstNode:         model.NewStringValue("dst-node"),
						constlabels.DnatIp:          model.NewStringValue("dnat-ip"),
						constlabels.DnatPort:        model.NewIntValue(80),

						// isSlow
						constlabels.IsSlow: model.NewBoolValue(false),
					}),
				123,
				[]*model.Gauge{
					{constvalues.RequestTotalTime, 123},
					{constvalues.RequestIo, 456},
				}...),
			},
			want: model.NewAttributeMapWithValues(map[string]model.AttributeValue{
				// instanceInfo is moved from agg
				constlabels.Ip:   model.NewStringValue("dst-ip"),
				constlabels.Port: model.NewIntValue(8080),
				// protocolInfo
				constlabels.Protocol:        model.NewStringValue("http"),
				constlabels.RequestContent:  model.NewStringValue("/test"),
				constlabels.ResponseContent: model.NewStringValue("200"),

				// k8sInfo
				constlabels.WorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.Namespace:    model.NewStringValue("dst-Namespace"),
				constlabels.WorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.Service:      model.NewStringValue("dst-service"),
				constlabels.Node:         model.NewStringValue("dst-node"),
				constlabels.Pod:          model.NewStringValue("dst-pod"),

				"const-labels1": model.NewStringValue("const-values1"),
			}),
		},
		{
			name:           "kindling_agg_net_entity",
			labelConverter: baseAdapter.aggEntityAdapter[0],
			args: args{group: model.NewGaugeGroup(
				constnames.AggregatedNetRequestGaugeGroup,
				model.NewAttributeMapWithValues(
					map[string]model.AttributeValue{
						// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
						constlabels.SrcIp:   model.NewStringValue("src-ip"),
						constlabels.SrcPort: model.NewIntValue(33333),
						constlabels.DstIp:   model.NewStringValue("dst-ip"),
						constlabels.DstPort: model.NewIntValue(8080),

						// protocolInfo
						constlabels.Protocol:       model.NewStringValue("http"),
						constlabels.ContentKey:     model.NewStringValue("/test"),
						constlabels.HttpStatusCode: model.NewIntValue(200),

						// k8sInfo
						constlabels.DstPod:          model.NewStringValue("dst-pod"),
						constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
						constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
						constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
						constlabels.SrcPod:          model.NewStringValue("src-pod"),
						constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
						constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
						constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
						constlabels.SrcService:      model.NewStringValue("src-service"),
						constlabels.DstService:      model.NewStringValue("dst-service"),
						constlabels.SrcNode:         model.NewStringValue("src-node"),
						constlabels.DstNode:         model.NewStringValue("dst-node"),
						constlabels.DnatIp:          model.NewStringValue("dnat-ip"),
						constlabels.DnatPort:        model.NewIntValue(80),

						// isSlow
						constlabels.IsSlow: model.NewBoolValue(false),
					}),
				123,
				[]*model.Gauge{
					{constvalues.RequestTotalTime, 123},
					{constvalues.RequestIo, 456},
				}...),
			},
			want: model.NewAttributeMapWithValues(map[string]model.AttributeValue{
				// protocolInfo
				constlabels.Protocol:        model.NewStringValue("http"),
				constlabels.RequestContent:  model.NewStringValue("/test"),
				constlabels.ResponseContent: model.NewStringValue("200"),

				// k8sInfo
				constlabels.WorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.Namespace:    model.NewStringValue("dst-Namespace"),
				constlabels.WorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.Service:      model.NewStringValue("dst-service"),

				"const-labels1": model.NewStringValue("const-values1"),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.labelConverter
			got, free, err := m.transform(tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for key, value := range got.GetValues() {
				if valueW, ok := tt.want.GetValues()[key]; ok {
					if value.ToString() != valueW.ToString() {
						t.Errorf("transform() get = %v, want %v", value.ToString(), valueW.ToString())
					}
				} else {
					if value.ToString() != "" {
						t.Errorf("transform() get = '%v =  %v', don't want this label", key, value.ToString())
					}
				}
			}
			for key, value := range tt.want.GetValues() {
				if _, ok := got.GetValues()[key]; !ok {
					t.Errorf("transform() expected key '%v' ,value '%v',but not exist", key, value.ToString())
				}
			}
			free(got)
		})
	}
}

func TestAdapter_adapt(t *testing.T) {
	type fields struct {
		labelsMap       map[extraLabelsKey]realAttributes
		updateKeys      []updateKey
		valueLabelsFunc valueToLabels
		adjustFunctions []adjustFunctions
	}
	type args struct {
		group *model.GaugeGroup
	}
	tests := []struct {
		name    string
		adapter *LabelConverter
		args    args
		want    []attribute.KeyValue
		wantErr bool
	}{
		{
			name:    "kindling_agg_net_topology",
			adapter: baseAdapter.aggTopologyAdapter[0],
			args: args{group: model.NewGaugeGroup(
				constnames.AggregatedNetRequestGaugeGroup,
				model.NewAttributeMapWithValues(
					map[string]model.AttributeValue{
						// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
						constlabels.SrcIp:   model.NewStringValue("src-ip"),
						constlabels.SrcPort: model.NewIntValue(33333),
						constlabels.DstIp:   model.NewStringValue("dst-ip"),
						constlabels.DstPort: model.NewIntValue(8080),

						// protocolInfo
						constlabels.Protocol:       model.NewStringValue("http"),
						constlabels.HttpUrl:        model.NewStringValue("/test"),
						constlabels.HttpStatusCode: model.NewIntValue(200),

						// k8sInfo
						constlabels.DstPod:          model.NewStringValue("dst-pod"),
						constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
						constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
						constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
						constlabels.SrcPod:          model.NewStringValue("src-pod"),
						constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
						constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
						constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
						constlabels.SrcService:      model.NewStringValue("src-service"),
						constlabels.DstService:      model.NewStringValue("dst-service"),
						constlabels.SrcNode:         model.NewStringValue("src-node"),
						constlabels.DstNode:         model.NewStringValue("dst-node"),

						// isSlow
						constlabels.IsSlow: model.NewBoolValue(false),
					}),
				123,
				[]*model.Gauge{
					{constvalues.RequestTotalTime, 123},
					{constvalues.RequestIo, 456},
				}...),
			},
			want: []attribute.KeyValue{
				// protocolInfo
				{constlabels.Protocol, attribute.StringValue("http")},
				{constlabels.StatusCode, attribute.StringValue("200")},

				// k8sInfo
				{constlabels.DstWorkloadName, attribute.StringValue("dst-workloadName")},
				{constlabels.DstNamespace, attribute.StringValue("dst-Namespace")},
				{constlabels.DstWorkloadKind, attribute.StringValue("dst-workloadKind")},
				{constlabels.SrcWorkloadName, attribute.StringValue("src-workloadName")},
				{constlabels.SrcNamespace, attribute.StringValue("src-Namespace")},
				{constlabels.SrcWorkloadKind, attribute.StringValue("src-workloadKind")},
				{constlabels.SrcService, attribute.StringValue("src-service")},
				{constlabels.DstService, attribute.StringValue("dst-service")},

				// remove but exist
				{constlabels.DstNode, attribute.StringValue("")},
				{constlabels.DstPod, attribute.StringValue("")},

				{"const-labels1", attribute.StringValue("const-values1")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.adapter
			got, free, err := m.convert(tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, keyValueG := range got {
				for _, keyValueW := range tt.want {
					if keyValueG.Key == keyValueW.Key && keyValueG.Value != keyValueW.Value {
						t.Errorf("transform() get = %v, want %v", keyValueG.Value.AsString(), keyValueW.Value.AsString())
					}
				}
			}
			free(got)
		})
	}
}

func TestAdapter_transform_async(t *testing.T) {
	type fields struct {
		labelsMap       map[extraLabelsKey]realAttributes
		updateKeys      []updateKey
		valueLabelsFunc valueToLabels
		adjustFunctions []adjustFunctions
	}
	type args struct {
		group *model.GaugeGroup
	}
	tests := []struct {
		name           string
		labelConverter *LabelConverter
		args           args
		want           *model.AttributeMap
		wantErr        bool
	}{
		{
			name:           "kindling_agg_net_topology",
			labelConverter: baseAdapter.aggTopologyAdapter[0],
			args: args{group: model.NewGaugeGroup(
				constnames.AggregatedNetRequestGaugeGroup,
				model.NewAttributeMapWithValues(
					map[string]model.AttributeValue{
						// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
						constlabels.SrcIp:   model.NewStringValue("src-ip"),
						constlabels.SrcPort: model.NewIntValue(33333),
						constlabels.DstIp:   model.NewStringValue("dst-ip"),
						constlabels.DstPort: model.NewIntValue(8080),

						// protocolInfo
						constlabels.Protocol:       model.NewStringValue("http"),
						constlabels.HttpUrl:        model.NewStringValue("/test"),
						constlabels.HttpStatusCode: model.NewIntValue(200),

						// k8sInfo
						constlabels.DstPod:          model.NewStringValue("dst-pod"),
						constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
						constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
						constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
						constlabels.SrcPod:          model.NewStringValue("src-pod"),
						constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
						constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
						constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
						constlabels.SrcService:      model.NewStringValue("src-service"),
						constlabels.DstService:      model.NewStringValue("dst-service"),
						constlabels.SrcNode:         model.NewStringValue("src-node"),
						constlabels.DstNode:         model.NewStringValue("dst-node"),

						// isSlow
						constlabels.IsSlow: model.NewBoolValue(false),
					}),
				123,
				[]*model.Gauge{
					{constvalues.RequestTotalTime, 123},
					{constvalues.RequestIo, 456},
				}...),
			},
			want: model.NewAttributeMapWithValues(map[string]model.AttributeValue{
				// instanceInfo is moved from agg
				// protocolInfo
				constlabels.Protocol:   model.NewStringValue("http"),
				constlabels.StatusCode: model.NewStringValue("200"),

				// k8sInfo
				constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
				constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
				constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
				constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
				constlabels.SrcService:      model.NewStringValue("src-service"),
				constlabels.DstService:      model.NewStringValue("dst-service"),

				// remove but exist
				constlabels.DstNode: model.NewStringValue(""),
				constlabels.DstPod:  model.NewStringValue(""),

				"const-labels1": model.NewStringValue("const-values1"),
			}),
		},
		{
			name:           "kindling_detail_net_topology",
			labelConverter: baseAdapter.detailTopologyAdapter[0],
			args: args{group: model.NewGaugeGroup(
				constnames.AggregatedNetRequestGaugeGroup,
				model.NewAttributeMapWithValues(
					map[string]model.AttributeValue{
						// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
						constlabels.SrcIp:   model.NewStringValue("src-ip"),
						constlabels.SrcPort: model.NewIntValue(33333),
						constlabels.DstIp:   model.NewStringValue("dst-ip"),
						constlabels.DstPort: model.NewIntValue(8080),

						// protocolInfo
						constlabels.Protocol:       model.NewStringValue("http"),
						constlabels.HttpUrl:        model.NewStringValue("/test"),
						constlabels.HttpStatusCode: model.NewIntValue(200),

						// k8sInfo
						constlabels.DstPod:          model.NewStringValue("dst-pod"),
						constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
						constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
						constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
						constlabels.SrcPod:          model.NewStringValue("src-pod"),
						constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
						constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
						constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
						constlabels.SrcService:      model.NewStringValue("src-service"),
						constlabels.DstService:      model.NewStringValue("dst-service"),
						constlabels.SrcNode:         model.NewStringValue("src-node"),
						constlabels.DstNode:         model.NewStringValue("dst-node"),
						constlabels.DnatIp:          model.NewStringValue("dnat-ip"),
						constlabels.DnatPort:        model.NewIntValue(80),

						// isSlow
						constlabels.IsSlow: model.NewBoolValue(false),
					}),
				123,
				[]*model.Gauge{
					{constvalues.RequestTotalTime, 123},
					{constvalues.RequestIo, 456},
				}...),
			},
			want: model.NewAttributeMapWithValues(map[string]model.AttributeValue{
				// instanceInfo is moved from agg
				constlabels.SrcIp:   model.NewStringValue("src-ip"),
				constlabels.DstIp:   model.NewStringValue("dnat-ip"),
				constlabels.DstPort: model.NewIntValue(80),
				// protocolInfo
				constlabels.Protocol:   model.NewStringValue("http"),
				constlabels.StatusCode: model.NewStringValue("200"),

				// k8sInfo
				constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
				constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
				constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
				constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
				constlabels.SrcService:      model.NewStringValue("src-service"),
				constlabels.DstService:      model.NewStringValue("dst-service"),
				constlabels.SrcNode:         model.NewStringValue("src-node"),
				constlabels.DstNode:         model.NewStringValue("dst-node"),
				constlabels.SrcPod:          model.NewStringValue("src-pod"),
				constlabels.DstPod:          model.NewStringValue("dst-pod"),

				"const-labels1": model.NewStringValue("const-values1"),
			}),
		},
		{
			name:           "kindling_detail_net_entity",
			labelConverter: baseAdapter.detailEntityAdapter[0],
			args: args{group: model.NewGaugeGroup(
				constnames.AggregatedNetRequestGaugeGroup,
				model.NewAttributeMapWithValues(
					map[string]model.AttributeValue{
						// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
						constlabels.SrcIp:   model.NewStringValue("src-ip"),
						constlabels.SrcPort: model.NewIntValue(33333),
						constlabels.DstIp:   model.NewStringValue("dst-ip"),
						constlabels.DstPort: model.NewIntValue(8080),

						// protocolInfo
						constlabels.Protocol:       model.NewStringValue("http"),
						constlabels.ContentKey:     model.NewStringValue("/test"),
						constlabels.HttpStatusCode: model.NewIntValue(200),

						// k8sInfo
						constlabels.DstPod:          model.NewStringValue("dst-pod"),
						constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
						constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
						constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
						constlabels.SrcPod:          model.NewStringValue("src-pod"),
						constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
						constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
						constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
						constlabels.SrcService:      model.NewStringValue("src-service"),
						constlabels.DstService:      model.NewStringValue("dst-service"),
						constlabels.SrcNode:         model.NewStringValue("src-node"),
						constlabels.DstNode:         model.NewStringValue("dst-node"),
						constlabels.DnatIp:          model.NewStringValue("dnat-ip"),
						constlabels.DnatPort:        model.NewIntValue(80),

						// isSlow
						constlabels.IsSlow: model.NewBoolValue(false),
					}),
				123,
				[]*model.Gauge{
					{constvalues.RequestTotalTime, 123},
					{constvalues.RequestIo, 456},
				}...),
			},
			want: model.NewAttributeMapWithValues(map[string]model.AttributeValue{
				// instanceInfo is moved from agg
				constlabels.Ip:   model.NewStringValue("dst-ip"),
				constlabels.Port: model.NewIntValue(8080),
				// protocolInfo
				constlabels.Protocol:        model.NewStringValue("http"),
				constlabels.RequestContent:  model.NewStringValue("/test"),
				constlabels.ResponseContent: model.NewStringValue("200"),

				// k8sInfo
				constlabels.WorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.Namespace:    model.NewStringValue("dst-Namespace"),
				constlabels.WorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.Service:      model.NewStringValue("dst-service"),
				constlabels.Node:         model.NewStringValue("dst-node"),
				constlabels.Pod:          model.NewStringValue("dst-pod"),

				"const-labels1": model.NewStringValue("const-values1"),
			}),
		},
		{
			name:           "kindling_agg_net_entity",
			labelConverter: baseAdapter.aggEntityAdapter[0],
			args: args{group: model.NewGaugeGroup(
				constnames.AggregatedNetRequestGaugeGroup,
				model.NewAttributeMapWithValues(
					map[string]model.AttributeValue{
						// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
						constlabels.SrcIp:   model.NewStringValue("src-ip"),
						constlabels.SrcPort: model.NewIntValue(33333),
						constlabels.DstIp:   model.NewStringValue("dst-ip"),
						constlabels.DstPort: model.NewIntValue(8080),

						// protocolInfo
						constlabels.Protocol:       model.NewStringValue("http"),
						constlabels.ContentKey:     model.NewStringValue("/test"),
						constlabels.HttpStatusCode: model.NewIntValue(200),

						// k8sInfo
						constlabels.DstPod:          model.NewStringValue("dst-pod"),
						constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
						constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
						constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
						constlabels.SrcPod:          model.NewStringValue("src-pod"),
						constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
						constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
						constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
						constlabels.SrcService:      model.NewStringValue("src-service"),
						constlabels.DstService:      model.NewStringValue("dst-service"),
						constlabels.SrcNode:         model.NewStringValue("src-node"),
						constlabels.DstNode:         model.NewStringValue("dst-node"),
						constlabels.DnatIp:          model.NewStringValue("dnat-ip"),
						constlabels.DnatPort:        model.NewIntValue(80),

						// isSlow
						constlabels.IsSlow: model.NewBoolValue(false),
					}),
				123,
				[]*model.Gauge{
					{constvalues.RequestTotalTime, 123},
					{constvalues.RequestIo, 456},
				}...),
			},
			want: model.NewAttributeMapWithValues(map[string]model.AttributeValue{
				// protocolInfo
				constlabels.Protocol:        model.NewStringValue("http"),
				constlabels.RequestContent:  model.NewStringValue("/test"),
				constlabels.ResponseContent: model.NewStringValue("200"),

				// k8sInfo
				constlabels.WorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.Namespace:    model.NewStringValue("dst-Namespace"),
				constlabels.WorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.Service:      model.NewStringValue("dst-service"),

				"const-labels1": model.NewStringValue("const-values1"),
			}),
		},
	}
	go func() {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m := tt.labelConverter
				got, free, err := m.transform(tt.args.group)
				if (err != nil) != tt.wantErr {
					t.Errorf("transform() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				for key, value := range got.GetValues() {
					if valueW, ok := tt.want.GetValues()[key]; ok {
						if value.ToString() != valueW.ToString() {
							t.Errorf("transform() get = %v, want %v", value.ToString(), valueW.ToString())
						}
					} else {
						if value.ToString() != "" {
							t.Errorf("transform() get = '%v =  %v', don't want this label", key, value.ToString())
						}
					}
				}
				for key, value := range tt.want.GetValues() {
					if _, ok := got.GetValues()[key]; !ok {
						t.Errorf("transform() expected key '%v' ,value '%v',but not exist", key, value.ToString())
					}
				}
				free(got)
			})
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.labelConverter
			got, free, err := m.transform(tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for key, value := range got.GetValues() {
				if valueW, ok := tt.want.GetValues()[key]; ok {
					if value.ToString() != valueW.ToString() {
						t.Errorf("transform() get = %v, want %v", value.ToString(), valueW.ToString())
					}
				} else {
					if value.ToString() != "" {
						t.Errorf("transform() get = '%v =  %v', don't want this label", key, value.ToString())
					}
				}
			}
			for key, value := range tt.want.GetValues() {
				if _, ok := got.GetValues()[key]; !ok {
					t.Errorf("transform() expected key '%v' ,value '%v',but not exist", key, value.ToString())
				}
			}
			free(got)
		})
	}
}
