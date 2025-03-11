// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package jvmxexporter // import "go.opentelemetry.io/collector/exporter/otlpexporter"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterbatcher"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper"
	"go.opentelemetry.io/collector/exporter/xexporter"
)

var (
	Type      = component.MustNewType("jvm")
	ScopeName = "go.opentelemetry.io/collector/exporter/otlpexporter"
)

const (
	ProfilesStability = component.StabilityLevelDevelopment
	TracesStability   = component.StabilityLevelStable
	MetricsStability  = component.StabilityLevelStable
	LogsStability     = component.StabilityLevelStable
)

// NewFactory creates a factory for OTLP exporter.
func NewFactory() exporter.Factory {
	return xexporter.NewFactory(
		Type,
		createDefaultConfig,
		//xexporter.WithTraces(createTraces, TracesStability),
		xexporter.WithMetrics(createMetrics, MetricsStability),
		//xexporter.WithLogs(createLogs, LogsStability),
		//xexporter.WithProfiles(createProfilesExporter, ProfilesStability),
	)
}

func createDefaultConfig() component.Config {
	batcherCfg := exporterbatcher.NewDefaultConfig()
	batcherCfg.Enabled = false

	clientCfg := *configgrpc.NewDefaultClientConfig()
	// Default to gzip compression
	clientCfg.Compression = configcompression.TypeGzip
	// We almost read 0 bytes, so no need to tune ReadBufferSize.
	clientCfg.WriteBufferSize = 512 * 1024
	// For backward compatibility:
	clientCfg.Keepalive = nil
	clientCfg.BalancerName = ""

	return &Config{
		TimeoutConfig: exporterhelper.NewDefaultTimeoutConfig(),
		RetryConfig:   configretry.NewDefaultBackOffConfig(),
		QueueConfig:   exporterhelper.NewDefaultQueueConfig(),
		BatcherConfig: batcherCfg,
		ClientConfig:  clientCfg,
	}
}

func createTraces(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	oce := newExporter(cfg, set)
	oCfg := cfg.(*Config)
	return exporterhelper.NewTraces(ctx, set, cfg,
		oce.pushTraces,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(oCfg.TimeoutConfig),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig),
		exporterhelper.WithBatcher(oCfg.BatcherConfig),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown),
	)
}

func createMetrics(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	oce := newExporter(cfg, set)
	oCfg := cfg.(*Config)
	return exporterhelper.NewMetrics(ctx, set, cfg,
		oce.pushMetrics,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(oCfg.TimeoutConfig),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig),
		exporterhelper.WithBatcher(oCfg.BatcherConfig),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown),
	)
}

func createLogs(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	oce := newExporter(cfg, set)
	oCfg := cfg.(*Config)
	return exporterhelper.NewLogs(ctx, set, cfg,
		oce.pushLogs,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(oCfg.TimeoutConfig),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig),
		exporterhelper.WithBatcher(oCfg.BatcherConfig),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown),
	)
}

func createProfilesExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (xexporter.Profiles, error) {
	oce := newExporter(cfg, set)
	oCfg := cfg.(*Config)
	return xexporterhelper.NewProfilesExporter(ctx, set, cfg,
		oce.pushProfiles,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(oCfg.TimeoutConfig),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig),
		exporterhelper.WithBatcher(oCfg.BatcherConfig),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown),
	)
}
