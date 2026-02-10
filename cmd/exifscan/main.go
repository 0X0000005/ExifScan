package main

import (
	"exifScan/internal/config"
	"exifScan/internal/db"
	"exifScan/internal/web"
	"flag"
	"log"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	if err := config.LoadConfig(*configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if config.AppConfig.Database.Enabled {
		if err := db.InitDB(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
	} else {
		log.Println("Database disabled in config.")
	}

	web.StartServer()
}
