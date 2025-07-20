package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	Env         string       `yaml:"env" env-default:"local"`
	DatabaseUrl string       `yaml:"database_url" env-required:"true"`
	Server      ServerConfig `yaml:"server" env-required:"true"`
}

type ServerConfig struct {
	Port string `yaml:"port" env-default:":8080"`
}

func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("Config file not found in path")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("Config file not found in path")
	}

	var config Config
	log.Printf("Loading config from %s", path)
	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic(err)
	}
	return &config

}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "./config/config.yaml", "config path")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
