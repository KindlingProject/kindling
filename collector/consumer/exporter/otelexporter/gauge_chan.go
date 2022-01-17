package otelexporter

import (
	"github.com/dxsup/kindling-collector/model"
	"go.uber.org/zap"
	"sync"
)

type gaugeChannel struct {
	chanMap            map[string]chan *model.GaugeGroup
	mutex              sync.RWMutex
	defaultChannelSize int
}

func newGaugeChannel(size int) *gaugeChannel {
	ret := &gaugeChannel{
		chanMap:            make(map[string]chan *model.GaugeGroup),
		defaultChannelSize: size,
	}
	return ret
}

func (g *gaugeChannel) put(gaugeGroup *model.GaugeGroup, logger *zap.Logger) {
	values := gaugeGroup.Values
	for _, value := range values {
		name := value.Name
		g.mutex.RLock()
		channel, ok := g.chanMap[name]
		g.mutex.RUnlock()
		if !ok {
			channel = g.newChannel(name)
		}
		select {
		case channel <- gaugeGroup:
			continue
		default:
			logger.Warn("gaugeGroup cannot be added to channel, skip", zap.String("valueName", value.Name))
		}
	}
}

func (g *gaugeChannel) getChannel(gaugeName string) chan *model.GaugeGroup {
	var restChan chan *model.GaugeGroup
	g.mutex.RLock()
	channel, ok := g.chanMap[gaugeName]
	g.mutex.RUnlock()
	if ok {
		restChan = channel
		// New a channel to replace the original one
		g.newChannel(gaugeName)
		return restChan
	} else {
		// New a channel if no previous one was found
		return g.newChannel(gaugeName)
	}
}

func (g *gaugeChannel) newChannel(gaugeName string) chan *model.GaugeGroup {
	channel := make(chan *model.GaugeGroup, g.defaultChannelSize)
	g.mutex.Lock()
	g.chanMap[gaugeName] = channel
	g.mutex.Unlock()
	return channel
}
