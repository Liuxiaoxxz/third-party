package simpleexporter

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
	"net/http"
	"runtime"
)

type baseExporter struct {
	// Input configuration.
	config      *Config
	client      *http.Client
	tracesURL   string
	metricsURL  string
	logsURL     string
	profilesURL string
	logger      *zap.Logger
	settings    component.TelemetrySettings
	// Default user-agent header.
	userAgent string
}

const (
	headerRetryAfter         = "Retry-After"
	maxHTTPResponseReadBytes = 64 * 1024

	jsonContentType     = "application/json"
	protobufContentType = "application/x-protobuf"
)

func newExporter(cfg component.Config, set exporter.Settings) (*baseExporter, error) {
	logger.Info("newExporter ....")
	oCfg := cfg.(*Config)

	userAgent := fmt.Sprintf("%s/%s (%s/%s)",
		set.BuildInfo.Description, set.BuildInfo.Version, runtime.GOOS, runtime.GOARCH)

	// client construction is deferred to start
	return &baseExporter{
		config:    oCfg,
		logger:    set.Logger,
		userAgent: userAgent,
		settings:  set.TelemetrySettings,
	}, nil
}

func (e *baseExporter) start(ctx context.Context, host component.Host) error {
	e.logger.Info("Starting JVM HTTP exporter ......")
	return nil
}

func (e *baseExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	e.logger.Info("pushMetrics ......")
	return nil
}
