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
	"go.uber.org/zap/zapcore"
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

func (p *NodeMetricProcessor) Consume(dataGroup *model.DataGroup) error {
	labels := dataGroup.Labels
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
	return p.process(dataGroup, role)
}

func (p *NodeMetricProcessor) process(dataGroup *model.DataGroup, role string) error {
	labels := dataGroup.Labels
	dstNodeIp := labels.GetStringValue(constlabels.DstNodeIp)
	srcNodeIp := labels.GetStringValue(constlabels.SrcNodeIp)
	if dstNodeIp == "" || srcNodeIp == "" {
		if ce := p.telemetry.Logger.Check(zapcore.DebugLevel, "dstNodeIp or srcNodeIp is empty which is not expected, skip: "); ce != nil {
			ce.Write(
				zap.String("dataGroup", dataGroup.String()),
			)
		}
		return nil
	}
	// NodeName could be empty
	dstNodeName := labels.GetStringValue(constlabels.DstNode)
	srcNodeName := labels.GetStringValue(constlabels.SrcNode)

	var retError error
	// For request, the transmit direction is SrcNode->DstNode
	requestIo, ok := dataGroup.GetMetric(constvalues.RequestIo)
	if ok {
		newLabels := model.NewAttributeMapWithValues(map[string]model.AttributeValue{
			constlabels.SrcNodeIp: model.NewStringValue(srcNodeIp),
			constlabels.SrcNode:   model.NewStringValue(srcNodeName),
			constlabels.DstNodeIp: model.NewStringValue(dstNodeIp),
			constlabels.DstNode:   model.NewStringValue(dstNodeName),
			"role":                model.NewStringValue(role),
		})
		newValue := model.NewIntMetric("kindling_node_transmit_bytes_total", requestIo.GetInt().Value)
		newMetricGroup := model.NewDataGroup(
			constnames.NodeMetricGroupName,
			newLabels,
			uint64(time.Now().UnixNano()),
			newValue,
		)
		err := p.nextConsumer.Consume(newMetricGroup)
		if err != nil {
			retError = multierror.Append(retError, err)
		}
	}
	// For response, the transmit direction is DstNode->SrcNode
	responseIo, ok := dataGroup.GetMetric(constvalues.ResponseIo)
	if ok {
		newLabels := model.NewAttributeMapWithValues(map[string]model.AttributeValue{
			constlabels.SrcNodeIp: model.NewStringValue(dstNodeIp),
			constlabels.SrcNode:   model.NewStringValue(dstNodeName),
			constlabels.DstNodeIp: model.NewStringValue(srcNodeIp),
			constlabels.DstNode:   model.NewStringValue(srcNodeName),
			"role":                model.NewStringValue(role),
		})
		newValue := model.NewIntMetric("kindling_node_transmit_bytes_total", responseIo.GetInt().Value)
		newMetricGroup := model.NewDataGroup(
			constnames.NodeMetricGroupName,
			newLabels,
			uint64(time.Now().UnixNano()),
			newValue,
		)
		err := p.nextConsumer.Consume(newMetricGroup)
		if err != nil {
			retError = multierror.Append(retError, err)
		}
	}

	return retError
}
