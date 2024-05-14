package ports

type TelegramPort interface {
	UploadFile(filename string, data []byte, tag string) (string, int, error)
	DownloadFile(fileID string) ([]byte, error)
	DeleteFile(fileID string) error
	SaveMapping(fileID string, messageID int) error
	FindMessageIDByFileID(fileID string) (int, error)
}
