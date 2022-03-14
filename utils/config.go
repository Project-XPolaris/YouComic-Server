package utils

import "github.com/spf13/viper"

func ReadConfig(filename string) (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigType("json")
	config.AddConfigPath("./conf")
	config.SetConfigName(filename)
	err := config.ReadInConfig()

	return config, err
}
