package cloudeventexporter

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr   = "cloudeventexporter"
	stability = component.StabilityLevelAlpha
)

var exporterCapabilities = consumer.Capabilities{MutatesData: false}
var componentType = component.MustNewType("cloudeventexporter")

// NewFactory creates a factory for the routing exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		componentType,
		CreateDefaultConfig,
		exporter.WithLogs(createLogsExporter, stability),
	)
}

func CreateDefaultConfig() component.Config {
	return &Config{
		Ce: CloudEventSpec{
			SpecVersion: "1.0",
		},
	}
}

func createLogsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {

	eCfg, ok := cfg.(*Config)
	if !ok {
		return nil, errors.New("could not initialize cloud-event transform exporter")
	}

	ceExporter, err := newExporter(cfg, set)
	if err != nil {
		return nil, errors.New("Failed to create cloud-event exporter")
	}

	return exporterhelper.NewLogs(ctx, set, eCfg, ceExporter.pushLogs,
		exporterhelper.WithStart(ceExporter.start),
		exporterhelper.WithShutdown(ceExporter.shutdown),
		exporterhelper.WithCapabilities(exporterCapabilities),
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}),
		exporterhelper.WithRetry(eCfg.RetrySettings),
		exporterhelper.WithQueue(eCfg.QueueSettings),
	)
}
