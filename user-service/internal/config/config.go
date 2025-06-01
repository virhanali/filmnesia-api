package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServicePort string `mapstructure:"USER_SERVICE_PORT"`

	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBSSLMode  string `mapstructure:"DB_SSLMODE"`

	JWTSecretKey       string `mapstructure:"JWT_SECRET_KEY"`
	JWTExpirationHours int    `mapstructure:"JWT_EXPIRATION_HOURS"`
	RabbitMQURL        string `mapstructure:"RABBITMQ_URL"`
}

func LoadConfig(configPath string) (config Config, err error) {
	if configPath == "" {
		configPath = "../"
		log.Printf("INFO: Configuration path not provided to LoadConfig, using default: '%s' (assuming .env is in service root)", configPath)
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if readErr := viper.ReadInConfig(); readErr != nil {
		if _, ok := readErr.(viper.ConfigFileNotFoundError); ok {
			log.Printf("INFO: User Service config file '.env' not found in '%s'. Relying on environment variables.", configPath)
		} else {
			log.Printf("WARNING: Error reading User Service config file from '%s/.env': %v. Will attempt to use environment variables.", configPath, readErr)
		}
	} else {
		log.Printf("INFO: Successfully loaded User Service configuration from: %s", viper.ConfigFileUsed())
	}

	viper.BindEnv("USER_SERVICE_PORT")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("DB_SSLMODE")
	viper.BindEnv("JWT_SECRET_KEY")
	viper.BindEnv("JWT_EXPIRATION_HOURS")
	viper.BindEnv("RABBITMQ_URL")

	if unmarshalErr := viper.Unmarshal(&config); unmarshalErr != nil {
		log.Printf("ERROR: Error unmarshalling User Service config into struct: %v", unmarshalErr)
		return Config{}, unmarshalErr
	}

	if config.ServicePort == "" {
		config.ServicePort = "8081"
		log.Printf("INFO: USER_SERVICE_PORT not set, using default: %s", config.ServicePort)
	}
	if config.JWTExpirationHours <= 0 {
		config.JWTExpirationHours = 24
		log.Printf("INFO: JWT_EXPIRATION_HOURS not set or invalid, using default: %d hours", config.JWTExpirationHours)
	}

	log.Printf("User Service - Service Port loaded: [%s]", config.ServicePort)
	log.Printf("User Service - DB Host loaded: [%s] (DBName: [%s])", config.DBHost, config.DBName)
	log.Printf("User Service - RabbitMQ URL loaded: [%s]", config.RabbitMQURL)
	log.Printf("User Service - JWT Expiration Hours: [%d]", config.JWTExpirationHours)
	if config.JWTSecretKey == "" {
		log.Println("WARNING: JWT_SECRET_KEY IS EMPTY. THIS IS INSECURE.")
	}

	return config, nil
}
