package jvmhttpexporter

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper"
	"go.opentelemetry.io/collector/exporter/xexporter"
)

var (
	Type      = component.MustNewType("jvmhttp")
	ScopeName = "go.opentelemetry.io/collector/exporter/jvmhttpexporter"
	logger    = zap.NewNop()
)

const (
	ProfilesStability = component.StabilityLevelDevelopment
	TracesStability   = component.StabilityLevelStable
	MetricsStability  = component.StabilityLevelStable
	LogsStability     = component.StabilityLevelStable
)

func NewFactory() exporter.Factory {
	return xexporter.NewFactory(
		Type,
		createDefaultConfig,
		xexporter.WithTraces(createTraces, TracesStability),
		xexporter.WithMetrics(createMetrics, MetricsStability),
		xexporter.WithLogs(createLogs, LogsStability),
		xexporter.WithProfiles(createProfiles, ProfilesStability),
	)
}

func createMetrics(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	set.Logger.Info("jvmhttp exporter createMetrics ......")
	oce, err := newExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)
	oce.metricsURL, err = composeSignalURL(oCfg, oCfg.MetricsEndpoint, "metrics", "v1")
	set.Logger.Info("metrics URL：" + oce.metricsURL)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewMetrics(ctx, set, cfg,
		oce.pushMetrics,
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		// explicitly disable since we rely on http.Client timeout logic.
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig))
}

func createDefaultConfig() component.Config {
	logger.Info("jvmhttp exporter createDefaultConfig ......")
	clientConfig := confighttp.NewDefaultClientConfig()
	clientConfig.Timeout = 30 * time.Second
	// Default to gzip compression
	clientConfig.Compression = configcompression.TypeGzip
	// We almost read 0 bytes, so no need to tune ReadBufferSize.
	clientConfig.WriteBufferSize = 512 * 1024

	return &Config{
		RetryConfig:  configretry.NewDefaultBackOffConfig(),
		QueueConfig:  exporterhelper.NewDefaultQueueConfig(),
		Encoding:     EncodingJSON,
		ClientConfig: clientConfig,
	}
}

// composeSignalURL composes the final URL for the signal (traces, metrics, logs) based on the configuration.
// oCfg is the configuration of the exporter.
// signalOverrideURL is the URL specified in the signal specific configuration (empty if not specified).
// signalName is the name of the signal, e.g. "traces", "metrics", "logs".
// signalVersion is the version of the signal, e.g. "v1" or "v1development".
func composeSignalURL(oCfg *Config, signalOverrideURL string, signalName string, signalVersion string) (string, error) {
	switch {
	case signalOverrideURL != "":
		_, err := url.Parse(signalOverrideURL)
		if err != nil {
			return "", fmt.Errorf("%s_endpoint must be a valid URL", signalName)
		}
		return signalOverrideURL, nil
	case oCfg.Endpoint == "":
		return "", fmt.Errorf("either endpoint or %s_endpoint must be specified", signalName)
	default:
		if strings.HasSuffix(oCfg.Endpoint, "/") {
			return oCfg.Endpoint + signalVersion + "/" + signalName, nil
		}
		return oCfg.Endpoint + "/" + signalVersion + "/" + signalName, nil
	}
}

func createTraces(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {

	oce, err := newExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)

	oce.tracesURL, err = composeSignalURL(oCfg, oCfg.TracesEndpoint, "traces", "v1")
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTraces(ctx, set, cfg,
		oce.pushTraces,
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		// explicitly disable since we rely on http.Client timeout logic.
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig))
}

func createLogs(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	oce, err := newExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)
	oce.logsURL, err = composeSignalURL(oCfg, oCfg.LogsEndpoint, "logs", "v1")
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewLogs(ctx, set, cfg,
		oce.pushLogs,
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		// explicitly disable since we rely on http.Client timeout logic.
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig))
}

func createProfiles(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (xexporter.Profiles, error) {
	oce, err := newExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)

	oce.profilesURL, err = composeSignalURL(oCfg, "", "profiles", "v1development")
	if err != nil {
		return nil, err
	}

	return xexporterhelper.NewProfilesExporter(ctx, set, cfg,
		oce.pushProfiles,
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		// explicitly disable since we rely on http.Client timeout logic.
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig))
}
