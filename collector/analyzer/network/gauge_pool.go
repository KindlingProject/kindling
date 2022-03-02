package network

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"sync"
	"time"
)

func createGaugeGroup() interface{} {
	values := []*model.Gauge{
		{Name: constvalues.ConnectTime, Value: 0},
		{Name: constvalues.RequestSentTime, Value: 0},
		{Name: constvalues.WaitingTtfbTime, Value: 0},
		{Name: constvalues.ContentDownloadTime, Value: 0},
		{Name: constvalues.RequestTotalTime, Value: 0},
		{Name: constvalues.RequestIo, Value: 0},
		{Name: constvalues.ResponseIo, Value: 0},
	}
	gaugeGroup := model.NewGaugeGroup("", model.NewAttributeMap(), uint64(time.Now().UnixNano()), values...)
	return gaugeGroup
}

type GaugeGroupPool struct {
	pool *sync.Pool
}

func NewGaugePool() *GaugeGroupPool {
	return &GaugeGroupPool{pool: &sync.Pool{New: createGaugeGroup}}
}

func (p *GaugeGroupPool) Get() *model.GaugeGroup {
	return p.pool.Get().(*model.GaugeGroup)
}

func (p *GaugeGroupPool) Free(gaugeGroup *model.GaugeGroup) {
	gaugeGroup.Reset()
	p.pool.Put(gaugeGroup)
}
