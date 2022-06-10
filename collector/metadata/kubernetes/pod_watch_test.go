package kubernetes

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTruncateContainerId(t *testing.T) {
	testCases := []struct {
		containerId string
		expected    string
	}{
		{"docker://a1b2c3d4e5f6g7h8i9j0k1l2m3n", "a1b2c3d4e5f6"},
		{"docker://a1b2c3d4e5f6", "a1b2c3d4e5f6"},
		{"docker://a1b2c3", "a1b2c3"},
		{"containerd://a1b2c3d4e5f6g7h8i9j0k1l2m3n", "a1b2c3d4e5f6"},
		{"a1b2c3", ""},
		{"", ""},
	}
	for _, test := range testCases {
		res := TruncateContainerId(test.containerId)
		if res != test.expected {
			t.Errorf("containerId=%s, get=%s, but expected=%s", test.containerId, res, test.expected)
		}
	}
}

func TestOnAdd(t *testing.T) {
	globalPodInfo = &podMap{
		Info: make(map[string]map[string]*K8sPodInfo),
	}
	globalServiceInfo = &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
	}
	globalRsInfo = &ReplicaSetMap{
		Info: make(map[string]Controller),
	}
	// First add service, and then add pod
	onAddService(CreateService())
	onAddReplicaSet(CreateReplicaSet())
	onAdd(CreatePod(true))
	t.Log(MetaDataCache)
	// Delete podInfo must not affect serviceMap
	onDelete(CreatePod(true))
	t.Log(MetaDataCache)
	// Empty all the metadata
	onDeleteService(CreateService())
	t.Log(MetaDataCache)
}

// ISSUE https://github.com/CloudDectective-Harmonycloud/kindling/issues/229
func TestOnAddPodWhileReplicaSetUpdating(t *testing.T) {
	globalPodInfo = &podMap{
		Info: make(map[string]map[string]*K8sPodInfo),
	}
	globalServiceInfo = &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
	}
	globalRsInfo = &ReplicaSetMap{
		Info: make(map[string]Controller),
	}
	// Firstly deployment created and add old RS and old POD
	controller := true
	oldRs := CreateReplicaSet()
	oldRs.SetResourceVersion("old")
	newRs := CreateReplicaSet()
	newRs.SetResourceVersion("new")
	oldPOD := CreatePod(true)
	oldPOD.SetResourceVersion("old")
	oldPOD.OwnerReferences[0].Controller = &controller
	newPOD := CreatePod(true)
	newPOD.SetResourceVersion("new")
	newPOD.OwnerReferences[0].Controller = &controller
	onAddReplicaSet(oldRs)
	onAdd(oldPOD)

	// Secondly POD&RS were been updated

	go func() {
		for i := 0; i < 1000; i++ {
			OnUpdateReplicaSet(oldRs, newRs)
		}
	}()

	for i := 0; i < 100; i++ {
		OnUpdate(oldPOD, newPOD)
		// Thirdly check the pod's workload_kind
		pod, ok := MetaDataCache.GetPodByContainerId(TruncateContainerId(newPOD.Status.ContainerStatuses[0].ContainerID))
		require.True(t, ok, "failed to get target POD")
		require.Equal(t, "deployment", pod.WorkloadKind, "failed to get the real workload_kind")
	}
}

func TestOnAddLowercaseWorkload(t *testing.T) {
	globalPodInfo = &podMap{
		Info: make(map[string]map[string]*K8sPodInfo),
	}
	globalServiceInfo = &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
	}
	globalRsInfo = &ReplicaSetMap{
		Info: make(map[string]Controller),
	}
	higherCase := "DaemonSet"
	lowerCase := "daemonset"
	isController := true
	onAdd(&corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       higherCase,
				Controller: &isController,
			}},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container1",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
				},
			}},
		Status: corev1.PodStatus{
			PodIP: "172.10.1.2",
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "container1",
					ContainerID: "docker://1a2b3c4d5e6f7g8h9i0j1k2",
				},
			},
		},
	})
	podInfo, ok := MetaDataCache.GetPodByContainerId("1a2b3c4d5e6f")
	if !ok || podInfo.WorkloadKind != lowerCase {
		t.Errorf("%s wanted, but get %s", higherCase, lowerCase)
	}
}

func CreatePod(hasPort bool) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy-1a2b3c4d-5e6f7",
			Namespace: "CustomNamespace",
			Labels: map[string]string{
				"a": "1",
				"b": "1",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: ReplicaSetKind,
					Name: "deploy-1a2b3c4d",
				},
			},
		},
		Spec: corev1.PodSpec{
			NodeName:    "node1",
			HostNetwork: false,
		},
		Status: corev1.PodStatus{
			PodIP: "172.10.1.2",
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "container1",
					ContainerID: "docker://1a2b3c4d5e6f7g8h9i0j1k2",
				},
			},
		},
	}
	if hasPort {
		pod.Spec.Containers = []corev1.Container{
			{
				Name: "container1",
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 80,
					},
				},
			},
		}
	} else {
		pod.Spec.Containers = []corev1.Container{
			{
				Name: "container1",
			},
		}
	}
	return pod
}

func TestUpdateAndDelayDelete(t *testing.T) {
	addObjJson := "{\"metadata\":{\"name\":\"testdemo2-5c86748464-26crb\",\"namespace\":\"test-ns\",\"resourceVersion\":\"44895976\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.210\",\"hostIP\":\"10.10.10.101\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:36Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d505f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	updateObjJson := "{\"metadata\":{\"name\":\"testdemo2-5c86748464-26crb\",\"namespace\":\"test-ns\",\"resourceVersion\":\"44895977\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.212\",\"hostIP\":\"10.10.10.102\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:36Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d000f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	addObj := new(corev1.Pod)
	err := json.Unmarshal([]byte(addObjJson), addObj)
	if err != nil {
		t.Errorf("error unmarshalling %v", err)
	}
	updateObj := new(corev1.Pod)
	err = json.Unmarshal([]byte(updateObjJson), updateObj)
	if err != nil {
		t.Fatalf("error unmarshalling %v", err)
	}
	podIp := addObj.Status.PodIP
	port := addObj.Spec.Containers[0].Ports[0].ContainerPort
	onAdd(addObj)
	_, ok := MetaDataCache.GetContainerByIpPort(podIp, uint32(port))
	if !ok {
		t.Fatalf("Not found container [%s:%d]", podIp, port)
	} else {
		t.Logf("Found container [%s:%d]", podIp, port)
	}
	stopCh := make(chan struct{})
	go podDeleteLoop(100*time.Millisecond, 500*time.Millisecond, stopCh)
	OnUpdate(addObj, updateObj)

	// Check if new Container can be find
	_, find := MetaDataCache.GetByContainerId(TruncateContainerId(updateObj.Status.ContainerStatuses[0].ContainerID))
	assert.True(t, find, "NewContainerId did't find in MetaDataCache")
	_, find = MetaDataCache.GetContainerByIpPort(updateObj.Status.PodIP, uint32(port))
	assert.True(t, find, "NewContainer IP Port did't find in MetaDataCache")
	_, find = MetaDataCache.GetContainerByHostIpPort(updateObj.Status.HostIP, uint32(port))
	assert.True(t, find, "NewHostIp Port did't find in MetaDataCache")

	// Wait for deletes
	time.Sleep(1000 * time.Millisecond)

	// Double Check for NewContainer
	_, find = MetaDataCache.GetByContainerId(TruncateContainerId(updateObj.Status.ContainerStatuses[0].ContainerID))
	assert.True(t, find, "NewContainerId did't find in MetaDataCache")
	_, find = MetaDataCache.GetContainerByIpPort(updateObj.Status.PodIP, uint32(port))
	assert.True(t, find, "NewContainer IP Port did't find in MetaDataCache")
	_, find = MetaDataCache.GetContainerByHostIpPort(updateObj.Status.HostIP, uint32(port))
	assert.True(t, find, "NewHostIp Port did't find in MetaDataCache")

	// Check the old Container has been delete
	_, find = MetaDataCache.GetByContainerId(TruncateContainerId(addObj.Status.ContainerStatuses[0].ContainerID))
	assert.False(t, find, "OldContainerId should be deletedin MetaDataCache")
	_, find = MetaDataCache.GetContainerByIpPort(addObj.Status.PodIP, uint32(port))
	assert.False(t, find, "OldContainer IP should be deleted in MetaDataCache")
	_, find = MetaDataCache.GetContainerByHostIpPort(addObj.Status.HostIP, uint32(port))
	assert.False(t, find, "OldHostIp Port should be deleted in MetaDataCache")

	stopCh <- struct{}{}
}
