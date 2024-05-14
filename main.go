package main

import (
	"flag"
	"github.com/dlc-01/config"
	"github.com/dlc-01/filesystem"
	"github.com/dlc-01/telegram"
	"log"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
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
	tgService, err := telegram.NewTelegramService(cfg)
	if err != nil {
		log.Fatalf("Error creating Telegram service: %v", err)
	}

	// Mount FUSE filesystem
	fuseConn, err := fuse.Mount(
		*mountpoint,
		fuse.FSName("telegramfs"),
		fuse.Subtype("telegramfs"),
	)
	if err != nil {
		log.Fatalf("Error mounting FUSE filesystem: %v", err)
	}
	defer fuseConn.Close()

	fs := &filesystem.FileSystem{
		TelegramService: tgService,
		Files:           make(map[string]filesystem.File),
		FuseConn:        fuseConn,
	}

	filesys := filesystem.FS{FileSystem: fs}
	if err := fusefs.Serve(fuseConn, filesys); err != nil {
		log.Fatalf("Error serving FUSE filesystem: %v", err)
	}

	// Wait for unmount.
	if err := fuseConn.Close(); err != nil {
		log.Fatalf("Mount process has exited with error: %v", err)
	}
}
