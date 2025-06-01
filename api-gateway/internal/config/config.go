package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	APIGatewayPort string `mapstructure:"API_GATEWAY_PORT"`
	UserServiceURL string `mapstructure:"USER_SERVICE_URL"`
}

func LoadConfig(configDirHint string) (config Config, err error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current working directory: %v", err)
	}
	log.Printf("DEBUG: Current working directory: %s", wd)
	log.Printf("DEBUG: configDirHint received: %s", configDirHint)

	viper.AddConfigPath(".")

	if configDirHint != "" {
		absHintPath, pathErr := filepath.Abs(configDirHint)
		if pathErr == nil {
			log.Printf("DEBUG: Adding hint path to Viper (absolute): %s", absHintPath)
			viper.AddConfigPath(absHintPath)
		} else {
			log.Printf("DEBUG: Could not make hint path absolute, adding as is: %s", configDirHint)
			viper.AddConfigPath(configDirHint)
		}
	}

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	readErr := viper.ReadInConfig()
	if readErr != nil {
		if _, ok := readErr.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading API Gateway config file: %v. Searched paths included '.' and hint '%s'", readErr, configDirHint)
			return Config{}, readErr
		}
		log.Printf("API Gateway config file '.env' not found in searched paths (e.g., '.', '%s'). Relying on environment variables.", configDirHint)
	} else {
		configFileUsed := viper.ConfigFileUsed()
		log.Printf("Successfully loaded API Gateway configuration from: %s", configFileUsed)
	}
	unmarshalErr := viper.Unmarshal(&config)
	if unmarshalErr != nil {
		log.Printf("Error unmarshalling API Gateway config into struct: %v", unmarshalErr)
		return Config{}, unmarshalErr
	}

	if config.APIGatewayPort == "" {
		config.APIGatewayPort = os.Getenv("API_GATEWAY_PORT")
		if config.APIGatewayPort == "" {
			config.APIGatewayPort = "8000"
		}
	}
	if config.UserServiceURL == "" {
		config.UserServiceURL = os.Getenv("USER_SERVICE_URL")
	}

	log.Printf("API Gateway Port loaded: [%s]", config.APIGatewayPort)
	log.Printf("User Service URL loaded: [%s]", config.UserServiceURL)

	if config.UserServiceURL == "" {
		return Config{}, errors.New("USER_SERVICE_URL must be configured either in .env or as an environment variable")
	}

	return config, nil
}
