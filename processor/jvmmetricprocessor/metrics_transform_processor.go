package jvmmetricprocessor

import (
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
	return md, nil
}

func newMetricsTransformProcessor(logger *zap.Logger) *metricsTransformProcessor {
	return &metricsTransformProcessor{
		logger: logger,
	}
}
