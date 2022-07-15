package network

import (
	"sync"
	"time"

	"github.com/Kindling-project/kindling/collector/metadata/conntracker"
	"github.com/Kindling-project/kindling/collector/model"
)

const (
	LOWER32 = 0x00000000FFFFFFFF
	LOWER16 = 0x000000000000FFFF
)

type mergableEvent struct {
	events []*model.KindlingEvent // Keep No more than 10 events.

	latency uint64
	resVal  int64
	ts      uint64
	data    []byte
}

type events struct {
	event    *model.KindlingEvent
	mergable *mergableEvent
}

func newEvents(evt *model.KindlingEvent) *events {
	return &events{
		event: evt,
	}
}

func (evts *events) size() int {
	if evts.mergable == nil {
		return 1
	}
	return len(evts.mergable.events)
}

func (evts *events) getEvent(index int) *model.KindlingEvent {
	if index == 0 {
		return evts.event
	}
	if index > 0 && evts.mergable != nil && index < len(evts.mergable.events) {
		return evts.mergable.events[index]
	}
	return nil
}

func (evts *events) mergeEvent(evt *model.KindlingEvent) {
	if evts.mergable == nil {
		firstEvt := evts.event
		evts.mergable = &mergableEvent{
			events:  []*model.KindlingEvent{firstEvt},
			latency: firstEvt.GetLatency(),
			resVal:  firstEvt.GetResVal(),
			ts:      firstEvt.Timestamp,
			data:    firstEvt.GetData(),
		}
	}

	if len(evts.mergable.events) < 10 {
		// persistent connect
		evts.mergable.events = append(evts.mergable.events, evt)
	}
	evts.mergable.latency = evt.Timestamp - evts.mergable.ts + evts.mergable.latency
	evts.mergable.resVal += evt.GetResVal()
	evts.mergable.ts = evt.Timestamp

	if len(evts.mergable.data) < 80 {
		var appendLength int
		newData := evt.GetData()
		if 80-len(evts.mergable.data) > len(newData) {
			appendLength = len(newData)
		} else {
			appendLength = 80 - len(evts.mergable.data)
		}
		evts.mergable.data = append(evts.mergable.data, newData[0:appendLength]...)
	}
}

func (evts *events) getData() []byte {
	if evts.mergable == nil {
		return evts.event.GetData()
	}
	return evts.mergable.data
}

func (evts *events) getFirstTimestamp() uint64 {
	return evts.event.Timestamp
}

func (evts *events) getLastTimestamp() uint64 {
	if evts.mergable == nil {
		return evts.event.Timestamp
	}
	return evts.mergable.ts
}

func (evts *events) IsTimeout(newEvt *model.KindlingEvent, timeout int) bool {
	// old event is stale if timeout or sPort mismatched
	// be careful about overflow
	firstEvt := evts.event
	firstStartTime := firstEvt.Timestamp - firstEvt.GetLatency()
	if newEvt.Timestamp < firstStartTime {
		return false
	}
	if newEvt.Timestamp-firstStartTime > uint64(timeout)*uint64(time.Second) || newEvt.GetSport() != firstEvt.GetSport() {
		return true
	}
	return false
}

func (evts *events) getDuration() uint64 {
	if evts == nil {
		return 0
	}

	if evts.mergable == nil {
		return evts.event.GetLatency()
	}

	return evts.mergable.latency
}

type messagePairs struct {
	connects  *events
	requests  *events
	responses *events
	natTuple  *conntracker.IPTranslation
	isSend    bool
	mutex sync.RWMutex // only for update latency and resval now
}

func (mps *messagePairs) getKey() messagePairKey {
	if mps.connects != nil {
		return getMessagePairKey(mps.connects.event)
	} else if mps.requests != nil {
		return getMessagePairKey(mps.requests.event)
	} else if mps.responses != nil {
		return getMessagePairKey(mps.responses.event)
	}
	return messagePairKey{}
}

func (mps *messagePairs) mergeConnect(evt *model.KindlingEvent) {
	mps.mutex.Lock()
	if mps.requests == nil {
		mps.connects = newEvents(evt)
	} else {
		mps.connects.mergeEvent(evt)
	}
	mps.mutex.Unlock()
}

func (mps *messagePairs) mergeRequest(evt *model.KindlingEvent) {
	mps.mutex.Lock()
	if mps.requests == nil {
		mps.requests = newEvents(evt)
	} else {
		mps.requests.mergeEvent(evt)
	}
	mps.mutex.Unlock()
}

func (mps *messagePairs) mergeResponse(evt *model.KindlingEvent) {
	mps.mutex.Lock()
	if mps.responses == nil {
		mps.responses = newEvents(evt)
	} else {
		mps.responses.mergeEvent(evt)
	}
	mps.mutex.Unlock()
}

func (mps *messagePairs) getPort() uint32 {
	if mps.requests != nil {
		return mps.requests.event.GetDport()
	}
	if mps.responses != nil {
		return mps.responses.event.GetDport()
	}
	return 0
}

func (mps *messagePairs) getTimeoutTs() uint64 {
	if mps.responses != nil {
		return mps.responses.getLastTimestamp()
	}
	if mps.requests != nil {
		return mps.requests.getLastTimestamp()
	}
	if mps.connects != nil {
		return mps.connects.getLastTimestamp()
	}
	return 0
}

func (mps *messagePairs) getConnectDuration() uint64 {
	if mps.connects == nil {
		return 0
	}

	return mps.connects.getDuration()
}

func (mps *messagePairs) getSentTime() int64 {
	if mps.requests == nil {
		return -1
	}

	return int64(mps.requests.getDuration())
}

func (mps *messagePairs) getWaitingTime() int64 {
	if mps.responses == nil {
		return -1
	}

	return int64(mps.responses.getFirstTimestamp() - mps.responses.event.GetLatency() - mps.requests.getLastTimestamp())
}

func (mps *messagePairs) getDownloadTime() int64 {
	if mps.responses == nil {
		return -1
	}

	return int64(mps.responses.getDuration())
}

func (mps *messagePairs) getDuration() uint64 {
	if mps.responses == nil {
		return 0
	}

	return mps.responses.getLastTimestamp() - mps.requests.event.GetStartTime()
}

func (mps *messagePairs) getRquestSize() uint64 {
	if mps.requests == nil {
		return 0
	}

	if mps.requests.mergable == nil {
		return uint64(mps.requests.event.GetResVal())
	}
	return uint64(mps.requests.mergable.resVal)
}

func (mps *messagePairs) getResponseSize() uint64 {
	if mps.responses == nil {
		return 0
	}

	if mps.responses.mergable == nil {
		return uint64(mps.responses.event.GetResVal())
	}
	return uint64(mps.responses.mergable.resVal)
}

type messagePair struct {
	request  *model.KindlingEvent
	response *model.KindlingEvent
}

func (mp *messagePair) getSentTime() int64 {
	if mp.request == nil {
		return -1
	}

	return int64(mp.request.GetLatency())
}

func (mp *messagePair) getWaitingTime() int64 {
	if mp.response == nil {
		return -1
	}

	return int64(mp.response.Timestamp - mp.response.GetLatency() - mp.request.Timestamp)
}

func (mp *messagePair) getDownloadTime() int64 {
	if mp.response == nil {
		return -1
	}

	return int64(mp.response.GetLatency())
}

func (mp *messagePair) getRquestSize() uint64 {
	if mp.request == nil {
		return 0
	}
	return uint64(mp.request.GetResVal())
}

func (mp *messagePair) getResponseSize() uint64 {
	if mp.response == nil {
		return 0
	}
	return uint64(mp.response.GetResVal())
}

func (mp *messagePair) getDuration() uint64 {
	if mp.response == nil {
		return 0
	}

	return mp.response.Timestamp + mp.request.GetLatency() - mp.request.Timestamp
}

// DNS will send different ip and port data with sharing fd and pid socket.
type messagePairKey struct {
	pid   uint32
	fd    int32
	sip   string
	dip   string
	sport uint32
	dport uint32
}

func getMessagePairKey(evt *model.KindlingEvent) messagePairKey {
	if evt.IsUdp() == 1 {
		return messagePairKey{
			pid:   evt.GetPid(),
			fd:    evt.GetFd(),
			sip:   evt.GetSip(),
			dip:   evt.GetDip(),
			sport: evt.GetSport(),
			dport: evt.GetDport(),
		}
	} else {
		return messagePairKey{
			pid: evt.GetPid(),
			fd:  evt.GetFd(),
		}
	}
}
