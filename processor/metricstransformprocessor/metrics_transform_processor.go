// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package metricstransformprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type metricsTransformProcessor struct {
	transforms               []internalTransform
	logger                   *zap.Logger
	otlpDataModelGateEnabled bool
}

type internalTransform struct {
	MetricIncludeFilter internalFilter
	Action              ConfigAction
	NewName             string
	GroupResourceLabels map[string]string

	SubmatchCase submatchCase
	Operations   []internalOperation
}

type internalOperation struct {
	configOperation     Operation
	valueActionsMapping map[string]string
	labelSetMap         map[string]bool
	aggregatedValuesSet map[string]bool
}

type internalFilter interface {
	getSubexpNames() []string
	matchMetric(pmetric.Metric) bool
	extractMatchedMetric(pmetric.Metric) pmetric.Metric
	expand(string, string) string
	submatches(pmetric.Metric) []int
	matchAttrs(pcommon.Map) bool
}

type StringMatcher interface {
	MatchString(string) bool
}

type strictMatcher string

func (s strictMatcher) MatchString(cmp string) bool {
	return string(s) == cmp
}

type internalFilterStrict struct {
	include      string
	attrMatchers map[string]StringMatcher
}

func (f internalFilterStrict) getSubexpNames() []string {
	return nil
}

type internalFilterRegexp struct {
	attrMatchers map[string]StringMatcher
}

func newMetricsTransformProcessor(logger *zap.Logger, internalTransforms []internalTransform) *metricsTransformProcessor {
	return &metricsTransformProcessor{
		transforms: internalTransforms,
		logger:     logger,
	}
}

func replaceCaseOfSubmatch(replacement submatchCase, submatch string) string {
	switch replacement {

	}

	return submatch
}
