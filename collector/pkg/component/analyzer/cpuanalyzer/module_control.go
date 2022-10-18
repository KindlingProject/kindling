package cpuanalyzer

import "sync"

func (ca *CpuAnalyzer) ProfileModule() (submodule string, start func() error, stop func() error) {
	return "cpuanalyzer", ca.StartProfile, ca.StopProfile
}

func (ca *CpuAnalyzer) StartProfile() error {
	// control flow changed
	// Note that these two variables belongs to the package
	sendChannel = make(chan SendTriggerEvent, 3e5)
	enableProfile = true
	once = sync.Once{}
	go ca.ReceiveSendSignal()
	return nil
}

func (ca *CpuAnalyzer) StopProfile() error {
	// control flow changed
	ca.lock.Lock()
	defer ca.lock.Unlock()
	enableProfile = false
	// Clear the old events even if they are not sent
	ca.cpuPidEvents = make(map[uint32]map[uint32]TimeSegments, 100000)
	return nil
}
