package network

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"sync"
	"time"
)

func createGaugeGroup() interface{} {
	values := []*model.Gauge{
		model.NewIntGauge(constvalues.ConnectTime, 0),
		model.NewIntGauge(constvalues.RequestSentTime, 0),
		model.NewIntGauge(constvalues.WaitingTtfbTime, 0),
		model.NewIntGauge(constvalues.ContentDownloadTime, 0),
		model.NewIntGauge(constvalues.RequestTotalTime, 0),
		model.NewIntGauge(constvalues.RequestIo, 0),
		model.NewIntGauge(constvalues.ResponseIo, 0),
	}
	gaugeGroup := model.NewGaugeGroup(constnames.NetRequestGaugeGroupName, model.NewAttributeMap(), uint64(time.Now().UnixNano()), values...)
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
	gaugeGroup.Name = constnames.NetRequestGaugeGroupName
	p.pool.Put(gaugeGroup)
}
