package cgoreceiver

func (r *CgoReceiver) StartProfile() error {
	// TODO Change EventMask in Probe
	return nil
}

func (r *CgoReceiver) StopProfile() error {
	// TODO Change EventMask in Probe
	return nil
}

func (r *CgoReceiver) ProfileModule() (submodule string, start func() error, stop func() error) {
	return "cgoreceiver", r.StartProfile, r.StopProfile
}
