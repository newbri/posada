package util

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Name                 string        `yaml:"name"`
	DBDriver             string        `yaml:"db_driver"`
	DBSource             string        `yaml:"db_source"`
	MigrationURL         string        `yaml:"migration_url"`
	HTTPServerAddress    string        `yaml:"http_server_address"`
	TokenSymmetricKey    string        `yaml:"token_symmetric_key"`
	AccessTokenDuration  time.Duration `yaml:"access_token_duration"`
	RefreshTokenDuration time.Duration `yaml:"refresh_token_duration"`
	RedisAddress         string        `yaml:"http_redis_address"`
}

type App struct {
	Config map[string]*Config `yaml:"config"`
}

func LoadConfig(path string, env string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var app App
	err = yaml.Unmarshal(file, &app)
	if err != nil {
		return nil, err
	}

	config, ok := app.Config[env]
	if !ok {
		return nil, fmt.Errorf("the environment does not exist")
	}
	return config, err
}
