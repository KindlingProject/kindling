package network

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"sync"
	"time"
)

func createMetricGroup() interface{} {
	values := []*model.Metric{
		model.NewIntMetric(constvalues.ConnectTime, 0),
		model.NewIntMetric(constvalues.RequestSentTime, 0),
		model.NewIntMetric(constvalues.WaitingTtfbTime, 0),
		model.NewIntMetric(constvalues.ContentDownloadTime, 0),
		model.NewIntMetric(constvalues.RequestTotalTime, 0),
		model.NewIntMetric(constvalues.RequestIo, 0),
		model.NewIntMetric(constvalues.ResponseIo, 0),
	}
	metricGroup := model.NewDataGroup(constnames.NetRequestMetricGroupName, model.NewAttributeMap(), uint64(time.Now().UnixNano()), values...)
	return metricGroup
}

type MetricGroupPool struct {
	pool *sync.Pool
}

func NewMetricPool() *MetricGroupPool {
	return &MetricGroupPool{pool: &sync.Pool{New: createMetricGroup}}
}

func (p *MetricGroupPool) Get() *model.DataGroup {
	return p.pool.Get().(*model.DataGroup)
}

func (p *MetricGroupPool) Free(metricGroup *model.DataGroup) {
	metricGroup.Reset()
	metricGroup.Name = constnames.NetRequestMetricGroupName
	p.pool.Put(metricGroup)
}
