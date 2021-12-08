package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type ElasticsearchCluster struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	API struct {
		Url string `yaml:"url"`
		Key string `yaml:"key"`
	} `yaml:"api"`

	UsageCluster ElasticsearchCluster `yaml:"usage_cluster"`
	StateCluster ElasticsearchCluster `yaml:"state_cluster"`
}

func LoadFromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file [%s]: %w", path, err)
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("unable to parse config file [%s]: %w", path, err)
	}

	return &c, nil
}
