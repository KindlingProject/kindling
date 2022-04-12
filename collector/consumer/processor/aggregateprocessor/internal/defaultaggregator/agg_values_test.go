package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/model"
	"reflect"
	"sync"
	"testing"
)

func startTask(workerNum int, task func(wg *sync.WaitGroup)) {
	wg := sync.WaitGroup{}
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go task(&wg)
	}
	wg.Wait()
}

func assertTest(t *testing.T, aggValues aggregatedValues, workerNum int, expectedNum int64,
	task func(wg *sync.WaitGroup)) {
	startTask(workerNum, task)
	got := aggValues.get()
	if expectedNum != aggValues.get() {
		t.Errorf("The result is expected to be %d, but got %d", expectedNum, got)
	}
}

func Test_aggValues_sum(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	expectedNum := int64(workerNum * loopNum)
	aggValues := &sumValue{}
	assertTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := 0; j < loopNum; j++ {
			aggValues.calculate(1)
		}
		wg.Done()
	})
}

func Test_aggValues_max(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	expectedNum := int64(loopNum - 1)
	aggValues := &maxValue{}
	assertTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := 0; j < loopNum; j++ {
			aggValues.calculate(int64(j))
		}
		wg.Done()
	})
}

func Test_aggValues_avg(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	var expectedNum int64 = 1
	aggValues := &avgValue{}
	assertTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := 0; j < loopNum; j++ {
			aggValues.calculate(1)
		}
		wg.Done()
	})
}

func Test_aggValues_last(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	var expectedNum int64 = 1
	aggValues := &lastValue{}
	assertTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := loopNum; j > 0; j-- {
			aggValues.calculate(int64(j))
		}
		wg.Done()
	})
}

func Test_defaultValuesMap_sum(t *testing.T) {
	kindMap := make(map[string][]aggregatorKind)
	kindMap["sum_value"] = []aggregatorKind{sumKind}
	gauges := []*model.Gauge{{Name: "sum_value"}}
	m := newAggValuesMap(gauges, kindMap)
	for i := 0; i < 10000; i++ {
		m.calculate("sum_value", 1)
	}
	got := m.get("sum_value")
	if got[0].Name != "sum_value_sum" || got[0].Value != 10000 {
		t.Errorf("sum result is %v, expected %v", got, 10000)
	}
}

func Test_defaultValuesMap_avg(t *testing.T) {
	kindMap := make(map[string][]aggregatorKind)
	kindMap["avg_value"] = []aggregatorKind{avgKind}
	gauges := []*model.Gauge{{Name: "avg_value"}}
	m := newAggValuesMap(gauges, kindMap)
	for i := 0; i < 10000; i++ {
		m.calculate("avg_value", 1)
	}
	got := m.get("avg_value")
	if got[0].Name != "avg_value_avg" || got[0].Value != 1 {
		t.Errorf("avg result is %v, expected %v", got, 1)
	}
}

func Test_defaultValuesMap_max(t *testing.T) {
	kindMap := make(map[string][]aggregatorKind)
	kindMap["max_value"] = []aggregatorKind{maxKind}
	gauges := []*model.Gauge{{Name: "max_value"}}
	m := newAggValuesMap(gauges, kindMap)
	for i := 0; i < 10000; i++ {
		m.calculate("max_value", int64(i))
	}
	got := m.get("max_value")
	var expected int64 = 9999
	if got[0].Name != "max_value_max" || got[0].Value != expected {
		t.Errorf("max result is %v, expected %v", got, expected)
	}

	kindMap["reserve_max_value"] = []aggregatorKind{maxKind}
	gauges = []*model.Gauge{{Name: "reserve_max_value"}}
	m = newAggValuesMap(gauges, kindMap)
	for i := 10000; i > 0; i-- {
		m.calculate("reserve_max_value", int64(i))
	}
	got = m.get("reserve_max_value")
	if got[0].Name != "reserve_max_value_max" || got[0].Value != 10000 {
		t.Errorf("max result is %v, expected %v", got, 10000)
	}
}

func Test_defaultValuesMap_lastValue(t *testing.T) {
	kindMap := make(map[string][]aggregatorKind)
	kindMap["last_value"] = []aggregatorKind{lastKind}
	gauges := []*model.Gauge{{Name: "last_value"}}
	m := newAggValuesMap(gauges, kindMap)
	for i := 10000; i > 0; i-- {
		m.calculate("last_value", 1)
	}
	got := m.get("last_value")
	if got[0].Name != "last_value_last" || got[0].Value != 1 {
		t.Errorf("lastValue result is %v, expected %v", got, 1)
	}
}

func Test_toAggKindMap(t *testing.T) {
	type args struct {
		input map[string][]string
	}
	tests := []struct {
		name string
		args args
		want map[string][]aggregatorKind
	}{
		{
			name: "one kind",
			args: args{input: map[string][]string{"metric1": {"sum"}}},
			want: map[string][]aggregatorKind{"metric1": {sumKind}},
		},
		{
			name: "multiple kinds",
			args: args{input: map[string][]string{"metric1": {"sum", "avg", "last", "max"}}},
			want: map[string][]aggregatorKind{"metric1": {sumKind, avgKind, lastKind, maxKind}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toAggKindMap(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toAggKindMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
