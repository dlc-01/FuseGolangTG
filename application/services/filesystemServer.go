package services

import (
	"fmt"
	filesystem_adapter "github.com/dlc-01/adapters/filesystem"
	"github.com/dlc-01/domain"
	"github.com/dlc-01/ports"
	"log"
	"os/exec"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var execCommand = exec.Command

type FileSystemService struct {
	telegramService ports.TelegramPort
	fileSystem      *filesystem_adapter.FileSystemAdapter
	mountpoint      string
}

func NewFileSystemService(tgService ports.TelegramPort, fsAdapter *filesystem_adapter.FileSystemAdapter, mountpoint string) *FileSystemService {
	return &FileSystemService{
		telegramService: tgService,
		fileSystem:      fsAdapter,
		mountpoint:      mountpoint,
	}
}

func (s *FileSystemService) Serve() error {
	// Mount FUSE filesystem
	fuseConn, err := fuse.Mount(
		s.mountpoint,
		fuse.FSName("telegramfs"),
		fuse.Subtype("telegramfs"),
	)
	if err != nil {
		return fmt.Errorf("error mounting FUSE filesystem: %w", err)
	}
	defer fuseConn.Close()

	s.fileSystem.FuseConn = fuseConn
	filesys := filesystem_adapter.FS{FileSystem: s.fileSystem}
	if err := fs.Serve(fuseConn, filesys); err != nil {
		return fmt.Errorf("error serving FUSE filesystem: %w", err)
	}

	//TODO красиво сделай
	if err := fuseConn.Close(); err != nil {
		return fmt.Errorf("mount process has exited with error: %w", err)
	}
	return nil
}

func (s *FileSystemService) SaveFile(file domain.File, content []byte) error {
	telegramID, messageID, err := s.telegramService.UploadFile(file.Name, content, file.Tag)
	if err != nil {
		return err
	}

	file.OtherID = telegramID
	s.telegramService.SaveMapping(telegramID, messageID)
	s.fileSystem.Files[file.Name] = file
	return nil
}

func (s *FileSystemService) Shutdown() {
	cmd := execCommand("fusermount", "-u", s.mountpoint)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error unmounting FUSE filesystem: %v", err)
	}
	log.Println("FUSE filesystem unmounted successfully")
}
