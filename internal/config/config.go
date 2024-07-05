package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Env        string `yaml:"env" env-default:"prod"`
	AiToken    string `yaml:"ai_token" env-required:"true"`
	TgToken    string `yaml:"tg_token" env-required:"true"`
	StaticPath string `yaml:"static_path" env-required:"true"`
	Proxy      Proxy  `yaml:"proxy" env-required:"true"`
}

type Proxy struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"user"`
	Password string `yaml:"pass"`
}

const configPathEnv = "config_path"

func MustLoad() *Config {
	const op = "internal.config.MustLoad"

	path, ok := os.LookupEnv(configPathEnv)
	if !ok || path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}
