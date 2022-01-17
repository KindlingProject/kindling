package k8sprocessor

import (
	"github.com/dxsup/kindling-collector/consumer"
	"github.com/dxsup/kindling-collector/consumer/processor"
	"github.com/dxsup/kindling-collector/metadata/kubernetes"
	"github.com/dxsup/kindling-collector/model"
	"github.com/dxsup/kindling-collector/model/constlabels"
	"github.com/dxsup/kindling-collector/model/constnames"
	"go.uber.org/zap"
)

const (
	K8sMetadata = "k8smetadataprocessor"
	loopbackIp  = "127.0.0.1"
)

type K8sMetadataProcessor struct {
	metadata      *kubernetes.K8sMetaDataCache
	nextConsumer  consumer.Consumer
	localNodeIp   string
	localNodeName string
	logger        *zap.Logger
}

func NewKubernetesProcessor(cfg interface{}, logger *zap.Logger, nextConsumer consumer.Consumer) processor.Processor {
	config, ok := cfg.(*Config)
	if !ok {
		logger.Panic("Cannot convert Component config", zap.String("componentType", K8sMetadata))
	}
	var options []kubernetes.Option
	options = append(options, kubernetes.WithAuthType(config.KubeAuthType))
	options = append(options, kubernetes.WithKubeConfigDir(config.KubeConfigDir))
	err := kubernetes.InitK8sHandler(options...)
	if err != nil {
		logger.Sugar().Panicf("Failed to initialize [%s]: %v", K8sMetadata, err)
		return nil
	}

	var localNodeIp, localNodeName string
	if localNodeIp, err = getHostIpFromEnv(); err != nil {
		logger.Warn("Local NodeIp can not found", zap.Error(err))
	}
	if localNodeName, err = getHostNameFromEnv(); err != nil {
		logger.Warn("Local NodeName can not found", zap.Error(err))
	}
	return &K8sMetadataProcessor{
		metadata:      kubernetes.MetaDataCache,
		nextConsumer:  nextConsumer,
		localNodeIp:   localNodeIp,
		localNodeName: localNodeName,
		logger:        logger,
	}
}

type Config struct {
	KubeAuthType  kubernetes.AuthType `mapstructure:"kube_auth_type"`
	KubeConfigDir string              `mapstructure:"kube_config_dir"`
}

func (p *K8sMetadataProcessor) Consume(gaugeGroup *model.GaugeGroup) error {
	name := gaugeGroup.Name
	switch name {
	case constnames.NetRequestGaugeGroupName:
		p.processNetRequestMetric(gaugeGroup)
	case constnames.TcpGaugeGroupName:
		p.processTcpMetric(gaugeGroup)
	default:
		p.processNetRequestMetric(gaugeGroup)
	}
	return p.nextConsumer.Consume(gaugeGroup)
}

func (p *K8sMetadataProcessor) processNetRequestMetric(gaugeGroup *model.GaugeGroup) {
	isServer := gaugeGroup.Labels.GetBoolValue(constlabels.IsServer)
	if isServer {
		p.addK8sMetaDataForServerLabel(gaugeGroup.Labels)
	} else {
		p.addK8sMetaDataForClientLabel(gaugeGroup.Labels)
	}
}

func (p *K8sMetadataProcessor) processTcpMetric(gaugeGroup *model.GaugeGroup) {
	p.addK8sMetaDataViaIp(gaugeGroup.Labels)
}

func (p *K8sMetadataProcessor) addK8sMetaDataForClientLabel(labelMap *model.AttributeMap) {
	// add metadata for src
	containerId := labelMap.GetStringValue(constlabels.ContainerId)
	labelMap.AddStringValue(constlabels.SrcContainerId, containerId)
	resInfo, ok := p.metadata.GetByContainerId(containerId)
	if ok {
		addContainerMetaInfoLabelSRC(labelMap, resInfo)
	} else {
		labelMap.AddStringValue(constlabels.SrcNodeIp, p.localNodeIp)
		labelMap.AddStringValue(constlabels.SrcNode, p.localNodeName)
		labelMap.AddStringValue(constlabels.SrcNamespace, constlabels.ExternalClusterNamespace)
	}
	// add metadata for dst
	dstIp := labelMap.GetStringValue(constlabels.DstIp)
	if dstIp == loopbackIp {
		labelMap.AddStringValue(constlabels.DstNodeIp, p.localNodeIp)
		labelMap.AddStringValue(constlabels.DstNode, p.localNodeName)
	}
	dstPort := labelMap.GetIntValue(constlabels.DstPort)
	// DstIp is IP of a service
	if svcInfo, ok := p.metadata.GetServiceByIpPort(dstIp, uint32(dstPort)); ok {
		labelMap.AddStringValue(constlabels.DstNamespace, svcInfo.Namespace)
		labelMap.AddStringValue(constlabels.DstService, svcInfo.ServiceName)
		labelMap.AddStringValue(constlabels.DstWorkloadKind, svcInfo.WorkloadKind)
		labelMap.AddStringValue(constlabels.DstWorkloadName, svcInfo.WorkloadName)
		// find podInfo using dnat_ip
		dNatIp := labelMap.GetStringValue(constlabels.DnatIp)
		dNatPort := labelMap.GetIntValue(constlabels.DnatPort)
		if dNatIp != "" && dNatPort != -1 {
			resInfo, ok := p.metadata.GetContainerByIpPort(dNatIp, uint32(dNatPort))
			if ok {
				addContainerMetaInfoLabelDST(labelMap, resInfo)
			} else {
				// maybe dnat_ip is NodeIP
				if nodeName, ok := p.metadata.GetNodeNameByIp(dNatIp); ok {
					labelMap.AddStringValue(constlabels.DstNodeIp, dNatIp)
					labelMap.AddStringValue(constlabels.DstNode, nodeName)
				}
			}
		}
	} else if resInfo, ok := p.metadata.GetContainerByIpPort(dstIp, uint32(dstPort)); ok {
		// DstIp is IP of a container
		addContainerMetaInfoLabelDST(labelMap, resInfo)
	} else {
		// DstIp is a IP from external
		if nodeName, ok := p.metadata.GetNodeNameByIp(dstIp); ok {
			labelMap.AddStringValue(constlabels.DstNodeIp, dstIp)
			labelMap.AddStringValue(constlabels.DstNode, nodeName)
		}
		labelMap.AddStringValue(constlabels.DstNamespace, constlabels.ExternalClusterNamespace)
	}
}

func (p *K8sMetadataProcessor) addK8sMetaDataForServerLabel(labelMap *model.AttributeMap) {
	srcIp := labelMap.GetStringValue(constlabels.SrcIp)
	if srcIp == loopbackIp {
		labelMap.AddStringValue(constlabels.SrcNodeIp, p.localNodeIp)
		labelMap.AddStringValue(constlabels.SrcNode, p.localNodeName)
	}
	podInfo, ok := p.metadata.GetPodByIp(srcIp)
	if ok {
		addPodMetaInfoLabelSRC(labelMap, podInfo)
	} else {
		if nodeName, ok := p.metadata.GetNodeNameByIp(srcIp); ok {
			labelMap.AddStringValue(constlabels.SrcNodeIp, srcIp)
			labelMap.AddStringValue(constlabels.SrcNode, nodeName)
		}
		labelMap.AddStringValue(constlabels.SrcNamespace, constlabels.ExternalClusterNamespace)
	}
	containerId := labelMap.GetStringValue(constlabels.ContainerId)
	labelMap.AddStringValue(constlabels.DstContainerId, containerId)
	containerInfo, ok := p.metadata.GetByContainerId(containerId)
	if ok {
		addContainerMetaInfoLabelDST(labelMap, containerInfo)
		if containerInfo.RefPodInfo.ServiceInfo != nil {
			labelMap.AddStringValue(constlabels.DstService, containerInfo.RefPodInfo.ServiceInfo.ServiceName)
		}
	} else {
		labelMap.AddStringValue(constlabels.DstNodeIp, p.localNodeIp)
		labelMap.AddStringValue(constlabels.DstNode, p.localNodeName)
		labelMap.AddStringValue(constlabels.DstNamespace, constlabels.ExternalClusterNamespace)
	}
}

func (p *K8sMetadataProcessor) addK8sMetaDataViaIp(labelMap *model.AttributeMap) {
	// Both Src and Dst should try:
	// 1. (Only Dst)Use Ip Port to find Service (when found a Service,also use DNatIp to find the Pod)
	// 2. Use Ip Port to find Container And Pod
	// 3. Use Ip to find Pod

	// add metadata for src
	p.addK8sMetaDataViaIpSRC(labelMap)
	// add metadata for dst
	p.addK8sMetaDataViaIpDST(labelMap)
}

func (p *K8sMetadataProcessor) addK8sMetaDataViaIpSRC(labelMap *model.AttributeMap) {
	// 1. Use Ip Port to find Container And Pod
	// 2. Use Ip to find Pod
	defer labelMap.RemoveAttribute(constlabels.SrcPort)

	srcIp := labelMap.GetStringValue(constlabels.SrcIp)
	srcPort := labelMap.GetIntValue(constlabels.SrcPort)
	srcContainerInfo, ok := p.metadata.GetContainerByIpPort(srcIp, uint32(srcPort))
	if ok {
		addContainerMetaInfoLabelSRC(labelMap, srcContainerInfo)
		return
	}

	srcPodInfo, ok := p.metadata.GetPodByIp(srcIp)
	if ok {
		addPodMetaInfoLabelSRC(labelMap, srcPodInfo)
		return
	}

	labelMap.AddStringValue(constlabels.SrcNamespace, constlabels.ExternalClusterNamespace)
}

func (p *K8sMetadataProcessor) addK8sMetaDataViaIpDST(labelMap *model.AttributeMap) {
	// 1. (Only Dst)Use Ip Port to find Service (when found a Service,also use DNatIp to find the Pod)
	// 2. Use Ip Port to find Container And Pod
	// 3. Use Ip to find Pod
	defer labelMap.RemoveAttribute(constlabels.DstPort)
	defer labelMap.RemoveAttribute(constlabels.DnatPort)

	dstIp := labelMap.GetStringValue(constlabels.DstIp)
	dstPort := labelMap.GetIntValue(constlabels.DstPort)
	dstSvcInfo, ok := p.metadata.GetServiceByIpPort(dstIp, uint32(dstPort))
	if ok {
		labelMap.AddStringValue(constlabels.DstNamespace, dstSvcInfo.Namespace)
		labelMap.AddStringValue(constlabels.DstService, dstSvcInfo.ServiceName)
		labelMap.AddStringValue(constlabels.DstWorkloadKind, dstSvcInfo.WorkloadKind)
		labelMap.AddStringValue(constlabels.DstWorkloadName, dstSvcInfo.WorkloadName)
		// find podInfo using dnat_ip
		dNatIp := labelMap.GetStringValue(constlabels.DnatIp)
		dNatPort := labelMap.GetIntValue(constlabels.DnatPort)
		if dNatIp != "" && dNatPort != -1 {
			resInfo, ok := p.metadata.GetContainerByIpPort(dNatIp, uint32(dNatPort))
			if ok {
				addContainerMetaInfoLabelDST(labelMap, resInfo)
			}
		}
		return
	}

	dstContainerInfo, ok := p.metadata.GetContainerByIpPort(dstIp, uint32(dstPort))
	if ok {
		addContainerMetaInfoLabelDST(labelMap, dstContainerInfo)
		return
	}

	dstPodInfo, ok := p.metadata.GetPodByIp(dstIp)
	if ok {
		addPodMetaInfoLabelDST(labelMap, dstPodInfo)
		return
	}

	labelMap.AddStringValue(constlabels.DstNamespace, constlabels.ExternalClusterNamespace)
}

func addContainerMetaInfoLabelSRC(labelMap *model.AttributeMap, containerInfo *kubernetes.K8sContainerInfo) {
	labelMap.AddStringValue(constlabels.SrcContainer, containerInfo.Name)
	addPodMetaInfoLabelSRC(labelMap, containerInfo.RefPodInfo)
}

func addPodMetaInfoLabelSRC(labelMap *model.AttributeMap, podInfo *kubernetes.K8sPodInfo) {
	labelMap.AddStringValue(constlabels.SrcNode, podInfo.NodeName)
	labelMap.AddStringValue(constlabels.SrcNodeIp, podInfo.NodeAddress)
	labelMap.AddStringValue(constlabels.SrcNamespace, podInfo.Namespace)
	labelMap.AddStringValue(constlabels.SrcWorkloadKind, podInfo.WorkloadKind)
	labelMap.AddStringValue(constlabels.SrcWorkloadName, podInfo.WorkloadName)
	labelMap.AddStringValue(constlabels.SrcPod, podInfo.PodName)
	labelMap.AddStringValue(constlabels.SrcIp, podInfo.Ip)
	if podInfo.ServiceInfo != nil {
		labelMap.AddStringValue(constlabels.SrcService, podInfo.ServiceInfo.ServiceName)
	}
}

func addContainerMetaInfoLabelDST(labelMap *model.AttributeMap, containerInfo *kubernetes.K8sContainerInfo) {
	labelMap.AddStringValue(constlabels.DstContainer, containerInfo.Name)
	addPodMetaInfoLabelDST(labelMap, containerInfo.RefPodInfo)
}

func addPodMetaInfoLabelDST(labelMap *model.AttributeMap, podInfo *kubernetes.K8sPodInfo) {
	labelMap.AddStringValue(constlabels.DstNode, podInfo.NodeName)
	labelMap.AddStringValue(constlabels.DstNodeIp, podInfo.NodeAddress)
	labelMap.AddStringValue(constlabels.DstNamespace, podInfo.Namespace)
	labelMap.AddStringValue(constlabels.DstWorkloadKind, podInfo.WorkloadKind)
	labelMap.AddStringValue(constlabels.DstWorkloadName, podInfo.WorkloadName)
	labelMap.AddStringValue(constlabels.DstPod, podInfo.PodName)
	if labelMap.GetStringValue(constlabels.DstIp) == "" {
		labelMap.AddStringValue(constlabels.DstIp, podInfo.Ip)
	}
}
