package cameraexporter

import (
	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/filepathhelper"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestWriteTrace(t *testing.T) {
	fileConfig := &fileConfig{
		StoragePath:  "/tmp/kindling/",
		MaxFileCount: 20,
	}
	writer, err := newFileWriter(fileConfig, component.NewDefaultTelemetryTools().Logger)
	if err != nil {
		t.Fatal(err)
	}
	pid := 13403
	timestamp := 1660790515752561400
	// Test file rotating
	for i := 0; i < 100; i++ {
		traceData := traceData(int64(pid), uint64(timestamp+i))
		writer.writeTrace(traceData)
		for j := 0; j < 10; j++ {
			keyPathElements := filepathhelper.GetFilePathElements(traceData, uint64(timestamp+i))
			writer.writeCpuEvents(cpuEvent(int64(1660790510000000000+j*1e9), keyPathElements.ToAttributes()))
		}
	}
	pathElements := filepathhelper.GetFilePathElements(traceData(int64(pid), uint64(timestamp)), uint64(timestamp))
	filePath := writer.pidFilePath(pathElements.WorkloadName, pathElements.PodName, pathElements.ContainerName, pathElements.Pid)
	filesName, err := getFilesName(filePath)
	assert.NoError(t, err)
	assert.Equal(t, fileConfig.MaxFileCount, len(filesName))
}

func traceData(pid int64, timestamp uint64) *model.DataGroup {
	labels := model.NewAttributeMap()
	labels.AddIntValue(constlabels.Pid, pid)
	labels.AddStringValue(constlabels.ContentKey, "/cpu/hot")
	labels.AddStringValue(constlabels.HttpApmTraceId, "88786150748bd45008f7510696bce47c4^391")
	labels.AddStringValue(constlabels.HttpApmTraceType, "harmonycloud")
	labels.AddBoolValue(constlabels.IsServer, true)
	labels.AddStringValue(constlabels.Protocol, "http")
	return model.NewDataGroup("test", labels, timestamp)
}

func cpuEvent(startTime int64, extraLabels *model.AttributeMap) *model.DataGroup {
	labels := model.NewAttributeMap()
	labels.AddIntValue(constlabels.Pid, int64(13403))
	labels.AddIntValue(constlabels.Tid, int64(13600))
	labels.AddStringValue("threadName", "XNIO-1 task-3")
	labels.AddIntValue("startTime", startTime)
	labels.AddIntValue("endTime", startTime+1e9)
	labels.AddStringValue("cpuEvents", "[{\"stack\": \"\",\"log\": \"\",\"typeSpecs\": \"56651993,4709969826,2611411,1315845777,\",\"onInfo\": \"net@write@10.10.103.148:55738->10.10.103.148:9999@1660790509723744474@77660@1@296#net@write@10.10.103.148:55738->10.10.103.148:9999@1660790509724168721@52734@1@5|net@write@10.10.103.148:55744->10.10.103.148:9999@1660790514436177779@90940@1@297#net@write@10.10.103.148:55744->10.10.103.148:9999@1660790514436641680@71964@1@5|\",\"timeType\": \"0,3,0,3,\",\"startTime\": 1660790509667848205,\"endTime\": 1660790515752931068,\"runqLatency\": \"0,0,\",\"offInfo\": \"futex@addr140553945667444@1660790509724485534@4710000334|futex@addr140553945667444@1660790514437063928@1315884467|\"}]")
	labels.AddStringValue("javaFutexEvents", "[{\"dataValue\": \"kd-jf@1660790509724364808!1660790514434516975!13600!f501cc48!UnsafePark!XNIO-1 task-3!4710152167!13529!park.Lsun/misc/Unsafe;parkNanos.Ljava/util/concurrent/locks/LockSupport;park.Lorg/jboss/threads/EnhancedQueueExecutor$PoolThreadNode;run.Lorg/jboss/threads/EnhancedQueueExecutor$ThreadBody;run.Ljava/lang/Thread;!\\n\",\"startTime\": 1660790509724364808,\"endTime\": 1660790514434516975}]")
	labels.AddStringValue("transactionIds", "[]")
	labels.Merge(extraLabels)
	return model.NewDataGroup(constnames.CameraEventGroupName, labels, uint64(startTime))
}

func triggerKey(data *model.DataGroup) string {
	timestamp := data.Timestamp
	isServer := data.Labels.GetBoolValue(constlabels.IsServer)
	var podName string
	if isServer {
		podName = data.Labels.GetStringValue(constlabels.DstPod)
	} else {
		podName = data.Labels.GetStringValue(constlabels.SrcPod)
	}
	if len(podName) == 0 {
		podName = "null"
	}
	var isServerString string
	if isServer {
		isServerString = "true"
	} else {
		isServerString = "false"
	}
	protocol := data.Labels.GetStringValue(constlabels.Protocol)
	return podName + "_" + isServerString + "_" + protocol + "_" + strconv.FormatUint(timestamp, 10)
}
