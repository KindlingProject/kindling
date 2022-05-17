package prometheusexporter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator/defaultaggregator"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"time"
)

type collector struct {
	aggregator *defaultaggregator.DefaultAggregator
}

func (c *collector) Collect(metrics chan<- prometheus.Metric) {
	gaugeGroups := c.aggregator.Dump()

	for i := 0; i < len(gaugeGroups); i++ {
		gaugeGroup := gaugeGroups[i]
		labelMap := gaugeGroup.Labels.GetValues()
		ts := getTimestamp(gaugeGroup.Timestamp)
		keys := make([]string, len(labelMap))
		values := make([]string, len(labelMap))
		for k, v := range labelMap {
			keys = append(keys, k)
			values = append(values, v.ToString())
		}
		for s := 0; s < len(gaugeGroup.Values); s++ {
			gauge := gaugeGroup.Values[s]
			switch gauge.DataType() {
			case model.IntGaugeType:
				metric, error := prometheus.NewConstMetric(prometheus.NewDesc(
					gauge.Name,
					"",
					keys,
					nil,
				), prometheus.CounterValue, float64(gauge.GetInt().Value), values...)
				tm := prometheus.NewMetricWithTimestamp(ts, metric)
				if error != nil {
					metrics <- tm
				}
			case model.HistogramGaugeType:
				histogram := gauge.GetHistogram()
				buckets := make(map[float64]uint64, len(histogram.ExplicitBoundaries))
				for x := 0; x < len(histogram.ExplicitBoundaries); x++ {
					bound := histogram.ExplicitBoundaries[x]
					buckets[float64(bound)] = histogram.BucketCounts[x]
				}
				metric, error := prometheus.NewConstHistogram(prometheus.NewDesc(
					gauge.Name,
					"",
					keys,
					nil,
				), histogram.Count, float64(histogram.Sum), buckets, values...)
				tm := prometheus.NewMetricWithTimestamp(ts, metric)
				if error != nil {
					metrics <- tm
				}
			}
		}
	}
}

func (c *collector) recordGaugeGroups(group *model.GaugeGroup) {
	c.aggregator.Aggregate(group, getSelector(group.Name))
}

func newCollector(config *Config, logger *zap.Logger) *collector {
	return &collector{aggregator: defaultaggregator.NewDefaultAggregator(&defaultaggregator.AggregatedConfig{
		//TODO create Aggregator...
	})}
}

func getTimestamp(ts uint64) time.Time {
	return time.UnixMilli(int64(ts))
}

func (c *collector) Describe(_ chan<- *prometheus.Desc) {}

// TODO
func getSelector(metricName string) *aggregator.LabelSelectors {
	switch metricName {
	//case constnames.TraceAsMetric:
	//	return i.traceAsMetricSelector
	//case constnames.TcpRttMetricName:
	//	return i.TcpRttMillsSelector
	default:
		return nil
	}
}
