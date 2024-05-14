package main

import (
	"flag"
	filesystem_adapter "github.com/dlc-01/adapters/filesystem"
	telegram_adapter "github.com/dlc-01/adapters/telegram"
	"github.com/dlc-01/application/services"
	"github.com/dlc-01/config"
	"github.com/dlc-01/domain"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Parse command-line flags
	mountpoint := flag.String("mountpoint", "./mnt", "directory to mount the filesystem")
	configFile := flag.String("config", "config.json", "path to the configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize Telegram Service
	tgService, err := telegram_adapter.NewTelegramAdapter(cfg)
	if err != nil {
		log.Fatalf("Error creating Telegram service: %v", err)
	}

	// Initialize FileSystem Adapter
	fsAdapter := &filesystem_adapter.FileSystemAdapter{
		TelegramService: tgService,
		Files:           make(map[string]domain.File),
	}

	// Initialize FileSystem Service
	fsService := services.NewFileSystemService(tgService, fsAdapter, *mountpoint)

	// Set up signal handling to ensure clean shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := fsService.Serve(); err != nil {
			log.Fatalf("Error serving FUSE filesystem: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down the FUSE filesystem service...")
	fsService.Shutdown()
}
