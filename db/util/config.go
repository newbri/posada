package util

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Configuration interface {
	GetConfig(env string) (*Config, error)
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
	Config map[string]*Config `yaml:"config"`
}

func NewYAMLConfiguration(path string) (Configuration, error) {
	var yamlConfig configYAML

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("cannot load config")
	}

	err = yaml.Unmarshal(fileBytes, &yamlConfig)
	if err != nil {
		return nil, err
	}

	return &yamlConfig, nil
}

func (app *configYAML) GetConfig(env string) (*Config, error) {
	config, ok := app.Config[env]
	if !ok {
		return nil, fmt.Errorf("the environment does not exist")
	}
	return config, nil
}
