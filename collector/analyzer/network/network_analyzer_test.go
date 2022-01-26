package network

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/logger"

	"github.com/Kindling-project/kindling/collector/model"
	"github.com/spf13/viper"
)

func TestHttpProtocol(t *testing.T) {
	testProtocol(t, "http/server-event.yml",
		"http/server-trace-slow.yml",
		"http/server-trace-error.yml",
		"http/server-trace-normal.yml")
}

func TestMySqlProtocol(t *testing.T) {
	testProtocol(t, "mysql/server-event.yml",
		"mysql/server-trace-query-split.yml",
		"mysql/server-trace-query.yml")
}

func TestRedisProtocol(t *testing.T) {
	testProtocol(t, "redis/server-event.yml",
		"redis/server-trace-get.yml")
}

func TestDnsProtocol(t *testing.T) {
	testProtocol(t, "dns/server-event.yml",
		"dns/server-trace-multi.yml")
}

func TestKafkaProtocol(t *testing.T) {
	testProtocol(t, "kafka/provider-event.yml",
		"kafka/provider-trace-produce-split.yml")

	testProtocol(t, "kafka/consumer-event.yml",
		"kafka/consumer-trace-fetch-split.yml")
}

type NopProcessor struct {
}

func (n NopProcessor) Consume(gaugeGroup *model.GaugeGroup) error {
	// fmt.Printf("Consume %v\n", gaugeGroup)
	return nil
}

var na *NetworkAnalyzer

func prepareNetworkAnalyzer() *NetworkAnalyzer {
	if na == nil {
		config := &Config{}
		viper := viper.New()
		viper.SetConfigFile("protocol/testdata/na-protocol-config.yaml")
		err := viper.ReadInConfig()
		if err != nil {
			fmt.Printf("Read Config File failed%v\n", err)
			return nil
		}
		viper.UnmarshalKey("analyzers.networkanalyzer", config)

		var loggerConfig = logger.Config{
			ConsoleLogLevel: "none",
			FileLogLevel:    "none",
		}
		na = &NetworkAnalyzer{
			cfg:           config,
			nextConsumers: []consumer.Consumer{&NopProcessor{}},
			logger:        logger.InitLogger(loggerConfig),
		}
		na.Start()
	}
	return na
}

func testProtocol(t *testing.T, eventYaml string, traceYamls ...string) {
	na := prepareNetworkAnalyzer()
	if na == nil {
		return
	}

	eventCommon := getEventCommon("protocol/testdata/" + eventYaml)
	if eventCommon == nil {
		t.Errorf("Parse %v Failed", eventYaml)
		return
	}

	for _, yaml := range traceYamls {
		trace := getTrace("protocol/testdata/" + yaml)
		if trace == nil {
			t.Errorf("Parse %v Failed", yaml)
			return
		}

		t.Run(trace.Key, func(t *testing.T) {
			mps := trace.PrepareMessagePairs(eventCommon)
			result := na.parseProtocols(mps)
			trace.Validate(t, result)
		})
	}
}

func getEventCommon(path string) *EventCommon {
	eventCommon := &EventCommon{}
	viper := viper.New()
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error%v\n", err)
		return nil
	}
	viper.UnmarshalKey("eventCommon", eventCommon)
	return eventCommon
}

func getTrace(path string) *Trace {
	trace := &Trace{}
	viper := viper.New()
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error%v\n", err)
		return nil
	}
	viper.UnmarshalKey("trace", trace)
	return trace
}

type EventCommon struct {
	Source   int `mapstructure:"source"`
	Category int `mapstructure:"category"`
	Ctx      Ctx `mapstructure:"ctx"`
}

type Ctx struct {
	Thread ThreadInfo `mapstructure:"thread_info"`
	Fd     FdInfo     `mapstructure:"fd_info"`
}

type ThreadInfo struct {
	Pid         uint32 `mapstructure:"pid"`
	Tid         uint32 `mapstructure:"tid"`
	Uid         uint32 `mapstructure:"uid"`
	Gid         uint32 `mapstructure:"gid"`
	Comm        string `mapstructure:"comm"`
	ContainerId string `mapstructure:"container_id"`
}

type FdInfo struct {
	Num      int32    `mapstructure:"num"`
	TypeFd   int32    `mapstructure:"type_fd"`
	Protocol uint32   `mapstructure:"protocol"`
	Role     bool     `mapstructure:"role"`
	Sip      []uint32 `mapstructure:"sip"`
	Dip      []uint32 `mapstructure:"dip"`
	Sport    uint32   `mapstructure:"sport"`
	Dport    uint32   `mapstructure:"dport"`
}

type Trace struct {
	Key       string        `mapstructure:"key"`
	Connects  []TraceEvent  `mapstructure:"connects"`
	Requests  []TraceEvent  `mapstructure:"requests"`
	Responses []TraceEvent  `mapstructure:"responses"`
	Expects   []TraceExpect `mapstructure:"expects"`
}

func (trace *Trace) PrepareMessagePairs(common *EventCommon) *messagePairs {
	mps := &messagePairs{
		connects:  nil,
		requests:  nil,
		responses: nil,
		mutex:     sync.RWMutex{},
	}
	if trace.Connects != nil {
		for _, connect := range trace.Connects {
			mps.mergeConnect(connect.exchange(common))
		}
	}
	if trace.Requests != nil {
		for _, request := range trace.Requests {
			mps.mergeRequest(request.exchange(common))
		}
	}
	if trace.Responses != nil {
		for _, response := range trace.Responses {
			mps.mergeResponse(response.exchange(common))
		}
	}
	return mps
}

func (trace *Trace) Validate(t *testing.T, results []*model.GaugeGroup) {
	checkSize(t, "Expect Size", len(trace.Expects), len(results))

	for i, result := range results {
		expect := trace.Expects[i]
		checkUint64Equal(t, "Timestamp", expect.Timestamp, result.Timestamp)

		// Validate Gauge Values
		checkSize(t, "Values Size", len(expect.Values), len(result.Values))
		for _, value := range result.Values {
			expectValue, ok := expect.Values[value.Name]
			if !ok {
				t.Errorf("[Miss %s] want=nil, got=%d", value.Name, value.Value)
			} else {
				checkInt64Equal(t, value.Name, expectValue, value.Value)
			}
		}

		// Validate Gauge Attributes
		checkSize(t, "Labels Size", len(expect.Labels), result.Labels.Size())
		for labelKey, labelValue := range expect.Labels {
			if reflect.TypeOf(labelValue).Name() == "int" {
				gotValue := result.Labels.GetIntValue(labelKey)
				checkInt64Equal(t, labelKey, int64(labelValue.(int)), gotValue)
			} else if reflect.TypeOf(labelValue).Name() == "bool" {
				gotValue := result.Labels.GetBoolValue(labelKey)
				checkBoolEqual(t, labelKey, labelValue.(bool), gotValue)
			} else {
				gotValue := result.Labels.GetStringValue(labelKey)
				checkStringEqual(t, labelKey, labelValue.(string), gotValue)
			}
		}

		for labelKey, labelValue := range result.Labels.ToStringMap() {
			if _, ok := expect.Labels[labelKey]; !ok {
				t.Errorf("[Miss %s] want=nil, got=%s", labelKey, labelValue)
			}
		}
	}
}

func checkBoolEqual(t *testing.T, key string, expect bool, got bool) {
	if expect != got {
		t.Errorf("[Check %s] want=%t, got=%t", key, expect, got)
	}
}

func checkStringEqual(t *testing.T, key string, expect string, got string) {
	if expect != got {
		t.Errorf("[Check %s] want=%s, got=%s", key, expect, got)
	}
}

func checkUint64Equal(t *testing.T, key string, expect uint64, got uint64) {
	if expect != got {
		t.Errorf("[Check %s] want=%d, got=%d", key, expect, got)
	}
}

func checkInt64Equal(t *testing.T, key string, expect int64, got int64) {
	if expect != got {
		t.Errorf("[Check %s] want=%d, got=%d", key, expect, got)
	}
}

func checkSize(t *testing.T, key string, expect int, got int) {
	if expect != got {
		t.Errorf("[Check %s] want=%d, got=%d", key, expect, got)
	}
}

type TraceEvent struct {
	Name           string         `mapstructure:"name"`
	Timestamp      uint64         `mapstructure:"timestamp"`
	UserAttributes UserAttributes `mapstructure:"user_attributes"`
}

func (evt *TraceEvent) exchange(common *EventCommon) *model.KindlingEvent {
	var byteData = getData(evt.UserAttributes.Data)

	modelEvt := &model.KindlingEvent{
		Source:           model.Source(common.Source),
		Timestamp:        evt.Timestamp,
		Name:             evt.Name,
		Category:         model.Category(common.Category),
		NativeAttributes: &model.Property{},
		UserAttributes: []*model.KeyValue{
			{Key: "latency", Value: &model.AnyValue{Value: &model.AnyValue_IntValue{IntValue: evt.UserAttributes.Latency}}},
			{Key: "res", Value: &model.AnyValue{Value: &model.AnyValue_IntValue{IntValue: evt.UserAttributes.Res}}},
			{Key: "data", Value: &model.AnyValue{Value: &model.AnyValue_BytesValue{BytesValue: byteData}}},
		},
		Ctx: &model.Context{
			ThreadInfo: &model.Thread{
				Pid:  common.Ctx.Thread.Pid,
				Tid:  common.Ctx.Thread.Tid,
				Uid:  common.Ctx.Thread.Uid,
				Gid:  common.Ctx.Thread.Gid,
				Comm: common.Ctx.Thread.Comm,
			},
			FdInfo: &model.Fd{
				Num:      common.Ctx.Fd.Num,
				TypeFd:   model.FDType(common.Ctx.Fd.TypeFd),
				Protocol: model.L4Proto(common.Ctx.Fd.Protocol),
				Role:     common.Ctx.Fd.Role,
				Sip:      common.Ctx.Fd.Sip,
				Dip:      common.Ctx.Fd.Dip,
				Sport:    common.Ctx.Fd.Sport,
				Dport:    common.Ctx.Fd.Dport,
			},
		},
	}
	return modelEvt
}

/**
* Convert Following format to byte array.
* size + | + ASCII string
* Hex string
* ASCII string
 */
func getData(datas []string) []byte {
	dataBytes := make([]byte, 0)

	for _, data := range datas {
		if len(data) > 0 {
			dataSplit := getSplit(data)
			if dataSplit > 0 {
				sizeArray, _ := hex.DecodeString(data[0:dataSplit])
				dataBytes = append(dataBytes, sizeArray...)

				dataLen := len(data)
				for i := dataSplit + 1; i < dataLen; i++ {
					dataBytes = append(dataBytes, data[i])
				}
			} else if isHex(data[0]) {
				hexArray, _ := hex.DecodeString(data)
				dataBytes = append(dataBytes, hexArray...)
			} else {
				byteArray := []byte(data)
				dataBytes = append(dataBytes, byteArray...)
			}
		}
	}
	return dataBytes
}

func getSplit(data string) int {
	if len(data) >= 3 && data[2] == '|' {
		return 2
	}
	if len(data) >= 5 && data[4] == '|' {
		return 4
	}
	return 0
}

func isHex(b byte) bool {
	if b >= '0' && b <= '9' {
		return true
	}
	if b >= 'a' && b <= 'z' {
		return true
	}
	return false
}

type UserAttributes struct {
	Latency int64    `mapstructure:"latency"`
	Res     int64    `mapstructure:"res"`
	Data    []string `mapstructure:"data"`
}

type TraceExpect struct {
	Timestamp uint64                 `mapstructure:"Timestamp"`
	Values    map[string]int64       `mapstructure:"Values"`
	Labels    map[string]interface{} `mapstructure:"Labels"`
}
