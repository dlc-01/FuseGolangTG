package domain

import "time"

type File struct {
	ID      string
	Name    string
	Size    int64
	CTime   time.Time
	MTime   time.Time
	OtherID string
	Tag     string
	Content []byte
}
