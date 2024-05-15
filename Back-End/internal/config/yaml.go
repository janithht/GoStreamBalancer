package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var config *Config

func Readconfig(configFilepath string) (*Config, error) {

	if config != nil {
		return config, nil
	}

	data, err := os.ReadFile(configFilepath) // Read the file
	if err != nil {
		return nil, fmt.Errorf("error reading config file from path '%s': '%v'", configFilepath, err)
	}

	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)

	}

	return config, nil
}
