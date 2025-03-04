package jvmmetricprocessor

import (
	"fmt"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type metricsTransformProcessor struct {
	logger *zap.Logger
}

func (p *metricsTransformProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	p.logger.Info("processMetrics() start ......")
	p.logger.Info("received metrics: {}")
	metrics := md.ResourceMetrics()
	i := metrics.Len()
	p.logger.Info("metrics len :", zap.Int("len", i))
	PrintMetrics(md)
	return md, nil
}

func newMetricsTransformProcessor(logger *zap.Logger) *metricsTransformProcessor {
	return &metricsTransformProcessor{
		logger: logger,
	}
}

func PrintMetrics(metrics pmetric.Metrics) {
	fmt.Println("=== Metrics Data ===")
	fmt.Printf("Metric Count: %d\n", metrics.MetricCount())
	fmt.Printf("Data Point Count: %d\n", metrics.DataPointCount())

	rms := metrics.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		fmt.Printf("ResourceMetrics[%d]:\n", i)

		ilms := rm.ScopeMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			fmt.Printf("  ScopeMetrics[%d]:\n", j)

			ms := ilm.Metrics()
			for k := 0; k < ms.Len(); k++ {
				m := ms.At(k)
				fmt.Printf("    Metric[%d]: Name=%s, Type=%v\n", k, m.Name(), m.Type())

				// 遍历数据点
				switch m.Type() {
				case pmetric.MetricTypeGauge:
					dataPoints := m.Gauge().DataPoints()
					printDataPoints(dataPoints)
				case pmetric.MetricTypeSum:
					dataPoints := m.Sum().DataPoints()
					printDataPoints(dataPoints)
				case pmetric.MetricTypeHistogram:
					dataPoints := m.Histogram().DataPoints()
					printDataPoints(dataPoints)
				case pmetric.MetricTypeExponentialHistogram:
					dataPoints := m.ExponentialHistogram().DataPoints()
					printDataPoints(dataPoints)
				case pmetric.MetricTypeSummary:
					dataPoints := m.Summary().DataPoints()
					printDataPoints(dataPoints)
				}
			}
		}
	}
	fmt.Println("====================")
}

func printDataPoints[T any](dataPoints T) {
	fmt.Printf("DataPoints[%d]:\n", dataPoints)
}
