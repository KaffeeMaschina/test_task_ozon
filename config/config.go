package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	Env           string `yaml:"env" env-default:"local" env-required:"true"`
	StorageConfig `yaml:"storage"`
	HTTPServer    `yaml:"http_server"`
}

type StorageConfig struct {
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DBHost   string `yaml:"host" env-required:"true"`
	DBPort   string `yaml:"port" env-required:"true"`
	Database string `yaml:"database" env-required:"true"`
}

type HTTPServer struct {
	ServerHost string `yaml:"host" env-default:"localhost"`
	ServerPort string `yaml:"port" env-default:"8080"`
}

func MustLoad() *Config {
	a := godotenv.Load()
	_ = a
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s doesn't exist", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %s", err)
	}
	return &cfg
}
