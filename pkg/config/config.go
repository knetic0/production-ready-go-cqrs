package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	configName = "config"
	configType = "yaml"
	configPath = "$PWD/"
)

func Read() *ApplicationConfig {
	env := os.Getenv("PROFILE")
	if env == "" {
		env = "local"
	}

	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var applicationConfig ApplicationConfig
	c := viper.Sub(env)
	err = c.Unmarshal(&applicationConfig)
	if err != nil {
		panic(fmt.Errorf("fatal error unmarshalling config: %w", err))
	}

	return &applicationConfig
}
