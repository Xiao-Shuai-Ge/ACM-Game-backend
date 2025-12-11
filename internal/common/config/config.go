package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type ServerConfig struct {
	HTTP struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"http"`
	GRPC struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"grpc"`
}

type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

func load() (*Config, error) {
	vp := viper.New()
	vp.SetConfigName("config")
	vp.SetConfigType("yaml")
	vp.AddConfigPath("configs")

	if err := vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read base config: %w", err)
	}

	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if env != "" {
		vp.SetConfigName(fmt.Sprintf("config.%s", env))
		vp.SetConfigType("yaml")
		vp.AddConfigPath("configs")
		_ = vp.MergeInConfig()
	}

	vp.SetEnvPrefix("ACMGAME")
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.AutomaticEnv()

	var cfg Config
	if err := vp.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

var Module = fx.Options(
	fx.Provide(func() (*Config, error) { return load() }),
)
