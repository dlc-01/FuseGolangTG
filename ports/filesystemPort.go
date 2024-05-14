package ports

import (
	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// FileSystemPort defines the interface for file system operations.
type FileSystemPort interface {
	Root() (fs.Node, error)
}

// DirPort defines the interface for directory operations.
type DirPort interface {
	Attr(ctx context.Context, a *fuse.Attr) error
	Lookup(ctx context.Context, name string) (fs.Node, error)
	ReadDirAll(ctx context.Context) ([]fuse.Dirent, error)
	Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error)
	Remove(ctx context.Context, req *fuse.RemoveRequest) error
}

// FileNodePort defines the interface for file operations.
type FileNodePort interface {
	Attr(ctx context.Context, a *fuse.Attr) error
	ReadAll(ctx context.Context) ([]byte, error)
	Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error
	Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error
}
