package kubernetes

import (
	"sync"
)

type IpPortKey struct {
	Ip   string
	Port uint32
}

type HostPortMap struct {
	HostPortInfo map[IpPortKey]*K8sContainerInfo
	mutex        sync.RWMutex
}

func NewHostPortMap() *HostPortMap {
	return &HostPortMap{
		HostPortInfo: make(map[IpPortKey]*K8sContainerInfo),
	}
}

func (m *HostPortMap) add(ip string, port uint32, containerInfo *K8sContainerInfo) {
	key := IpPortKey{ip, port}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.HostPortInfo[key] = containerInfo
}

func (m *HostPortMap) get(ip string, port uint32) (*K8sContainerInfo, bool) {
	key := IpPortKey{ip, port}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	containerInfo, ok := m.HostPortInfo[key]
	if !ok {
		return nil, false
	}
	return containerInfo, true
}

func (m *HostPortMap) delete(ip string, port uint32) {
	key := IpPortKey{ip, port}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.HostPortInfo, key)
}
