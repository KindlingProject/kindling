package cpuanalyzer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/atomic"
)

var eventsWindowsDuration = 5 * time.Second
var SendChannel chan SendTriggerEvent

func init() {
	SendChannel = make(chan SendTriggerEvent, 3e5)
}

func ReceiveSendSignal(event SendTriggerEvent) {
	SendChannel <- event
}

type SendTriggerEvent struct {
	Pid       uint32 `json:"pid"`
	StartTime uint64 `json:"startTime"`
	SpendTime uint64 `json:"spendTime"`
}

func (ca *CpuAnalyzer) SendCircle() {
	for {
		sendContent := <-SendChannel
		if sendContent.StartTime+sendContent.SpendTime+uint64(10*nanoToSeconds) > uint64(time.Now().UnixNano()) {
			SendChannel <- sendContent
			time.Sleep(300 * time.Millisecond)
			continue
		}
		profilePid := os.Getenv("profilepid")
		if profilePid != "" {
			pidInt, _ := strconv.ParseInt(profilePid, 10, 32)
			if pidInt != int64(sendContent.Pid) {
				continue
			}
		}
		data, _ := json.Marshal(sendContent)
		ca.telemetry.Logger.Sugar().Infof("Receive a trace signal: %s", string(data))
		fmt.Println("start send ::" + strconv.Itoa(int(sendContent.StartTime/nanoToSeconds)) + "spend:" + strconv.Itoa(int(sendContent.SpendTime/nanoToSeconds)) + "now time: " + strconv.Itoa(int(time.Now().UnixNano()/int64(nanoToSeconds))))
		ca.SendCpuEvent(sendContent.Pid, sendContent.StartTime, sendContent.SpendTime)
	}
}

func (ca *CpuAnalyzer) ReceiveSendSignal() {
	for {
		sendContent := <-SendChannel
		task := &SendEventsTask{ca, &sendContent}
		// "Load" and "Delete" could be executed concurrently
		value, ok := ca.sendEventsRoutineMap.Load(sendContent.Pid)
		if !ok {
			ca.putNewRoutine(task)
		} else {
			// The routine may be found before but stopped and deleted after entering this branch.
			// So we much check twice.
			routine, _ := value.(*ScheduledTaskRoutine)
			// TODO Always replacing the task may cause that some events are skipped when the "spendTime"
			// becomes much smaller than before.
			err := routine.ResetExpiredTimerWithNewTask(task)
			if err != nil {
				// The routine has been expired.
				ca.putNewRoutine(task)
			}
		}
	}
}

func (ca *CpuAnalyzer) putNewRoutine(task *SendEventsTask) {
	expiredCallback := func() {
		ca.sendEventsRoutineMap.Delete(task.triggerEvent.Pid)
	}
	routine := NewAndStartScheduledTaskRoutine(1*time.Second, eventsWindowsDuration, task, expiredCallback)
	ca.sendEventsRoutineMap.Store(task.triggerEvent.Pid, routine)
}

func (ca *CpuAnalyzer) SendSegment(segment Segment) {
	ca.esClient.Index().Index("cpu_event").Type("_doc").BodyJson(segment).Do(context.Background())
}

func (ca *CpuAnalyzer) SendCpuEvent(pid uint32, startTime uint64, spendTime uint64) error {
	tmpTid := uint32(13717)
	ca.lock.Lock()
	defer ca.lock.Unlock()
	ca.telemetry.Logger.Sugar().Infof("Will send cpu events for pid=%d, start_time=%d, duration=%d", pid, startTime, spendTime)

	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		fmt.Println("send data0")
		ca.telemetry.Logger.Sugar().Infof("Not found the cpu events with the pid=%d", pid)
		return nil
	}
	for _, timeSegments := range tidCpuEvents {
		if timeSegments.Tid == tmpTid {
			fmt.Println("send data1")
		}
		if timeSegments.BaseTime+uint64(ca.cfg.GetSegmentSize()) < startTime/nanoToSeconds || timeSegments.BaseTime > startTime/nanoToSeconds {
			if timeSegments.Tid == tmpTid {
				fmt.Println("-----------------")
				fmt.Println("basetime:" + strconv.Itoa(int(timeSegments.BaseTime)))
				fmt.Println("starttime:" + strconv.Itoa(int(startTime/nanoToSeconds)))
				fmt.Println("-----------------")
			}
			continue
		}
		if timeSegments.Tid == tmpTid {
			fmt.Println("send data2 start time:" + strconv.Itoa(int(startTime/nanoToSeconds)) + "spend:" + strconv.Itoa(int(spendTime/nanoToSeconds)) + "now time: " + strconv.Itoa(int(time.Now().UnixNano()/int64(nanoToSeconds))))
		}
		for i := 0; i < int(spendTime/nanoToSeconds)+1+2; i++ {
			index := int(startTime/nanoToSeconds-timeSegments.BaseTime) + i - 2
			if index < 0 {
				index = 0
			}
			val, _ := timeSegments.Segments.GetByIndex(index)
			if val == nil {
				continue
			}
			segment := val.(*Segment)

			if len(segment.CpuEvents) != 0 && segment.IsSend != 1 {
				if segment.Tid == tmpTid {
					fmt.Println("send data3:" + strconv.Itoa(int(startTime/nanoToSeconds)+i-2))
				}
				segment.IsSend = 1
				segment.IndexTimestamp = time.Now().String()
				ca.esClient.Index().Index("cpu_event").Type("_doc").BodyJson(segment).Do(context.Background())
			}
		}
	}
	return nil
}

type ScheduledTask interface {
	run()
}

type ScheduledTaskRoutine struct {
	expiredDuration time.Duration
	ticker          *time.Ticker
	timer           *time.Timer
	stopCh          chan struct{}

	task      ScheduledTask
	isRunning *atomic.Bool

	expiredCallback func()
}

// NewAndStartScheduledTaskRoutine creates a new routine and start it immediately.
func NewAndStartScheduledTaskRoutine(
	tickerDuration time.Duration,
	expiredDuration time.Duration,
	task ScheduledTask,
	expiredCallback func()) *ScheduledTaskRoutine {
	ret := &ScheduledTaskRoutine{
		expiredDuration: expiredDuration,
		ticker:          time.NewTicker(tickerDuration),
		timer:           time.NewTimer(expiredDuration),
		task:            task,
		isRunning:       atomic.NewBool(false),
		stopCh:          make(chan struct{}),
		expiredCallback: expiredCallback,
	}
	// Start the routine once it is created.
	ret.Start()
	return ret
}

func (s *ScheduledTaskRoutine) Start() error {
	swapped := s.isRunning.CAS(false, true)
	if !swapped {
		return errors.New("the routine has been started")
	}
	go func() {
		if s.expiredCallback != nil {
			defer s.expiredCallback()
		}
		for {
			select {
			case <-s.ticker.C:
				// do some work
				s.task.run()
			case <-s.timer.C:
				// The current task is expired.
				s.isRunning.CAS(true, false)
				s.ticker.Stop()
				return
			case <-s.stopCh:
				s.timer.Stop()
				s.ticker.Stop()
				return
			}
		}
	}()
	return nil
}

// ResetExpiredTimer resets the timer to extend its expired time if it is running.
// If the routine is not running, an error will be returned and nothing will happen.
func (s *ScheduledTaskRoutine) ResetExpiredTimer() error {
	if !s.isRunning.Load() {
		return errors.New("the routine is not running, can't reset the timer")
	}
	if !s.timer.Stop() {
		<-s.timer.C
	}
	s.timer.Reset(s.expiredDuration)
	return nil
}

func (s *ScheduledTaskRoutine) ResetExpiredTimerWithNewTask(task ScheduledTask) error {
	s.task = task
	return s.ResetExpiredTimer()
}

func (s *ScheduledTaskRoutine) Stop() error {
	swapped := s.isRunning.CAS(true, false)
	if !swapped {
		return errors.New("the routine is not running")
	}
	s.stopCh <- struct{}{}
	return nil
}

type SendEventsTask struct {
	cpuAnalyzer  *CpuAnalyzer
	triggerEvent *SendTriggerEvent
}

// |________________|______________|_________________|
// 0      (5s)      1  (duration)  2      (5s)       3
// 0: The start time of the windows where the events we need are.
// 1: The start time of the "trace".
// 2: The end time of the "trace". This is nerely equal to the creating time of the task.
// 3: The end time of the windows where the events we need are.
func (t *SendEventsTask) run() {
	currentNanoTime := uint64(time.Now().UnixNano())
	currentWindowsStartTime := currentNanoTime - t.triggerEvent.SpendTime - uint64(eventsWindowsDuration)
	currentWindowsEndTime := currentNanoTime
	t.cpuAnalyzer.sendEvents(t.triggerEvent.Pid, currentWindowsStartTime, currentWindowsEndTime)
}

func (ca *CpuAnalyzer) sendEvents(pid uint32, startTime uint64, endTime uint64) {
	ca.lock.Lock()
	defer ca.lock.Unlock()

	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		ca.telemetry.Logger.Sugar().Infof("Not found the cpu events with the pid=%d, startTime=%d, endTime=%d",
			pid, startTime, endTime)
		return
	}
	startTimeSecond := startTime / nanoToSeconds
	endTimeSecond := endTime / nanoToSeconds

	for _, timeSegments := range tidCpuEvents {
		if endTimeSecond < timeSegments.BaseTime || startTimeSecond > timeSegments.BaseTime+uint64(ca.cfg.GetSegmentSize()) {
			continue
		}
		startIndex := int(startTimeSecond - timeSegments.BaseTime)
		if startIndex < 0 {
			startIndex = 0
		}
		endIndex := endTimeSecond - timeSegments.BaseTime
		if endIndex > timeSegments.BaseTime+uint64(ca.cfg.GetSegmentSize()) {
			endIndex = timeSegments.BaseTime + uint64(ca.cfg.GetSegmentSize())
		}
		ca.telemetry.Logger.Sugar().Infof("pid=%d tid=%d sends events. startSecond=%d, endSecond=%d",
			pid, timeSegments.Tid, startTimeSecond, endTimeSecond)
		for i := startIndex; i <= int(endIndex); i++ {
			val, _ := timeSegments.Segments.GetByIndex(i)
			if val == nil {
				continue
			}
			segment := val.(*Segment)
			if len(segment.CpuEvents) != 0 && segment.IsSend != 1 {
				segment.IsSend = 1
				segment.IndexTimestamp = time.Now().String()
				ca.esClient.Index().Index("cpu_event").Type("_doc").BodyJson(segment).Do(context.Background())
			}
		}
	}
}
