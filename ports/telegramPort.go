package ports

// TelegramPort defines the interface for Telegram operations.
type TelegramPort interface {
	DeleteMessage(fileID string) error
	FindMessageIDByFileID(fileID string) int
	SaveMapping(fileID string, messageID int)
	RemoveMapping(fileID string)
	FetchFile(fileID string) ([]byte, error)
	UploadFile(fileName string, fileBytes []byte, tag string) (string, int, error)
}
