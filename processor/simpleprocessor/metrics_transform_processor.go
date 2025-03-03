// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package simpleprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"

import (
	"context"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type metricsTransformProcessor struct {
	logger                   *zap.Logger
	otlpDataModelGateEnabled bool
}

func (p metricsTransformProcessor) processMetrics(ctx context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	p.logger.Debug("processing metrics", zap.Any("metrics", metrics))
	panic("processing metrics")
	return metrics, nil
}

func newMetricsTransformProcessor(logger *zap.Logger) *metricsTransformProcessor {
	return &metricsTransformProcessor{
		logger: logger,
	}
}
