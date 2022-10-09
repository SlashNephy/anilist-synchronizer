package main

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type config struct {
	AniListClientID     string `env:"ANILIST_CLIENT_ID,required"`
	AniListClientSecret string `env:"ANILIST_CLIENT_SECRET,required"`
}

func loadConfig() (*config, error) {
	envFile := flag.String("env-file", ".env", "path to .env file")
	flag.Parse()

	// .env がある場合だけ読み込む
	if _, err := os.Stat(*envFile); !os.IsNotExist(err) {
		if err = godotenv.Load(*envFile); err != nil {
			return nil, err
		}
	}

	config := &config{}
	if err := env.Parse(config); err != nil {
		return nil, err
	}

	return config, nil
}
