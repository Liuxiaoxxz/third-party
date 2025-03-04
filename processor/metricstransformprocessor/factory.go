// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package metricstransformprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"

import (
	"github.com/Liuxiaoxxz/third-party/processor/metricstransformprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"golang.org/x/net/context"
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
	oCfg := cfg.(*Config)
	if err := validateConfiguration(oCfg); err != nil {
		return nil, err
	}

	hCfg, err := buildHelperConfig(oCfg, set.BuildInfo.Version)
	if err != nil {
		return nil, err
	}
	metricsProcessor := newMetricsTransformProcessor(set.Logger, hCfg)

	return processorhelper.NewMetrics(
		ctx,
		set,
		cfg,
		nextConsumer,
		metricsProcessor.processMetrics,
		processorhelper.WithCapabilities(consumerCapabilities))
}

// validateConfiguration validates the input configuration has all of the required fields for the processor
// An error is returned if there are any invalid inputs.
func validateConfiguration(config *Config) error {
	return nil
}

// buildHelperConfig constructs the maps that will be useful for the operations
func buildHelperConfig(config *Config, version string) ([]internalTransform, error) {
	helperDataTransforms := make([]internalTransform, len(config.Transforms))
	for i, t := range config.Transforms {
		if t.MetricIncludeFilter.MatchType == "" {
			t.MetricIncludeFilter.MatchType = strictMatchType
		}

		filter, err := createFilter(t.MetricIncludeFilter)
		if err != nil {
			return nil, err
		}

		helperT := internalTransform{
			MetricIncludeFilter: filter,
			Action:              t.Action,
			NewName:             t.NewName,
			GroupResourceLabels: t.GroupResourceLabels,
			Operations:          make([]internalOperation, len(t.Operations)),
		}

		for j, op := range t.Operations {

			mtpOp := internalOperation{
				configOperation: op,
			}
			if len(op.ValueActions) > 0 {
				mtpOp.valueActionsMapping = createLabelValueMapping(op.ValueActions, version)
			}
			if op.Action == aggregateLabels {
				mtpOp.labelSetMap = sliceToSet(op.LabelSet)
			} else if op.Action == aggregateLabelValues {
				mtpOp.aggregatedValuesSet = sliceToSet(op.AggregatedValues)
			}
			helperT.Operations[j] = mtpOp
		}
		helperDataTransforms[i] = helperT
	}
	return helperDataTransforms, nil
}

func createFilter(filterConfig FilterConfig) (internalFilter, error) {
	switch filterConfig.MatchType {
	case strictMatchType:
		matchers, err := getMatcherMap(filterConfig.MatchLabels, func(str string) (StringMatcher, error) { return strictMatcher(str), nil })
		if err != nil {
			return nil, err
		}
		return internalFilterStrict{include: filterConfig.Include, attrMatchers: matchers}, nil
	case regexpMatchType:
		return internalFilterRegexp{}, nil
	}

	return nil, nil
}

// createLabelValueMapping creates the labelValue rename mappings based on the valueActions
func createLabelValueMapping(valueActions []ValueAction, version string) map[string]string {
	mapping := make(map[string]string)
	for i := 0; i < len(valueActions); i++ {

	}
	return mapping
}

// sliceToSet converts slice of strings to set of strings
// Returns the set of strings
func sliceToSet(slice []string) map[string]bool {
	set := make(map[string]bool, len(slice))
	for _, s := range slice {
		set[s] = true
	}
	return set
}

func getMatcherMap(strMap map[string]string, ctor func(string) (StringMatcher, error)) (map[string]StringMatcher, error) {
	out := make(map[string]StringMatcher)
	for k, v := range strMap {
		matcher, err := ctor(v)
		if err != nil {
			return nil, err
		}
		out[k] = matcher
	}
	return out, nil
}
