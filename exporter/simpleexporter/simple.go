package simpleexporter

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
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
	client, err := e.config.ClientConfig.ToClient(ctx, host, e.settings)
	if err != nil {
		return err
	}
	e.client = client
	return nil
}

func (e *baseExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	e.logger.Info("pushMetrics ......")
	tr := pmetricotlp.NewExportRequestFromMetrics(md)
	var err error
	var request []byte
	switch e.config.Encoding {
	case EncodingJSON:
		request, err = tr.MarshalJSON()
	case EncodingProto:
		request, err = tr.MarshalProto()
	default:
		err = fmt.Errorf("invalid encoding: %s", e.config.Encoding)
	}
	e.logger.Info("Encode_metric：" + string(request))
	if err != nil {
		return consumererror.NewPermanent(err)
	}
	return e.export(ctx, e.metricsURL, request, e.metricsPartialSuccessHandler)
}

func (e *baseExporter) metricsPartialSuccessHandler(protoBytes []byte, contentType string) error {
	if protoBytes == nil {
		return nil
	}
	exportResponse := pmetricotlp.NewExportResponse()
	switch contentType {
	case protobufContentType:
		err := exportResponse.UnmarshalProto(protoBytes)
		if err != nil {
			return fmt.Errorf("error parsing protobuf response: %w", err)
		}
	case jsonContentType:
		err := exportResponse.UnmarshalJSON(protoBytes)
		if err != nil {
			return fmt.Errorf("error parsing json response: %w", err)
		}
	default:
		return nil
	}

	partialSuccess := exportResponse.PartialSuccess()
	if !(partialSuccess.ErrorMessage() == "" && partialSuccess.RejectedDataPoints() == 0) {
		e.logger.Warn("Partial success response",
			zap.String("message", exportResponse.PartialSuccess().ErrorMessage()),
			zap.Int64("dropped_data_points", exportResponse.PartialSuccess().RejectedDataPoints()),
		)
	}
	return nil
}

func (e *baseExporter) export(ctx context.Context, url string, request []byte, handler func(protoBytes []byte, contentType string) error) error {
	e.logger.Debug("Preparing to make HTTP request", zap.String("url", url))
	return nil
}
