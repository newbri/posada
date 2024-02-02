package configuration

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
)

type ConfigYAML struct {
	config *Config
}

func NewYAMLConfiguration(path string, env string) Configuration {
	type data struct {
		Config map[string]*Config `yaml:"config"`
	}
	var yamlConfig *data

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}

	err = yaml.Unmarshal(fileBytes, &yamlConfig)
	if err != nil {
		log.Fatal().Msg("unable to unmarshal YAML config")
	}

	if "" == env {
		log.Fatal().Msg("the environment cannot be empty")
	}

	config, ok := yamlConfig.Config[env]
	if !ok {
		log.Fatal().Msg("the environment does not exist")
	}

	return &ConfigYAML{
		config: config,
	}
}

func (app *ConfigYAML) GetConfig() *Config {
	return app.config
}
