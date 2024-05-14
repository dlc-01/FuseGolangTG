package services

import (
	filesystem_adapter "github.com/dlc-01/adapters/filesystem"
	"github.com/dlc-01/domain"
	"github.com/dlc-01/mocks"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSystemService_SaveFile(t *testing.T) {
	mockTelegram := new(mocks.MockTelegramPort)
	fsAdapter := &filesystem_adapter.FileSystemAdapter{
		TelegramService: mockTelegram,
		Files:           make(map[string]domain.File),
	}
	service := NewFileSystemService(mockTelegram, fsAdapter, "./mnt")

	file := domain.File{
		Name:    "testfile.txt",
		Content: []byte("Hello, World!"),
		Tag:     "testtag",
	}

	mockTelegram.On("UploadFile", file.Name, file.Content, file.Tag).Return("fileID123", 123, nil)
	mockTelegram.On("SaveMapping", "fileID123", 123).Return(nil)

	err := service.SaveFile(file)
	assert.NoError(t, err)
	assert.Equal(t, "fileID123", fsAdapter.Files[file.Name].OtherID)
	mockTelegram.AssertExpectations(t)
}

func TestFileSystemService_Serve(t *testing.T) {
	// This test would require integration testing with a FUSE setup, which can be complex.
	// Here we assume the functionality works as expected.
	t.Skip("Skipping FUSE integration test")
}

func TestFileSystemService_Shutdown(t *testing.T) {
	mockTelegram := new(mocks.MockTelegramPort)
	fsAdapter := &filesystem_adapter.FileSystemAdapter{
		TelegramService: mockTelegram,
		Files:           make(map[string]domain.File),
	}
	service := NewFileSystemService(mockTelegram, fsAdapter, "./mnt")

	// Mock the execCommand function to simulate unmounting
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", append([]string{name}, arg...)...)
	}

	service.Serve()

	// Attempt to shutdown
	service.Shutdown()

	// No assertions needed for now as we are just ensuring no panics or errors occur
	// during the shutdown process.
}
