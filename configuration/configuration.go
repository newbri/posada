package configuration

import (
	"time"
)

type Configuration interface {
	GetConfig() *Config
}

type Config struct {
	Name                      string        `yaml:"name"`
	DBDriver                  string        `yaml:"db_driver"`
	DBSource                  string        `yaml:"db_source"`
	MigrationURL              string        `yaml:"migration_url"`
	HTTPServerAddress         string        `yaml:"http_server_address"`
	TokenSymmetricKey         string        `yaml:"token_symmetric_key"`
	AccessTokenDuration       time.Duration `yaml:"access_token_duration"`
	RefreshTokenDuration      time.Duration `yaml:"refresh_token_duration"`
	RedisAddress              string        `yaml:"http_redis_address"`
	DefaultRole               string        `yaml:"default_role"`
	AuthorizationHeaderKey    string        `yaml:"authorization_header_key"`
	AuthorizationTypeBearer   string        `yaml:"authorization_type_bearer"`
	AuthorizationPayloadKey   string        `yaml:"authorization_payload_key"`
	AccessControlAllowOrigin  string        `yaml:"access_control_allow_origin"`
	AccessControlAllowHeaders string        `yaml:"access_control_allow_headers"`
	AccessControlAllowMethods string        `yaml:"access_control_allow_methods"`
}
