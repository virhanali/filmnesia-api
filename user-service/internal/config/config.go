package config

import (
	"errors"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServicePort        string `mapstructure:"USER_SERVICE_PORT"`
	
	DBHost             string `mapstructure:"DB_HOST"`
	DBPort             string `mapstructure:"DB_PORT"`
	DBUser             string `mapstructure:"DB_USER"`
	DBPassword         string `mapstructure:"DB_PASSWORD"`
	DBName             string `mapstructure:"DB_NAME"`
	DBSSLMode          string `mapstructure:"DB_SSLMODE"`
	
	JWTSecretKey       string `mapstructure:"JWT_SECRET_KEY"`
	JWTExpirationHours int    `mapstructure:"JWT_EXPIRATION_HOURS"`
}

func LoadConfig(configPath string) (config Config, err error) {
	if configPath == "" {
		configPath = "../"
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Error reading config from '%s/.env': %v", configPath, err)
		return Config{}, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Printf("Error unmarshalling config from '%s/.env' to struct: %v", configPath, err)
		return Config{}, err
	}

	if config.ServicePort == "" || config.DBHost == "" || config.DBUser == "" || config.DBName == "" {
		log.Println("Error: One or more required config fields are empty after loading from .env.")
		return Config{}, errors.New("required config fields are empty in .env")
	}

	log.Printf("Successfully loaded config from '%s/.env'", configPath)
	log.Printf("Service Port config: %s", config.ServicePort)
	log.Printf("DB Host config: %s (DBName: %s)", config.DBHost, config.DBName)

	return config, nil
}
