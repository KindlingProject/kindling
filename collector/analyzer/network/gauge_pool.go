package network

import (
	"sync"
	"time"

	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
)

func createDataGroup() interface{} {
	values := []*model.Metric{
		model.NewIntMetric(constvalues.ConnectTime, 0),
		model.NewIntMetric(constvalues.RequestSentTime, 0),
		model.NewIntMetric(constvalues.WaitingTtfbTime, 0),
		model.NewIntMetric(constvalues.ContentDownloadTime, 0),
		model.NewIntMetric(constvalues.RequestTotalTime, 0),
		model.NewIntMetric(constvalues.RequestIo, 0),
		model.NewIntMetric(constvalues.ResponseIo, 0),
	}
	dataGroup := model.NewDataGroup(constnames.NetRequestMetricGroupName, model.NewAttributeMap(), uint64(time.Now().UnixNano()), values...)
	return dataGroup
}

type DataGroupPool struct {
	pool *sync.Pool
}

func NewDataGroupPool() *DataGroupPool {
	return &DataGroupPool{pool: &sync.Pool{New: createDataGroup}}
}

func (p *DataGroupPool) Get() *model.DataGroup {
	return p.pool.Get().(*model.DataGroup)
}

func (p *DataGroupPool) Free(dataGroup *model.DataGroup) {
	dataGroup.Reset()
	dataGroup.Name = constnames.NetRequestMetricGroupName
	p.pool.Put(dataGroup)
}
