package main

import (
	"flag"
	"fmt"
	"os"

	"goshare/config"
	"goshare/logger"
	"goshare/storage"
	"goshare/web"
)

func main() {
	configPath := flag.String("config", "/etc/goshare/config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Config file %s does not exist. Generating a default config...\n", *configPath)
			if genErr := config.GenerateDefaultConfig(*configPath); genErr != nil {
				fmt.Fprintf(os.Stderr, "Error generating default configuration: %v\n", genErr)
				os.Exit(1)
			}
			cfg, err = config.LoadConfig(*configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading generated configuration: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error loading configuration from %s: %v\n", *configPath, err)
			os.Exit(1)
		}
	}

	logger.Info("Starting GoShare...")

	manager, err := storage.NewManager(cfg)
	if err != nil {
		logger.Error("Failed to initialize storage manager: %v", err)
		os.Exit(1)
	}

	manager.StartCleanupTask()
	defer manager.StopCleanupTask()

	server := web.NewServer(cfg, manager)
	if err := server.Start(); err != nil {
		logger.Error("Server stopped: %v", err)
		os.Exit(1)
	}
}
