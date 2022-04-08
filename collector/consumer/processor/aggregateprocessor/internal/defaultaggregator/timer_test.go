package defaultaggregator

import (
	"testing"
	"time"
)

const (
	million = 1000000
	// 10 milliseconds
	allowedDiffDuration = 10 * million
)

type fields struct {
	firstDataTimestamp uint64
	createTimestamp    int64
	sleepDuration      time.Duration
}

type test struct {
	name   string
	fields fields
	want   uint64
}

func Test_timer_outputTimestamp(t *testing.T) {
	runTest(t, test{
		name: "zero timestamp",
		fields: fields{
			firstDataTimestamp: 0,
			createTimestamp:    time.Now().UnixNano(),
			sleepDuration:      100 * time.Millisecond,
		},
		want: 100 * million,
	})

	runTest(t, test{
		name: "normal timestamp",
		fields: fields{
			firstDataTimestamp: uint64(time.Now().UnixNano()),
			createTimestamp:    time.Now().UnixNano(),
			sleepDuration:      3 * time.Second,
		},
		want: uint64(time.Now().UnixNano()) + 3000*million,
	})
}

func runTest(t *testing.T, tt test) {
	t.Run(tt.name, func(t1 *testing.T) {
		t := &timer{
			firstDataTimestamp: tt.fields.firstDataTimestamp,
			createTimestamp:    tt.fields.createTimestamp,
		}
		time.Sleep(tt.fields.sleepDuration)
		if got := t.outputTimestamp(); got-tt.want < 0 || got-tt.want > allowedDiffDuration {
			t1.Errorf("actual difference is %v, want less than %v", got-tt.want, allowedDiffDuration)
		}
	})
}
