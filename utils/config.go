package utils

import (
	"github.com/spf13/viper"
	"os"
)

func ReadConfig(filename string) (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigType("json")
	configDir := os.Getenv("YOUCOMIC_INIT_CONFIG_DIR")
	if configDir == "" {
		configDir = "./init"
	}
	config.AddConfigPath(configDir)
	config.SetConfigName(filename)
	err := config.ReadInConfig()

	return config, err
}
