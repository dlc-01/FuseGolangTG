package main

import (
	"flag"
	"github.com/dlc-01/adapters/filesystem"
	"github.com/dlc-01/adapters/postgres"
	"github.com/dlc-01/adapters/telegram"
	"github.com/dlc-01/application/services"
	"github.com/dlc-01/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	mountpoint := flag.String("mountpoint", "./mnt", "directory to mount the filesystem")
	configFile := flag.String("config", "config.json", "path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	storageAdapter, err := postgres.NewPostgresAdapter(cfg)
	if err != nil {
		log.Fatalf("Error creating storage adapter: %v", err)
	}

	tgAdapter, err := telegram.NewTelegramAdapter(cfg, storageAdapter)
	if err != nil {
		log.Fatalf("Error creating Telegram adapter: %v", err)
	}

	fsAdapter := filesystem.NewFileSystemAdapter(storageAdapter, tgAdapter)

	fileService := services.NewFileSystemService(fsAdapter, *mountpoint)

	go func() {
		if err := fileService.Server(); err != nil {
			log.Fatalf("Error serving FUSE filesystem: %v", err)
		}
	}()

	// Handle termination signals
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc

	log.Println("Shutting down the FUSE filesystem service...")
	if err := fileService.Shutdown(); err != nil {
		log.Fatal(err)
	}
	log.Println("FUSE filesystem unmounted successfully")
}
