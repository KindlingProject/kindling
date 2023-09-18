package service

import (
	"github.com/Kindling-project/kindling/collector/pkg/metadata/metaprovider/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

type boardcast func(data api.MetaDataRsyncResponse)

type K8sResourceHandler struct {
	resType string

	add    api.AddObj
	update api.UpdateObj
	delete api.DeleteObj

	boardcast
}

func (m *K8sResourceHandler) AddObj(obj interface{}) {
	m.add(obj)
	m.boardcast(api.MetaDataRsyncResponse{
		Type:      m.resType,
		Operation: string(api.Add),
		NewObj:    obj,
	})
}

func (m *K8sResourceHandler) UpdateObj(objOld interface{}, objNew interface{}) {
	m.update(objOld, objNew)
	m.boardcast(api.MetaDataRsyncResponse{
		Type:      m.resType,
		Operation: string(api.Update),
		OldObj:    objOld,
		NewObj:    objNew,
	})
}

func (m *K8sResourceHandler) DeleteObj(obj interface{}) {
	m.delete(obj)
	m.boardcast(api.MetaDataRsyncResponse{
		Type:      m.resType,
		Operation: string(api.Delete),
		OldObj:    obj,
	})
}

func NewHandler(typeName string, add api.AddObj, update api.UpdateObj, delete api.DeleteObj, boardcast boardcast) cache.ResourceEventHandlerFuncs {
	handler := &K8sResourceHandler{
		resType:   "pod",
		add:       add,
		update:    update,
		delete:    delete,
		boardcast: boardcast,
	}

	if typeName == "pod" {
		clearBeforeAdd := func(obj interface{}) {
			decreasePodInfo(obj)
			handler.AddObj(obj)
		}

		return cache.ResourceEventHandlerFuncs{
			AddFunc:    clearBeforeAdd,
			UpdateFunc: handler.UpdateObj,
			DeleteFunc: handler.DeleteObj,
		}
	}

	return cache.ResourceEventHandlerFuncs{
		AddFunc:    handler.AddObj,
		UpdateFunc: handler.UpdateObj,
		DeleteFunc: handler.DeleteObj,
	}
}

func decreasePodInfo(objOld interface{}) {
	pod := objOld.(*corev1.Pod)
	//  Clear unnecessary Message
	pod.ManagedFields = nil
	pod.Spec.Volumes = nil
	pod.Status.Conditions = nil
}
