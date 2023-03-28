package pkg

import (
	"os"

	"gopkg.in/yaml.v2"
)

// QuotaConfig struct contains Jobs
type QuotaConfig struct {
	Jobs []JobConfig `yaml:"jobs"`
}

// JobConfig struct
type JobConfig struct {
	ServiceCode string   `yaml:"serviceCode"`
	Regions     []string `yaml:"regions"`
}

// NewQuotaConfig creates a new QuotaConfig
func NewQuotaConfig(configFile string) (*QuotaConfig, error) {
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	qcl := QuotaConfig{}
	err = yaml.Unmarshal(yamlFile, &qcl)
	if err != nil {
		return nil, err
	}
	return &qcl, nil
}
