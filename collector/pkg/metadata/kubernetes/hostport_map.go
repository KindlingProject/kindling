package kubernetes

import "sync"

type IpPortKey struct {
	Ip   string
	Port uint32
}

type HostPortMap struct {
	hostPortInfo sync.Map
}

func newHostPortMap() *HostPortMap {
	return &HostPortMap{
		hostPortInfo: sync.Map{},
	}
}

func (m *HostPortMap) add(ip string, port uint32, containerInfo *K8sContainerInfo) {
	key := IpPortKey{ip, port}
	m.hostPortInfo.Store(key, containerInfo)
}

func (m *HostPortMap) get(ip string, port uint32) (*K8sContainerInfo, bool) {
	key := IpPortKey{ip, port}
	containerInfo, ok := m.hostPortInfo.Load(key)
	if !ok {
		return nil, false
	}
	return containerInfo.(*K8sContainerInfo), true
}

func (m *HostPortMap) delete(ip string, port uint32) {
	key := IpPortKey{ip, port}
	m.hostPortInfo.Delete(key)
}
