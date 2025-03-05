package simpleexporter

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/xexporter"
	"go.uber.org/zap"
)

var (
	Type   = component.MustNewType("simple")
	logger = zap.NewNop()
)

const (
	ProfilesStability = component.StabilityLevelDevelopment
	TracesStability   = component.StabilityLevelStable
	MetricsStability  = component.StabilityLevelStable
	LogsStability     = component.StabilityLevelStable
)

func NewFactory() exporter.Factory {
	logger.Info("jvmhttp exporter created")
	return xexporter.NewFactory(
		Type,
		createDefaultConfig,
		xexporter.WithMetrics(createMetrics, MetricsStability),
	)
}

func createMetrics(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Metrics, error) {
	logger.Info("jvmhttp exporter started")
	oce, err := newExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewMetrics(ctx, set, cfg,
		oce.pushMetrics,
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		// explicitly disable since we rely on http.Client timeout logic.
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}),
	)

}
