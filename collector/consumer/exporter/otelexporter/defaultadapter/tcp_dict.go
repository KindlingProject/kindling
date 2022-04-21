package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

var tcpBaseDictList = []dictionary{
	{constlabels.SrcNode, constlabels.SrcNode, String},
	{constlabels.SrcNodeIp, constlabels.SrcNodeIp, String},
	{constlabels.SrcNamespace, constlabels.SrcNamespace, String},
	{constlabels.SrcPod, constlabels.SrcPod, String},
	{constlabels.SrcWorkloadName, constlabels.SrcWorkloadName, String},
	{constlabels.SrcWorkloadKind, constlabels.SrcWorkloadKind, String},
	{constlabels.SrcService, constlabels.SrcService, String},
	{constlabels.SrcIp, constlabels.SrcIp, String},
	{constlabels.SrcPort, constlabels.SrcPort, Int64},
	{constlabels.SrcContainerId, constlabels.SrcContainerId, String},
	{constlabels.SrcContainer, constlabels.SrcContainer, String},
	{constlabels.DstNode, constlabels.DstNode, String},
	{constlabels.DstNodeIp, constlabels.DstNodeIp, String},
	{constlabels.DstNamespace, constlabels.DstNamespace, String},
	{constlabels.DstPod, constlabels.DstPod, String},
	{constlabels.DstWorkloadName, constlabels.DstWorkloadName, String},
	{constlabels.DstWorkloadKind, constlabels.DstWorkloadKind, String},
	{constlabels.DstService, constlabels.DstService, String},
	{constlabels.DstIp, constlabels.DstIp, String},
	{constlabels.DstPort, constlabels.DstPort, Int64},
	{constlabels.DnatIp, constlabels.DnatIp, String},
	{constlabels.DnatPort, constlabels.DnatPort, Int64},
	{constlabels.DstContainerId, constlabels.DstContainerId, String},
	{constlabels.DstContainer, constlabels.DstContainer, String},
}
