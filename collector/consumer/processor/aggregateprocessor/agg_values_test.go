package aggregateprocessor

import "testing"

func Test_defaultValuesMap_sum(t *testing.T) {
	kindMap := make(map[string]aggregatorKind)
	kindMap["sum_value"] = sumKind
	m := newAggValuesMap(kindMap)
	for i := 0; i < 10000; i++ {
		m.calculate("sum_value", 1)
	}
	got := m.get("sum_value")
	if got != 10000 {
		t.Errorf("sum result is %v, expected %v", got, 10000)
	}
}

func Test_defaultValuesMap_avg(t *testing.T) {
	kindMap := make(map[string]aggregatorKind)
	kindMap["avg_value"] = avgKind
	m := newAggValuesMap(kindMap)
	for i := 0; i < 10000; i++ {
		m.calculate("avg_value", 1)
	}
	got := m.get("avg_value")
	if got != 1 {
		t.Errorf("avg result is %v, expected %v", got, 1)
	}
}

func Test_defaultValuesMap_max(t *testing.T) {
	kindMap := make(map[string]aggregatorKind)
	kindMap["max_value"] = maxKind
	m := newAggValuesMap(kindMap)
	for i := 0; i < 10000; i++ {
		m.calculate("max_value", int64(i))
	}
	got := m.get("max_value")
	var expected int64 = 9999
	if got != expected {
		t.Errorf("max result is %v, expected %v", got, expected)
	}

	kindMap["reserve_max_value"] = maxKind
	m = newAggValuesMap(kindMap)
	for i := 10000; i > 0; i-- {
		m.calculate("max_value", int64(i))
	}
	got = m.get("max_value")
	if got != 10000 {
		t.Errorf("max result is %v, expected %v", got, 10000)
	}
}

func Test_defaultValuesMap_lastValue(t *testing.T) {
	kindMap := make(map[string]aggregatorKind)
	kindMap["last_value"] = lastKind
	m := newAggValuesMap(kindMap)
	for i := 10000; i > 0; i-- {
		m.calculate("last_value", 1)
	}
	got := m.get("last_value")
	if got != 1 {
		t.Errorf("lastValue result is %v, expected %v", got, 1)
	}
}
