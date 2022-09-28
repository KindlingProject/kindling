package network

func (na *NetworkAnalyzer) ProfileModule() (submodule string, start func() error, stop func() error) {
	return "networkanalyzer", na.StartProfile, na.StopProfile
}

func (na *NetworkAnalyzer) StartProfile() error {
	// control flow changed
	na.enableProfile = true
	return nil
}

func (na *NetworkAnalyzer) StopProfile() error {
	// control flow changed
	na.enableProfile = false
	return nil
}
