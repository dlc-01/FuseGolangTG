package filesystem

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/dlc-01/domain"
	"github.com/dlc-01/ports"
	"golang.org/x/net/context"
	"os"
	"syscall"
)

type FileSystemAdapter struct {
	StoragePort ports.FileStoragePort
	FuseConn    *fuse.Conn
}

func NewFileSystemAdapter(storagePort ports.FileStoragePort) *FileSystemAdapter {
	return &FileSystemAdapter{
		StoragePort: storagePort,
	}
}

type FS struct {
	FileSystem *FileSystemAdapter
}

func (fs FS) Root() (fs.Node, error) {
	return &Dir{fs: fs.FileSystem}, nil
}

type Dir struct {
	fs *FileSystemAdapter
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = syscall.S_IFDIR | 0755
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	file, err := d.fs.StoragePort.Lookup(1, name)
	if err != nil {
		return nil, fuse.ENOENT
	}
	return &FileNode{file: file, fs: d.fs}, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	files, err := d.fs.StoragePort.ReadDirAll(1)
	if err != nil {
		return nil, err
	}

	var dirents []fuse.Dirent
	for _, file := range files {
		dirents = append(dirents, fuse.Dirent{Name: file.Name, Type: fuse.DT_File})
	}
	return dirents, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	file, err := d.fs.StoragePort.Create(1, req.Name, uint32(req.Mode), req.Header.Uid, req.Header.Gid)
	if err != nil {
		return nil, nil, err
	}

	node := &FileNode{file: file, fs: d.fs}
	return node, node, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	return d.fs.StoragePort.Remove(1, req.Name)
}

type FileNode struct {
	file domain.File
	fs   *FileSystemAdapter
}

func (f *FileNode) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = uint64(f.file.ID)
	a.Mode = os.FileMode(syscall.S_IFREG | uint32(f.file.Mode))
	a.Size = uint64(f.file.Size)
	a.Atime = f.file.Atime
	a.Mtime = f.file.Mtime
	a.Ctime = f.file.Ctime
	return nil
}

func (f *FileNode) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	data, err := f.fs.StoragePort.Read(f.file.ID, req.Offset, req.Size)
	if err != nil {
		return err
	}
	resp.Data = data
	return nil
}

func (f *FileNode) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	err := f.fs.StoragePort.Write(f.file.ID, req.Offset, req.Data)
	if err != nil {
		return err
	}
	resp.Size = len(req.Data)
	return nil
}
