package kubernetes

import (
	"testing"
	"time"
)

func TestDeleteLoop(t *testing.T) {
	pod := CreatePod(true)
	onAdd(pod)
	_, ok := globalPodInfo.get("CustomNamespace", "deploy-1a2b3c4d-5e6f7")
	if !ok {
		t.Fatalf("Finding pod at globalPodInfo. Expect %v, but get %v", true, ok)
	}
	verifyIfPodExist(true, t)
	if len(podDeleteQueue) != 0 {
		t.Fatalf("PodDeleteQueue should be 0, but is %d", len(podDeleteQueue))
	}

	onDelete(pod)
	_, ok = globalPodInfo.get("CustomNamespace", "deploy-1a2b3c4d-5e6f7")
	if ok {
		t.Fatalf("Finding pod at globalPodInfo. Expect %v, but get %v", false, ok)
	}
	verifyIfPodExist(true, t)
	if len(podDeleteQueue) != 1 {
		t.Fatalf("PodDeleteQueue should be 1, but is %d", len(podDeleteQueue))
	}

	gracePeriod := time.Millisecond * 500
	stopCh := make(chan struct{})
	go podDeleteLoop(time.Millisecond, gracePeriod, stopCh)
	go func() {
		time.Sleep(time.Millisecond * 50)
		verifyIfPodExist(true, t)
		podDeleteQueueMut.Lock()
		if len(podDeleteQueue) != 1 {
			t.Errorf("PodDeleteQueue should be 1, but is %d", len(podDeleteQueue))
		}
		podDeleteQueueMut.Unlock()

		time.Sleep(gracePeriod + (time.Millisecond * 50))
		verifyIfPodExist(false, t)
		podDeleteQueueMut.Lock()
		if len(podDeleteQueue) != 0 {
			t.Errorf("PodDeleteQueue should be 0, but is %d", len(podDeleteQueue))
		}
		podDeleteQueueMut.Unlock()
		close(stopCh)
	}()
	<-stopCh
}

func verifyIfPodExist(exist bool, t *testing.T) {
	_, ok := MetaDataCache.GetByContainerId("1a2b3c4d5e6f")
	if ok != exist {
		t.Errorf("Finding container using containerid. Expect %v, but get %v", exist, ok)
	}
	_, ok = MetaDataCache.GetContainerByIpPort("172.10.1.2", 80)
	if ok != exist {
		t.Errorf("Finding container using IP:Port. Expect %v, but get %v", exist, ok)
	}
}
