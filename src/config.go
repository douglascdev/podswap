package podswap

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/goccy/go-yaml"
)

type workflowYaml struct {
	Jobs struct {
		Podswap struct {
			Uses string `yaml:"uses"`
			With Config `yaml:"with"`
		} `yaml:"podswap"`
	} `yaml:"jobs"`
}

type Config struct {
	PreBuildCmd string `yaml:"pre-build-cmd"`
	BuildCmd    string `yaml:"build-cmd"`
	DeployCmd   string `yaml:"deploy-cmd"`
}

func (c *Config) Validate() error {
	fields := reflect.ValueOf(c).Elem()
	for i := 0; i < fields.NumField(); i++ {
		if fields.Field(i).IsZero() {
			field := fields.Type().Field(i).Tag.Get("yaml")
			return fmt.Errorf("field %s not set in the yaml file", field)
		}
	}
	return nil
}

func NewConfig(path string) (cfg *Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to open config file %q: %w", path, err)
	}

	cfgBytes, err := io.ReadAll(file)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	var workflowYaml workflowYaml
	err = yaml.Unmarshal(cfgBytes, &workflowYaml)
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config file %q, make sure the 'podswap' job exists and that your yaml is valid: %w", path, err)
	}

	cfg = &workflowYaml.Jobs.Podswap.With
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, err
}
