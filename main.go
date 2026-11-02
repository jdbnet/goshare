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

var Version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version")
	configPath := flag.String("config", "/etc/goshare/config.yaml", "Path to configuration file")
	flag.Parse()

	if *showVersion {
		fmt.Println(Version)
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
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
