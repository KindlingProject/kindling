package cpuanalyzer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduledTask(t *testing.T) {
	// test case 1: Normal expired exit
	task1 := &testIncreamentTask{0}
	routine1 := NewAndStartScheduledTaskRoutine(1*time.Millisecond, 5*time.Millisecond, task1, nil)
	routine1.Start()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, routine1.isRunning.Load())

	assert.Equal(t, 5, task1.count)

	// Case 2: Reset expired time when not expired
	task2 := &testIncreamentTask{0}
	routine2 := NewAndStartScheduledTaskRoutine(1*time.Millisecond, 5*time.Millisecond, task2, nil)
	time.Sleep(2 * time.Millisecond)
	routine2.ResetExpiredTimer()
	time.Sleep(2 * time.Millisecond)
	routine2.ResetExpiredTimer()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, routine1.isRunning.Load())
	assert.Equal(t, 9, task2.count)

	// Case 3: Reset expired time when it is expired
	routine2.ResetExpiredTimer()
	assert.Equal(t, false, routine1.isRunning.Load())

	// Case 4: Double start or double stop
	task3 := &testIncreamentTask{0}
	routine3 := NewAndStartScheduledTaskRoutine(1*time.Millisecond, 5*time.Millisecond, task3, nil)
	err := routine3.Start()
	assert.Error(t, err)
	err = routine3.Stop()
	assert.NoError(t, err)
	err = routine3.Stop()
	assert.Error(t, err)
}

type testIncreamentTask struct {
	count int
}

func (t *testIncreamentTask) run() {
	t.count++
}
