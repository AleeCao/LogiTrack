// Package config contains the config for the application
package config

import (
	"fmt"
	"log"

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
	IngestionSerPort string `mapstructure:"SERVICE_INGESTION_PORT"`
	IngestionSerHost string `mapstructure:"SERVICE_INGESTION_HOST"`
}

func VarConfig() (*EnvConfig, error) {
	viper.SetConfigFile("../../.env")
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
