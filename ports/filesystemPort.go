package ports

import (
	"github.com/dlc-01/domain"
)

type FileStoragePort interface {
	Lookup(parentInode uint64, name string) (domain.File, error)
	ReadDirAll(parentInode uint64) ([]domain.File, error)
	Create(parentInode uint64, name string, mode uint32, uid uint32, gid uint32) (domain.File, error)
	Remove(parentInode uint64, name string) error
	UpdateTelegramID(inode uint64, telegramID string) error
	FindMessageIDByFileID(fileID string) (int, error)
	SaveMapping(fileID string, messageID int) error
	Write(inode uint64, offset int64, data []byte) error
	Read(inode uint64, offset int64, size int) ([]byte, error)
}
