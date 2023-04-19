package pkg

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/awsutil"
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
	Role        string   `yaml:"role,omitempty"`
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

// String returns a string representation of QuotaConfig
func (q *QuotaConfig) String() string {
	return awsutil.Prettify(q)
}
