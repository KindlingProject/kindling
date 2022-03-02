package nodemetricprocessor

import (
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/consumer/processor"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	"time"
)

const (
	Type = "nodemetricprocessor"
)

type NodeMetricProcessor struct {
	cfg          *Config
	nextConsumer consumer.Consumer

	telemetry *component.TelemetryTools
}

func New(config interface{}, telemetry *component.TelemetryTools, nextConsumer consumer.Consumer) processor.Processor {
	cfg, ok := config.(*Config)
	if !ok {
		telemetry.Logger.Error("Cannot convert Component config", zap.String("componentType", Type))
	}
	return &NodeMetricProcessor{
		cfg:          cfg,
		nextConsumer: nextConsumer,
		telemetry:    telemetry,
	}
}

func (p *NodeMetricProcessor) Consume(gaugeGroup *model.GaugeGroup) error {
	labels := gaugeGroup.Labels
	// Filter the data which labels is nil
	if labels == nil {
		return nil
	}
	isServer := labels.GetBoolValue(constlabels.IsServer)
	var role string
	if isServer {
		role = "server"
	} else {
		role = "client"
	}
	return p.process(gaugeGroup, role)
}

func (p *NodeMetricProcessor) process(gaugeGroup *model.GaugeGroup, role string) error {
	labels := gaugeGroup.Labels
	dstNodeIp := labels.GetStringValue(constlabels.DstNodeIp)
	srcNodeIp := labels.GetStringValue(constlabels.SrcNodeIp)
	if dstNodeIp == "" || srcNodeIp == "" {
		p.telemetry.Logger.Debug("dstNodeIp or srcNodeIp is empty which is not expected, skip: ", zap.String("gaugeGroup", gaugeGroup.String()))
		return nil
	}
	// NodeName could be empty
	dstNodeName := labels.GetStringValue(constlabels.DstNode)
	srcNodeName := labels.GetStringValue(constlabels.SrcNode)

	var retError error
	// For request, the transmit direction is SrcNode->DstNode
	requestIo, ok := gaugeGroup.GetGauge(constvalues.RequestIo)
	if ok {
		newLabels := model.NewAttributeMapWithValues(map[string]model.AttributeValue{
			constlabels.SrcNodeIp: model.NewStringValue(srcNodeIp),
			constlabels.SrcNode:   model.NewStringValue(srcNodeName),
			constlabels.DstNodeIp: model.NewStringValue(dstNodeIp),
			constlabels.DstNode:   model.NewStringValue(dstNodeName),
			"role":                model.NewStringValue(role),
		})
		newValue := &model.Gauge{
			Name:  "kindling_node_transmit_bytes_total",
			Value: requestIo.Value,
		}
		newGaugeGroup := model.NewGaugeGroup(
			constnames.NodeGaugeGroupName,
			newLabels,
			uint64(time.Now().UnixNano()),
			newValue,
		)
		err := p.nextConsumer.Consume(newGaugeGroup)
		if err != nil {
			retError = multierror.Append(retError, err)
		}
	}
	// For response, the transmit direction is DstNode->SrcNode
	responseIo, ok := gaugeGroup.GetGauge(constvalues.ResponseIo)
	if ok {
		newLabels := model.NewAttributeMapWithValues(map[string]model.AttributeValue{
			constlabels.SrcNodeIp: model.NewStringValue(dstNodeIp),
			constlabels.SrcNode:   model.NewStringValue(dstNodeName),
			constlabels.DstNodeIp: model.NewStringValue(srcNodeIp),
			constlabels.DstNode:   model.NewStringValue(srcNodeName),
			"role":                model.NewStringValue(role),
		})
		newValue := &model.Gauge{
			Name:  "kindling_node_transmit_bytes_total",
			Value: responseIo.Value,
		}
		newGaugeGroup := model.NewGaugeGroup(
			constnames.NodeGaugeGroupName,
			newLabels,
			uint64(time.Now().UnixNano()),
			newValue,
		)
		err := p.nextConsumer.Consume(newGaugeGroup)
		if err != nil {
			retError = multierror.Append(retError, err)
		}
	}

	return retError
}
