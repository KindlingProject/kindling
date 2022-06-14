package network

import (
	"testing"

	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model"
)

type benchCase struct {
	protocol   string
	commonFile string
	dataFile   string
}

const (
	BENCH_CASE_HTTP           = "http"
	BENCH_CASE_MYSQL          = "mysql"
	BENCH_CASE_REDIS          = "redis"
	BENCH_CASE_DNS            = "dns"
	BENCH_CASE_KAFKA_PRODUCER = "kafka_producer"
	BENCH_CASE_KAFKA_FETCHER  = "kafka_fetcher"
	BENCH_CASE_DUBBO2         = "dubbo2"
)

var benchCaseMap = map[string]benchCase{
	BENCH_CASE_HTTP:           {protocol.HTTP, "http/server-event.yml", "http/1k-trace.yml"},
	BENCH_CASE_MYSQL:          {protocol.MYSQL, "mysql/server-event.yml", "mysql/1k-trace.yml"},
	BENCH_CASE_REDIS:          {protocol.REDIS, "redis/server-event.yml", "redis/1k-trace.yml"},
	BENCH_CASE_DNS:            {protocol.DNS, "dns/server-event.yml", "dns/1k-trace.yml"},
	BENCH_CASE_KAFKA_PRODUCER: {protocol.KAFKA, "kafka/provider-event.yml", "kafka/1k-provider-trace.yml"},
	BENCH_CASE_KAFKA_FETCHER:  {protocol.KAFKA, "kafka/consumer-event.yml", "kafka/1k-consumer-trace.yml"},
	BENCH_CASE_DUBBO2:         {protocol.DUBBO2, "dubbo2/server-event.yml", "dubbo2/1k-trace.yml"},
}

const (
	SIZE_MESSAGE_PAIR int = 500
)

func BenchmarkHttp(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_HTTP)
}

func BenchmarkMySql(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_MYSQL)
}

func BenchmarkRedis(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_REDIS)
}

func BenchmarkDns(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_DNS)
}

func BenchmarkKafkaProducer(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_KAFKA_PRODUCER)
}

func BenchmarkKafkaFetcher(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_KAFKA_FETCHER)
}

func BenchmarkDubo2(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_DUBBO2)
}

func testProtocolBench(b *testing.B, tps int, mpSize int, caseKey string) {
	na := prepareNetworkAnalyzer()
	if na == nil {
		return
	}

	benchCase := benchCaseMap[caseKey]
	eventCommon := getEventCommon("protocol/testdata/" + benchCase.commonFile)
	if eventCommon == nil {
		b.Errorf("Parse %v Failed", benchCase.commonFile)
		return
	}

	trace := getTrace("protocol/testdata/" + benchCase.dataFile)
	if trace == nil {
		b.Errorf("Parse %v Failed", benchCase.dataFile)
		return
	}

	// Prepare Base Events
	evts := prepareEvents(mpSize, eventCommon, trace)
	// Produce Event
	for i := 0; i < tps; i++ {
		na.ConsumeEvent(evts[i%mpSize])
	}
}

func prepareEvents(mpSize int, eventCommon *EventCommon, trace *Trace) []*model.KindlingEvent {
	baseEvents := make([]*model.KindlingEvent, 0)
	for _, event := range trace.Events {
		baseEvents = append(baseEvents, event.exchange(eventCommon))

	}

	size := len(baseEvents)
	events := make([]*model.KindlingEvent, size*mpSize)
	for i := 0; i < mpSize; i++ {
		for j, basebaseEvent := range baseEvents {
			events[i*size+j] = prepareEvent(basebaseEvent, i+1)
		}
	}
	return events
}

func prepareEvent(evt *model.KindlingEvent, fdNum int) *model.KindlingEvent {
	newEvt := &model.KindlingEvent{
		Source:           evt.Source,
		Timestamp:        evt.Timestamp,
		Name:             evt.Name,
		Category:         evt.Category,
		NativeAttributes: evt.NativeAttributes,
		UserAttributes:   evt.UserAttributes,
		Ctx: &model.Context{
			ThreadInfo: evt.Ctx.ThreadInfo,
			FdInfo: &model.Fd{
				Num:      int32(fdNum),
				TypeFd:   evt.Ctx.FdInfo.TypeFd,
				Protocol: evt.Ctx.FdInfo.Protocol,
				Role:     evt.Ctx.FdInfo.Role,
				Sip:      evt.Ctx.FdInfo.Sip,
				Dip:      evt.Ctx.FdInfo.Dip,
				Sport:    evt.Ctx.FdInfo.Sport,
				Dport:    evt.Ctx.FdInfo.Dport,
			},
		},
	}
	return newEvt
}
