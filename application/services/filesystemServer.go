package services

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"fmt"
	"github.com/dlc-01/adapters/filesystem"
	"github.com/dlc-01/domain"
	"github.com/dlc-01/ports"
)

type FileSystemService struct {
	telegramPort ports.TelegramPort
	fileSystem   *filesystem.FileSystemAdapter
	mountpoint   string
}

func NewFileSystemService(tgPort ports.TelegramPort, fileSystem *filesystem.FileSystemAdapter, mount string) *FileSystemService {
	return &FileSystemService{
		telegramPort: tgPort,
		fileSystem:   fileSystem,
		mountpoint:   mount,
	}
}

func (s *FileSystemService) Server() error {
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
	filesys := filesystem.FS{FileSystem: s.fileSystem}
	if err := fs.Serve(fuseConn, filesys); err != nil {
		return fmt.Errorf("error serving FUSE filesystem: %w", err)
	}

	//TODO красиво сделай
	if err := fuseConn.Close(); err != nil {
		return fmt.Errorf("mount process has exited with error: %w", err)
	}
	return nil
}

func (fs *FileSystemService) Lookup(parentInode uint64, name string) (domain.File, error) {
	return fs.fileSystem.StoragePort.Lookup(parentInode, name)
}

func (fs *FileSystemService) ReadDirAll(parentInode uint64) ([]domain.File, error) {
	return fs.fileSystem.StoragePort.ReadDirAll(parentInode)
}

func (fs *FileSystemService) Create(parentInode uint64, name string, mode uint32, uid uint32, gid uint32) (domain.File, error) {
	return fs.fileSystem.StoragePort.Create(parentInode, name, mode, uid, gid)
}

func (fs *FileSystemService) Remove(parentInode uint64, name string) error {
	return fs.fileSystem.StoragePort.Remove(parentInode, name)
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
		return fs.telegramPort.DownloadFile(file.TelegramID)
	}

	return data, nil
}

func (fs *FileSystemService) WriteFile(inode uint64, data []byte) error {
	file, err := fs.fileSystem.StoragePort.Lookup(inode, "")
	if err != nil {
		return err
	}

	telegramID, messageID, err := fs.telegramPort.UploadFile(file.Name, data, file.Tag)
	if err != nil {
		return err
	}

	err = fs.fileSystem.StoragePort.UpdateTelegramID(inode, telegramID)
	if err != nil {
		return err
	}

	return fs.telegramPort.SaveMapping(telegramID, messageID)
}

func (fs *FileSystemService) Shutdown() error {
	if err := fuse.Unmount(fs.mountpoint); err != nil {
		return fmt.Errorf("Error unmounting FUSE filesystem: %v", err)
	}
	return nil
}
