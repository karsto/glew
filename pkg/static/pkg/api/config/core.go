package config

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/spf13/viper"
)

type Core struct {
	Port     string
	DBConfig pgx.ConnConfig
}

func NewCore(dir string) *Core {
	if dir == "" {
		dir = "."
	}
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(dir)

	viper.ReadInConfig()
	viper.AutomaticEnv()

	dbConfig, err := pgx.ParseEnvLibpq()
	if err != nil {
		panic(err)
	}

	// if not found default to localhost settings
	if len(dbConfig.Host) < 1 {
		fmt.Printf("\nWARNING: PGX variables not found. Defaulting. Verify PGHOST, PGPORT, PGDATABASE, PGUSER, and PGPASSWORD are set.")
		dbConfig = pgx.ConnConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "core",
			User:     "postgres",
			Password: "postgres",
		}
	}

	viper.SetDefault("PORT", 8080)

	return &Core{
		Port:     viper.GetString("PORT"),
		DBConfig: dbConfig,
	}
}

func (c *Core) Print() {
	fmt.Printf("\nPORT : %v\n", c.Port)
}
