package cpuanalyzer

func (ca *CpuAnalyzer) ProfileModule() (submodule string, start func() error, stop func() error) {
	return "cpuanalyzer", ca.StartProfile, ca.StopProfile
}

func (ca *CpuAnalyzer) StartProfile() error {
	// control flow changed
	ca.enableProfile = true
	return nil
}

func (ca *CpuAnalyzer) StopProfile() error {
	// control flow changed
	ca.enableProfile = false

	// TODO Clear Expired Event
	return nil
}
