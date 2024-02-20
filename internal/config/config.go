package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

var config *Config

type Config struct {
	Environment string
	Db          string
}

func initialize(environment string) error {
	if config != nil {
		return errors.New("config is already initialized")
	}

	data, err := os.ReadFile(path.Join("config", fmt.Sprintf("%s.json", environment)))
	if err != nil {
		return err
	}

	cfg := &Config{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}

	config = cfg
	config.Environment = environment
	return nil
}

func Load(environment string) (Config, error) {
	var err error = nil
	if config == nil {
		err = initialize(environment)
	} else if config.Environment != environment {
		err = errors.New(fmt.Sprintf("config already loaded with environment %s", config.Environment))
	}

	return *config, err
}
