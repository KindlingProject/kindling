package aggregateprocessor

import "time"

type timer struct {
	firstDataTimestamp uint64
	createTimestamp    int64
}

func newTimer(firstDataTimestamp uint64) timer {
	return timer{
		firstDataTimestamp: firstDataTimestamp,
		createTimestamp:    time.Now().UnixNano(),
	}
}

func (t *timer) outputTimestamp() uint64 {
	currentTime := time.Now().UnixNano()
	diffDuration := currentTime - t.createTimestamp
	// The date could have been updated
	if diffDuration < 0 {
		diffDuration = 0
	}
	return t.firstDataTimestamp + uint64(diffDuration)
}
