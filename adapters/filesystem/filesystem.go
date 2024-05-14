package filesystem_adapter

import (
	"context"
	"fmt"
	"github.com/dlc-01/domain"
	"github.com/dlc-01/ports"
	"os"
	"strings"
	"sync"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type FileSystemAdapter struct {
	TelegramService ports.TelegramPort
	Files           map[string]domain.File
	FilesMutex      sync.Mutex
	FuseConn        *fuse.Conn
}

// FS represents the root of the filesystem.
type FS struct {
	FileSystem *FileSystemAdapter
}

// Root returns the root directory node.
func (f FS) Root() (fs.Node, error) {
	return Dir{Path: "", FileSystem: f.FileSystem}, nil
}

// Dir represents a directory in the filesystem.
type Dir struct {
	Path       string
	FileSystem *FileSystemAdapter
}

// Attr sets the attributes for the directory.
func (d Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0755
	return nil
}

// Lookup looks up a file in the directory by name.
func (d Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	d.FileSystem.FilesMutex.Lock()
	defer d.FileSystem.FilesMutex.Unlock()

	fullPath := d.Path + "/" + name
	if file, ok := d.FileSystem.Files[fullPath]; ok {
		return FileNode{file, d.FileSystem}, nil
	}
	return nil, fuse.ENOENT
}

// ReadDirAll reads all entries in the directory.
func (d Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.FileSystem.FilesMutex.Lock()
	defer d.FileSystem.FilesMutex.Unlock()

	var dirents []fuse.Dirent
	prefix := d.Path + "/"
	for name := range d.FileSystem.Files {
		if strings.HasPrefix(name, prefix) {
			dirName := strings.TrimPrefix(name, prefix)
			dirents = append(dirents, fuse.Dirent{Name: dirName, Type: fuse.DT_File})
		}
	}
	return dirents, nil
}

func containsSlash(path string) bool {
	for _, c := range path {
		if c == '/' {
			return true
		}
	}
	return false
}

// Create creates a new file in the directory.
func (d Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	d.FileSystem.FilesMutex.Lock()
	defer d.FileSystem.FilesMutex.Unlock()

	fullPath := d.Path + "/" + req.Name
	file := domain.File{
		Name:  req.Name,
		CTime: time.Now(),
		MTime: time.Now(),
		Tag:   extractTag(fullPath),
	}
	d.FileSystem.Files[fullPath] = file
	
	node := FileNode{file, d.FileSystem}
	return node, node, nil
}

// Remove removes a file or directory.
func (d Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	d.FileSystem.FilesMutex.Lock()
	defer d.FileSystem.FilesMutex.Unlock()

	fullPath := d.Path + "/" + req.Name
	if file, ok := d.FileSystem.Files[fullPath]; ok {
		err := d.FileSystem.TelegramService.DeleteMessage(file.OtherID)
		if err != nil {
			return err
		}
		delete(d.FileSystem.Files, fullPath)
		d.FileSystem.TelegramService.RemoveMapping(file.OtherID)

		// Invalidate the entry to notify the kernel that the file has been removed
		err = d.FileSystem.FuseConn.InvalidateEntry(fuse.RootID, fullPath)
		if err != nil {
			fmt.Printf("Failed to invalidate entry: %v\n", err)
		}

		return nil
	}

	if strings.HasPrefix(req.Name, "dir_") {
		tag := strings.TrimPrefix(req.Name, "dir_")
		err := deleteFilesWithTag(d.FileSystem, tag)
		if err != nil {
			return err
		}
	}

	return fuse.ENOENT
}

// FileNode represents a file in the filesystem.
type FileNode struct {
	domain.File
	FileSystem *FileSystemAdapter
}

// Attr sets the attributes for the file.
func (f FileNode) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = uint64(time.Now().UnixNano())
	a.Mode = 0644
	a.Size = uint64(f.Size)
	a.Ctime = f.CTime
	a.Mtime = f.MTime
	return nil
}

// ReadAll reads the contents of the file.
func (f FileNode) ReadAll(ctx context.Context) ([]byte, error) {
	data, err := f.FileSystem.TelegramService.FetchFile(f.OtherID)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Read reads the contents of the file.
func (f FileNode) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	data, err := f.FileSystem.TelegramService.FetchFile(f.OtherID)
	if err != nil {
		return err
	}

	buf := make([]byte, req.Size)
	n := copy(buf, data[req.Offset:])
	if n < req.Size {
		resp.Data = buf[:n]
	} else {
		resp.Data = buf
	}
	return nil
}

// Write writes data to the file.
func (f FileNode) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	f.FileSystem.FilesMutex.Lock()
	defer f.FileSystem.FilesMutex.Unlock()

	fileBytes := req.Data
	fileName := f.Name

	telegramID, messageID, err := f.FileSystem.TelegramService.UploadFile(fileName, fileBytes, f.Tag)
	if err != nil {
		return err
	}

	file := f.FileSystem.Files[fileName]
	file.Size = int64(len(fileBytes))
	file.MTime = time.Now()
	file.OtherID = telegramID
	file.Tag = f.Tag
	f.FileSystem.Files[fileName] = file
	f.FileSystem.TelegramService.SaveMapping(telegramID, messageID)

	resp.Size = len(req.Data)
	return nil
}

func extractTag(fileName string) string {
	parts := strings.Split(fileName, "_")
	if len(parts) > 0 {
		tag := parts[0]
		tag = strings.ReplaceAll(tag, ".", "_")
		tag = strings.ReplaceAll(tag, "-", "_")
		return tag
	}
	return ""
}

func deleteFilesWithTag(fs *FileSystemAdapter, tag string) error {
	fs.FilesMutex.Lock()
	defer fs.FilesMutex.Unlock()

	for name, file := range fs.Files {
		if file.Tag == tag {
			err := fs.TelegramService.DeleteMessage(file.OtherID)
			if err != nil {
				return err
			}
			delete(fs.Files, name)
			fs.TelegramService.RemoveMapping(file.OtherID)

			err = fs.FuseConn.InvalidateEntry(fuse.RootID, name)
			if err != nil {
				fmt.Printf("Failed to invalidate entry: %v\n", err)
			}
		}
	}
	return nil
}
