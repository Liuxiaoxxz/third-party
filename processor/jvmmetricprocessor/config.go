package jvmmetricprocessor

import (
	"fmt"
	"go.opentelemetry.io/collector/component"
)

type Config struct {
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func validateConfiguration(config *Config) error {
	fmt.Sprintf("validateConfiguration...{}", config)
	return nil
}
