package filesystem

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/dlc-01/domain"
	"github.com/dlc-01/ports"
	"golang.org/x/net/context"
	"log"
	"os"
	"sync"
	"time"
)

const FileMaxSizeBytes int = 2 * 1e9

type FileSystemAdapter struct {
	StoragePort  ports.FileStoragePort
	TelegramPort ports.TelegramPort
	FilesMutex   sync.Mutex
	FuseConn     *fuse.Conn
}

func NewFileSystemAdapter(storagePort ports.FileStoragePort, telegramPort ports.TelegramPort) *FileSystemAdapter {
	return &FileSystemAdapter{
		StoragePort:  storagePort,
		TelegramPort: telegramPort,
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
	a.Mode = os.ModeDir | 0o555
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
		log.Println(err)
		return nil, fuse.ENOENT
	}

	var dirents []fuse.Dirent
	for _, file := range files {
		log.Println(file.Name)
		dirents = append(dirents, fuse.Dirent{Name: file.Name, Type: fuse.DT_File})
	}
	return dirents, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {

	file := domain.File{

		Name:  req.Name,
		Uid:   req.Uid,
		Gid:   req.Gid,
		Mode:  uint32(req.Mode),
		Atime: time.Now(),
	}
	node := &FileNode{file: file, fs: d.fs}
	return node, node, nil
}

//func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
//	return d.fs.StoragePort.Remove(1, req.Name)
//}

type FileNode struct {
	file domain.File
	fs   *FileSystemAdapter
}

func (f *FileNode) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.file.ID
	a.Mode = os.FileMode(f.file.Mode)
	a.Size = uint64(f.file.Size)
	a.Atime = f.file.Atime
	a.Mtime = f.file.Mtime
	a.Ctime = f.file.Ctime
	a.Uid = f.file.Uid
	a.Gid = f.file.Gid
	return nil
}

func (f *FileNode) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	data, err := f.fs.TelegramPort.DownloadFile(f.file.TelegramID)
	if err != nil {
		log.Println(err)
		return err
	}

	resp.Data = data

	return nil
}

func (f *FileNode) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	//err := f.fs.StoragePort.Write(f.file.ID, req.Offset, req.Data)
	//if err != nil {
	//	return err
	//}
	//resp.Size = len(req.Data)
	//log.Println("loh")

	file, err := f.fs.StoragePort.Create(1, f.file.Name, f.file.Mode, req.Header.Uid, req.Header.Gid, uint64(len(req.Data)))
	if err != nil {
		return err
	}
	f.file.ID = file.ID
	telegramID, messageID, err := f.fs.TelegramPort.UploadFile(f.file.Name, req.Data, f.file.Tag)
	if err != nil {
		log.Printf("%v upload", err)
		return err
	}

	err = f.fs.StoragePort.UpdateTelegramID(f.file.ID, telegramID)
	if err != nil {
		log.Printf("%v tgId", err)
		return err
	}

	return f.fs.TelegramPort.SaveMapping(telegramID, messageID)
}
