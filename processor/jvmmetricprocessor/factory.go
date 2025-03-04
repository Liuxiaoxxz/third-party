package jvmmetricprocessor

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"golang.org/x/net/context"
)

var (
	Type                 = component.MustNewType("jvmmetricr")
	MetricsStability     = component.StabilityLevelBeta
	consumerCapabilities = consumer.Capabilities{MutatesData: true}
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		Type,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, MetricsStability))
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics) (processor.Metrics, error) {
	logger := set.Logger
	logger.Info("JVM_Metrics_Transform_Processor run ......")
	oCfg := cfg.(*Config)
	if err := validateConfiguration(oCfg); err != nil {
		return nil, err
	}
	logger.Info("validateConfiguration done ......")
	logger.Info("build start ......")
	metricsProcessor := newMetricsTransformProcessor(set.Logger)
	return processorhelper.NewMetrics(
		ctx,
		set,
		cfg,
		nextConsumer,
		metricsProcessor.processMetrics,
		processorhelper.WithCapabilities(consumerCapabilities))
}

func hello(s string) string {
	return "hello " + s
}
