package config

import "time"

const MinimalAPIVersionRequired int = 0 // compile-time config

// Config struct holds the configuration for the server
/* Notice that server configuration differs from client's despite they share the same name. */
type Config struct {
	Main struct {
		IsDebugMode bool `mapstructure:"isDebugMode" default:"true"`
	} `mapstructure:"main"`

	Http struct {
		Port      uint   `mapstructure:"port" validate:"port" default:"4703"`
		Address   string `mapstructure:"address" validate:"ip_addr" default:"127.0.0.1"`
		RateLimit struct {
			Enabled   bool          `mapstructure:"enabled" default:"true"`
			Rate      float64       `mapstructure:"rate" validate:"gte=0" default:"2"`
			Burst     int           `mapstructure:"burst" validate:"gte=0" default:"7"`
			ExpiresIn time.Duration `mapstructure:"expiresIn" default:"3m"`
		} `mapstructure:"rateLimit"`
	} `mapstructure:"http"`

	Database struct {
		IsDebugMode bool   `mapstructure:"isDebugMode" default:"false"`
		DSN         string `mapstructure:"DSN" validate:"required" default:"./data.db"`
	} `mapstructure:"database"`

	Log struct {
		// log level is determined by MainConfig, instead of LogConfig.
		// LogConifg controls those exact style params of logs.
		IsColored bool `mapstructure:"isColored" default:"false"`
	} `mapstructure:"log"`

	Security struct {
		RegistercodeEnckey    string `mapstructure:"registercodeEnckey" validate:"required"`
		RegistercodeDeckey    string `mapstructure:"registercodeDeckey" validate:"required"`
		PasswordEnckey        string `mapstructure:"passwordEnckey" validate:"required"`
		PasswordDeckey        string `mapstructure:"passwordDeckey" validate:"required"`
		JwtAccessTokenEnckey  string `mapstructure:"jwtAccessTokenEnckey" validate:"required"`
		JwtAccessTokenDeckey  string `mapstructure:"jwtAccessTokenDeckey" validate:"required"`
		JwtRefreshTokenEnckey string `mapstructure:"jwtRefreshTokenEnckey" validate:"required"`
		JwtRefreshTokenDeckey string `mapstructure:"jwtRefreshTokenDeckey" validate:"required"`

		// use format like 1m30s and confx can read it
		JwtAccessTokenExpireTime time.Duration `mapstructure:"jwtAccessTokenExpireTime" validate:"required" default:"15m"`

		// use format like 1m30s and confx can read it
		JwtRefreshTokenExpireTime time.Duration `mapstructure:"jwtRefreshTokenExpireTime" validate:"required" default:"168h"`

		// revoked tokens will be cleared after then
		JwtRevokedTokensRetainTime time.Duration `mapstructure:"jwtRevokedTokensRetainTime" validate:"required" default:"72h"`

		// https
		TlsEnabled  bool   `mapstructure:"tlsEnabled" default:"false"`
		TlsCertfile string `mapstructure:"tlsCertfile" default:"cert.pem" validate:"required_if=TlsEnabled true"`
		TlsKeyfile  string `mapstructure:"tlsKeyfile" default:"key.pem" validate:"required_if=TlsEnabled true"`
	} `mapstructure:"security"`
}
