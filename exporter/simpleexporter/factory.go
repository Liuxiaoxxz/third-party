package simpleexporter

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/xexporter"
	"go.uber.org/zap"
	"net/url"
	"strings"
	"time"
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
	return xexporter.NewFactory(
		Type,
		createDefaultConfig,
		xexporter.WithMetrics(createMetrics, MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	clientConfig := confighttp.NewDefaultClientConfig()
	clientConfig.Timeout = 30 * time.Second
	// Default to gzip compression
	clientConfig.Compression = configcompression.TypeGzip
	// We almost read 0 bytes, so no need to tune ReadBufferSize.
	clientConfig.WriteBufferSize = 512 * 1024

	return &Config{
		RetryConfig:  configretry.NewDefaultBackOffConfig(),
		QueueConfig:  exporterhelper.NewDefaultQueueConfig(),
		Encoding:     EncodingProto,
		ClientConfig: clientConfig,
	}
}

func createMetrics(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Metrics, error) {
	set.Logger.Info("simple exporter started")
	oce, err := newExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)

	oce.metricsURL, err = composeSignalURL(oCfg, oCfg.MetricsEndpoint, "metrics", "v1")
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
		exporterhelper.WithQueue(oCfg.QueueConfig),
	)

}

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
