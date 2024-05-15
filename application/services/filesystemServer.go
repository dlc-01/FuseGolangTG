package services

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"fmt"
	"github.com/dlc-01/adapters/filesystem"
)

type FileSystemService struct {
	fileSystem *filesystem.FileSystemAdapter
	mountpoint string
}

func NewFileSystemService(fileSystem *filesystem.FileSystemAdapter, mount string) *FileSystemService {
	return &FileSystemService{
		fileSystem: fileSystem,
		mountpoint: mount,
	}
}

func (s *FileSystemService) Server() error {
	// Mount FUSE filesystem
	fuse.Unmount(s.mountpoint)

	fuseConn, err := fuse.Mount(
		s.mountpoint,
		fuse.FSName("telegramfs"),
		fuse.Subtype("telegramfs"),
		fuse.AllowOther(),
	)
	if err != nil {
		return fmt.Errorf("error mounting FUSE filesystem: %w", err)
	}
	defer fuseConn.Close()

	s.fileSystem.FuseConn = fuseConn
	filesys := filesystem.FS{FileSystem: s.fileSystem}
	if err := fs.Serve(fuseConn, filesys); err != nil {
		return fmt.Errorf("error serving FUSE filesystem: %w", err)
	}
	return nil
}

func (fs *FileSystemService) ReadFile(inode uint64, offset int64, size int) ([]byte, error) {
	data, err := fs.fileSystem.StoragePort.Read(inode, offset, size)
	if err != nil {
		return nil, err
	}

	file, err := fs.fileSystem.StoragePort.Lookup(inode, "")
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return fs.fileSystem.TelegramPort.DownloadFile(file.TelegramID)
	}

	return data, nil
}

func (fs *FileSystemService) WriteFile(inode uint64, data []byte) error {
	file, err := fs.fileSystem.StoragePort.Lookup(inode, "")
	if err != nil {
		return err
	}

	telegramID, messageID, err := fs.fileSystem.TelegramPort.UploadFile(file.Name, data, file.Tag)
	if err != nil {
		return err
	}

	err = fs.fileSystem.StoragePort.UpdateTelegramID(inode, telegramID)
	if err != nil {
		return err
	}

	return fs.fileSystem.TelegramPort.SaveMapping(telegramID, messageID)
}

func (fs *FileSystemService) Shutdown() error {
	if err := fuse.Unmount(fs.mountpoint); err != nil {
		return fmt.Errorf("Error unmounting FUSE filesystem: %v", err)
	}
	return nil
}
