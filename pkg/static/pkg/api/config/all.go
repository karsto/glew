package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type All struct {
	Port string `env:"all_port"`
	Core *Core
}

func (cfg *All) setDefaults() {
	viper.SetDefault("all_port", 8084)
	cfg.Core = NewCore("")
}

func (cfg *All) Print() {
	fmt.Printf("%+v\n", cfg.Port)
	cfg.Core.Print()
}

func NewAll(path string) *All {
	if path == "" {
		path = "."
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	cfg := All{}
	cfg.setDefaults()

	viper.ReadInConfig()
	viper.AutomaticEnv()

	viper.Unmarshal(&cfg, func(mp *mapstructure.DecoderConfig) {
		mp.TagName = "env"
	})

	return &cfg
}
