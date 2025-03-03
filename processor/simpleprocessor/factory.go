// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package simpleprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"

import (
	"context"
	"github.com/Liuxiaoxxz/third-party/processor/simpleprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

var consumerCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Metrics transform processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability))
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	metricsProcessor := newMetricsTransformProcessor(set.Logger)

	return processorhelper.NewMetrics(
		ctx,
		set,
		cfg,
		nextConsumer,
		metricsProcessor.processMetrics,
		processorhelper.WithCapabilities(consumerCapabilities))
}
