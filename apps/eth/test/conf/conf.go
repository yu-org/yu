package conf

import (
	"os"

	"gopkg.in/yaml.v3"
)

var Config *TestConfig

func init() {
	Config = NewDefaultConfig()
}

func LoadConfig(path string) error {
	if len(path) < 1 {
		return nil
	}
	d, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(d, Config)
}

type TestConfig struct {
	EthCaseConf *EthCaseConf `yaml:"ethCaseConf"`
}

type EthCaseConf struct {
	HostUrl         string `yaml:"hostUrl"`
	GenWalletCount  int    `yaml:"genWalletCount"`
	InitialEthCount uint64 `yaml:"initialEthCount"`
	TestSteps       int    `yaml:"testSteps"`
	RetryCount      int    `yaml:"retryCount"`
}

func NewDefaultConfig() *TestConfig {
	tc := &TestConfig{}
	tc.EthCaseConf = DefaultEthCaseConf()
	return tc
}

func DefaultEthCaseConf() *EthCaseConf {
	return &EthCaseConf{
		GenWalletCount:  2,
		InitialEthCount: 100 * 100 * 100,
		TestSteps:       1,
		RetryCount:      3,
	}
}
