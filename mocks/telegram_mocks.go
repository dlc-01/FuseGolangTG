package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockTelegramPort struct {
	mock.Mock
}

func (m *MockTelegramPort) DeleteMessage(fileID string) error {
	args := m.Called(fileID)
	return args.Error(0)
}

func (m *MockTelegramPort) FindMessageIDByFileID(fileID string) int {
	args := m.Called(fileID)
	return args.Int(0)
}

func (m *MockTelegramPort) SaveMapping(fileID string, messageID int) {
	m.Called(fileID, messageID)
}

func (m *MockTelegramPort) RemoveMapping(fileID string) {
	m.Called(fileID)
}

func (m *MockTelegramPort) FetchFile(fileID string) ([]byte, error) {
	args := m.Called(fileID)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockTelegramPort) UploadFile(fileName string, fileBytes []byte, tag string) (string, int, error) {
	args := m.Called(fileName, fileBytes, tag)
	return args.String(0), args.Int(1), args.Error(2)
}
