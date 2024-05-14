package domain

import "time"

type File struct {
	ID         uint64
	Name       string
	Uid        uint32
	Gid        uint32
	Mode       uint32
	Size       int64
	Mtime      time.Time
	Atime      time.Time
	Ctime      time.Time
	Rdev       uint64
	TelegramID string
	Tag        string
}
