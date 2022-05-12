package kubernetes

import (
	"fmt"
	_ "path/filepath"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	_ "k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/client-go/util/homedir"
)

type PodInfo struct {
	Ip           string
	Ports        []int32
	Name         string
	Labels       map[string]string
	Namespace    string
	ContainerIds []string
}

type podMap struct {
	// namespace:
	//   podName: podInfo{}
	Info  map[string]map[string]*PodInfo
	mutex sync.RWMutex
}

var globalPodInfo = newPodMap()
var podUpdateMutex sync.Mutex

func newPodMap() *podMap {
	return &podMap{
		Info:  make(map[string]map[string]*PodInfo),
		mutex: sync.RWMutex{},
	}
}

func (m *podMap) add(info *PodInfo) {
	m.mutex.Lock()
	podInfoMap, ok := m.Info[info.Namespace]
	if !ok {
		podInfoMap = make(map[string]*PodInfo)
	}
	podInfoMap[info.Name] = info
	m.Info[info.Namespace] = podInfoMap
	m.mutex.Unlock()
}

func (m *podMap) delete(namespace string, name string) {
	m.mutex.Lock()
	podInfoMap, ok := m.Info[namespace]
	if ok {
		delete(podInfoMap, name)
	}
	m.mutex.Unlock()
}

func (m *podMap) get(namespace string, name string) (*PodInfo, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	podInfoMap, ok := m.Info[namespace]
	if !ok {
		return nil, false
	}
	podInfo, ok := podInfoMap[name]
	if !ok {
		return nil, false
	}
	return podInfo, true
}

// getPodsMatchSelectors gets PodInfo(s) whose labels match with selectors in such namespace.
// Return empty slice if not found. Note there may be multiple match.
func (m *podMap) getPodsMatchSelectors(namespace string, selectors map[string]string) []*PodInfo {
	retPodInfoSlice := make([]*PodInfo, 0)
	if len(selectors) == 0 {
		return retPodInfoSlice
	}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	podInfoMap, ok := m.Info[namespace]
	if !ok {
		return retPodInfoSlice
	}
	for _, podInfo := range podInfoMap {
		if SelectorsMatchLabels(selectors, podInfo.Labels) {
			retPodInfoSlice = append(retPodInfoSlice, podInfo)
		}
	}
	return retPodInfoSlice
}

func PodWatch(clientSet *kubernetes.Clientset, graceDeletePeriod time.Duration) {
	stopper := make(chan struct{})
	defer close(stopper)

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	podInformer := factory.Core().V1().Pods()
	informer := podInformer.Informer()
	defer runtime.HandleCrash()

	// Start informer, list & watch
	go factory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	go podDeleteLoop(10*time.Second, graceDeletePeriod, stopper)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: OnUpdate,
		DeleteFunc: onDelete,
	})
	// TODO: use workqueue to avoid blocking
	<-stopper
}

func onAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)
	pI := PodInfo{
		Ip:           pod.Status.PodIP,
		Ports:        make([]int32, 0, 1),
		Name:         pod.Name,
		Labels:       pod.Labels,
		Namespace:    pod.Namespace,
		ContainerIds: make([]string, 0, 2),
	}

	workloadTypeTmp := ""
	workloadNameTmp := ""

	for _, owner := range pod.OwnerReferences {
		// only care about the controller
		if owner.Controller == nil || *owner.Controller != true {
			continue
		}
		// TODO: recursion method to find the controller
		if owner.Kind == ReplicaSetKind {
			// The owner of Pod is ReplicaSet, and it is Workload such as Deployment for ReplicaSet.
			// Therefore, find ReplicaSet's name in 'globalRsInfo' to find which kind of workload
			// the Pod belongs to.
			if workload, ok := globalRsInfo.GetOwnerReference(mapKey(pod.Namespace, owner.Name)); ok {
				workloadTypeTmp = CompleteGVK(workload.APIVersion, strings.ToLower(workload.Kind))
				workloadNameTmp = workload.Name
				break
			}
		}
		// If the owner of pod is not ReplicaSet or the replicaset has no controller
		workloadTypeTmp = CompleteGVK(owner.APIVersion, strings.ToLower(owner.Kind))
		workloadNameTmp = owner.Name
		break
	}

	serviceInfoSlice := globalServiceInfo.GetServiceMatchLabels(pI.Namespace, pI.Labels)
	var serviceInfo *K8sServiceInfo
	if len(serviceInfoSlice) == 0 {
		serviceInfo = nil
	} else {
		// When span target is a kind of service, workload should also be filled in case to display
		// the real topology in model level. Service is considered as abstract level, instead.
		// So here the information of workload is assigned to serviceInfo.
		for _, service := range serviceInfoSlice {
			service.WorkloadKind = workloadTypeTmp
			service.WorkloadName = workloadNameTmp
		}
		// Only one of the matched services is cared, here we get the first one
		serviceInfo = serviceInfoSlice[0]
	}
	var kpi = &K8sPodInfo{
		Ip:            pod.Status.PodIP,
		Namespace:     pod.Namespace,
		PodName:       pod.Name,
		WorkloadKind:  workloadTypeTmp,
		WorkloadName:  workloadNameTmp,
		NodeName:      pod.Spec.NodeName,
		NodeAddress:   pod.Status.HostIP,
		isHostNetwork: pod.Spec.HostNetwork,
		ServiceInfo:   serviceInfo,
	}

	// Add containerId map
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerId := containerStatus.ContainerID
		realContainerId := TruncateContainerId(containerId)
		if realContainerId == "" {
			continue
		}
		pI.ContainerIds = append(pI.ContainerIds, realContainerId)
		containerInfo := &K8sContainerInfo{
			ContainerId: realContainerId,
			Name:        containerStatus.Name,
			RefPodInfo:  kpi,
		}
		MetaDataCache.AddByContainerId(realContainerId, containerInfo)
	}

	// Add pod IP and port map
	if len(pod.Status.PodIP) > 0 {
		for _, tmpContainer := range pod.Spec.Containers {
			containerInfo := &K8sContainerInfo{
				Name:       tmpContainer.Name,
				RefPodInfo: kpi,
			}
			if len(tmpContainer.Ports) == 0 {
				// When there are many pods in one pod and only some of them have ports,
				// the containers at the back will overwrite the one at the front here.
				MetaDataCache.AddContainerByIpPort(pod.Status.PodIP, 0, containerInfo)
			}
			for _, port := range tmpContainer.Ports {
				pI.Ports = append(pI.Ports, port.ContainerPort)
				// If hostPort is specified, add the container using HostIP and HostPort
				if port.HostPort != 0 {
					MetaDataCache.AddContainerByIpPort(pod.Status.HostIP, uint32(port.HostPort), containerInfo)
				}
				MetaDataCache.AddContainerByIpPort(pod.Status.PodIP, uint32(port.ContainerPort), containerInfo)
			}
		}
	}
	globalPodInfo.add(&pI)
}

func OnUpdate(objOld interface{}, objNew interface{}) {
	oldPod := objOld.(*corev1.Pod)
	newPod := objNew.(*corev1.Pod)
	if oldPod.ResourceVersion == newPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}
	podUpdateMutex.Lock()
	// TODO: re-implement the updated logic to reduce computation
	onDelete(objOld)
	onAdd(objNew)
	podUpdateMutex.Unlock()
}

func onDelete(obj interface{}) {
	pod := obj.(*corev1.Pod)
	// Delete the intermediate data first
	globalPodInfo.delete(pod.Namespace, pod.Name)
	// Wait for a few seconds to remove the cache data
	podDeleteQueueMut.Lock()
	podDeleteQueue = append(podDeleteQueue, deleteRequest{
		pod: pod,
		ts:  time.Now(),
	})
	podDeleteQueueMut.Unlock()
}

// TruncateContainerId slices the input containerId into two parts separated by "://",
// and return the first 12 bytes at most of the second part.
//
// If no second part found, return empty string.
func TruncateContainerId(containerId string) string {
	sep := "://"
	separated := strings.SplitN(containerId, sep, 2)
	if len(separated) < 2 {
		return ""
	}
	secondString := separated[1]
	l := len(secondString)
	if l > 12 {
		l = 12
	}
	return secondString[0:l]
}
