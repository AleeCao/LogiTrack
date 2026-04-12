// Package config contains the config for the application
package config

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type EnvConfig struct {
	DBUser           string `mapstructure:"DB_USER"`
	DBPassword       string `mapstructure:"DB_PASSWORD"`
	DBHost           string `mapstructure:"DB_HOST"`
	DBPort           string `mapstructure:"DB_PORT"`
	DBName           string `mapstructure:"DB_NAME"`
	DBSsl            string `mapstructure:"DB_SSL"`
	MaxConns         int    `mapstructure:"DB_MAXCONNS"`
	Lifetime         int    `mapstructure:"DB_LIFECONN"`
	MaxIdles         int    `mapstructure:"DB_MAXIDLE"`
	NginxPort        string `mapstructure:"NGINX_PORT"`
	KafkaPort        string `mapstructure:"KAFKA_PORT"`
	KafkaHost        string `mapstructure:"KAFKA_HOST"`
	KafkaTopic       string `mapstructure:"KAFKA_TOPIC"`
	EsPort           string `mapstructure:"ES_PORT"`
	EsAddr           string `mapstructure:"ES_ADDR"`
	IngestionSerPort string `mapstructure:"SERVICE_INGESTION_PORT"`
	IngestionSerHost string `mapstructure:"SERVICE_INGESTION_HOST"`
	RedisPort        string `mapstructure:"REDIS_PORT"`
	RedisHost        string `mapstructure:"REDIS_HOST"`
	RedisPassword    string `mapstructure:"REDIS_PASSWORD"`
}

func VarConfig() (*EnvConfig, error) {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	envPath := filepath.Join(basepath, "../../.env")

	viper.SetConfigFile(envPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading .env: %w", err)
	}

	var config EnvConfig

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Environment can't be loaded: ", err)
	}

	return &config, nil
}
