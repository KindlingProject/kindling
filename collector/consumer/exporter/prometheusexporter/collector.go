package prometheusexporter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator/defaultaggregator"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"time"
)

type collector struct {
	aggregator *defaultaggregator.CumulativeAggregator
}

func (c *collector) Collect(metrics chan<- prometheus.Metric) {
	// TODO debug
	metricGroups := c.aggregator.DumpAndRemoveExpired(time.Now())
	for i := 0; i < len(metricGroups); i++ {
		metricGroup := metricGroups[i]
		labelMap := metricGroup.Labels.GetValues()
		ts := getTimestamp(metricGroup.Timestamp)
		keys := make([]string, 0, len(labelMap))
		values := make([]string, 0, len(labelMap))
		for k, v := range labelMap {
			keys = append(keys, k)
			values = append(values, v.ToString())
		}
		for s := 0; s < len(metricGroup.Metrics); s++ {
			metric := metricGroup.Metrics[s]
			switch metric.DataType() {
			case model.IntMetricType:
				metric, error := prometheus.NewConstMetric(prometheus.NewDesc(
					sanitize(metric.Name, true),
					"",
					keys,
					nil,
					// TODO not all IntMetric are Counter, they can also be a Metric
				), prometheus.CounterValue, float64(metric.GetInt().Value), values...)
				if error == nil {
					tm := prometheus.NewMetricWithTimestamp(ts, metric)
					metrics <- tm
				}
			case model.HistogramMetricType:
				histogram := metric.GetHistogram()
				buckets := make(map[float64]uint64, len(histogram.ExplicitBoundaries))
				for x := 0; x < len(histogram.ExplicitBoundaries); x++ {
					bound := histogram.ExplicitBoundaries[x]
					buckets[float64(bound)] = histogram.BucketCounts[x]
				}
				metric, error := prometheus.NewConstHistogram(prometheus.NewDesc(
					sanitize(metric.Name, true),
					"",
					keys,
					nil,
				), histogram.Count, float64(histogram.Sum), buckets, values...)
				if error == nil {
					tm := prometheus.NewMetricWithTimestamp(ts, metric)
					metrics <- tm
				}
			}
		}
	}
}

func (c *collector) recordMetricGroups(group *model.DataGroup) {
	c.aggregator.AggregatorWithAllLabelsAndMetric(group, time.Now())
}

func newCollector(config *Config, logger *zap.Logger) *collector {
	// TODO Do this in config later !!!!
	requestTotalTimeTopologyMetric := constnames.ToKindlingNetMetricName(constvalues.RequestTotalTime, false)
	requestTotalTimeEntityMetric := constnames.ToKindlingNetMetricName(constvalues.RequestTotalTime, true)
	return &collector{
		aggregator: defaultaggregator.NewCumulativeAggregator(
			&defaultaggregator.AggregatedConfig{
				KindMap: map[string][]defaultaggregator.KindConfig{
					constnames.TcpRttMetricName: {{Kind: defaultaggregator.LastKind, OutputName: constnames.TcpRttMetricName}},
					constnames.TraceAsMetric:    {{Kind: defaultaggregator.LastKind, OutputName: constnames.TraceAsMetric}},
					requestTotalTimeTopologyMetric: {{
						Kind:               defaultaggregator.HistogramKind,
						OutputName:         requestTotalTimeTopologyMetric,
						ExplicitBoundaries: []int64{1e6, 2e6, 5e6, 1e7, 2e7, 5e7, 1e8, 2e8, 5e8, 1e9, 2e9, 5e9},
					}},
					requestTotalTimeEntityMetric: {{
						Kind:               defaultaggregator.HistogramKind,
						OutputName:         requestTotalTimeEntityMetric,
						ExplicitBoundaries: []int64{1e6, 2e6, 5e6, 1e7, 2e7, 5e7, 1e8, 2e8, 5e8, 1e9, 2e9, 5e9},
					}},
				},
			}, time.Minute*5)}
}

func getTimestamp(ts uint64) time.Time {
	return time.UnixMicro(int64(ts / 1000))
}

// Describe is a no-op, because the collector dynamically allocates metrics.
// https://github.com/prometheus/client_golang/blob/v1.9.0/prometheus/collector.go#L28-L40
func (c *collector) Describe(_ chan<- *prometheus.Desc) {}
