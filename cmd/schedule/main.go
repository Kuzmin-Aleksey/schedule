package main

import (
	"log"
	"os"
	"schedule/config"
	"schedule/internal/app"
)

func main() {
	configPath := "config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatal("read config error: ", err)
	}

	app.Run(cfg)
}
