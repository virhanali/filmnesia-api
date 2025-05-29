package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	RabbitMQURL string `mapstructure:"RABBITMQ_URL"`
}

func LoadConfig(configPath string) (config Config, err error) {
	if configPath == "" {
		configPath = "."
		log.Printf("Configuration path not provided, using default: %s", configPath)
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	readErr := viper.ReadInConfig()
	if readErr != nil {
		if _, ok := readErr.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file '%s/.env': %v", configPath, readErr)
			return Config{}, readErr
		}
		log.Printf("Config file '%s/.env' not found. Relying on environment variables.", configPath)
	} else {
		log.Printf("Successfully loaded configuration from '%s/.env'", configPath)
	}

	unmarshalErr := viper.Unmarshal(&config)
	if unmarshalErr != nil {
		log.Printf("Error unmarshalling config into struct: %v", unmarshalErr)
		return Config{}, unmarshalErr
	}

	if config.RabbitMQURL == "" {
		envRabbitURL := os.Getenv("RABBITMQ_URL")
		if envRabbitURL != "" {
			config.RabbitMQURL = envRabbitURL
		} else {
			log.Println("WARNING: RABBITMQ_URL is not set in config, using default amqp://guest:guest@localhost:5672/")
			config.RabbitMQURL = "amqp://guest:guest@localhost:5672/"
		}
	}
	log.Printf("RabbitMQ URL loaded: %s", config.RabbitMQURL)
	return config, nil
}
