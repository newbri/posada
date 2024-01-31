package util

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Configuration interface {
	GetConfig() *Config
}

type Config struct {
	Name                    string        `yaml:"name"`
	DBDriver                string        `yaml:"db_driver"`
	DBSource                string        `yaml:"db_source"`
	MigrationURL            string        `yaml:"migration_url"`
	HTTPServerAddress       string        `yaml:"http_server_address"`
	TokenSymmetricKey       string        `yaml:"token_symmetric_key"`
	AccessTokenDuration     time.Duration `yaml:"access_token_duration"`
	RefreshTokenDuration    time.Duration `yaml:"refresh_token_duration"`
	RedisAddress            string        `yaml:"http_redis_address"`
	DefaultRole             string        `yaml:"default_role"`
	AuthorizationHeaderKey  string        `yaml:"authorization_header_key"`
	AuthorizationTypeBearer string        `yaml:"authorization_type_bearer"`
	AuthorizationPayloadKey string        `yaml:"authorization_payload_key"`
}

type configYAML struct {
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

	return &configYAML{
		config: config,
	}
}

func (app *configYAML) GetConfig() *Config {
	return app.config
}
