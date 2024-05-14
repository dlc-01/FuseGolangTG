package telegram

import (
	"fmt"
	"github.com/dlc-01/config"
	"github.com/dlc-01/ports"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"net/http"
)

type TelegramAdapter struct {
	bot         *tgbotapi.BotAPI
	chatID      int64
	storagePort ports.FileStoragePort
}

func NewTelegramAdapter(cfg *config.Config, storagePort ports.FileStoragePort) (ports.TelegramPort, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}

	return &TelegramAdapter{
		bot:         bot,
		chatID:      cfg.TelegramChatID,
		storagePort: storagePort,
	}, nil
}

func (s *TelegramAdapter) UploadFile(filename string, data []byte, tag string) (string, int, error) {
	fileBytes := tgbotapi.FileBytes{Name: filename, Bytes: data}
	msg := tgbotapi.NewDocument(s.chatID, fileBytes)
	msg.Caption = tag

	message, err := s.bot.Send(msg)
	if err != nil {
		return "", 0, fmt.Errorf("failed to upload file: %w", err)
	}

	return message.Document.FileID, message.MessageID, nil
}

func (s *TelegramAdapter) DownloadFile(fileID string) ([]byte, error) {
	fileConfig := tgbotapi.FileConfig{FileID: fileID}
	tgFile, err := s.bot.GetFile(fileConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", s.bot.Token, tgFile.FilePath)
	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	return data, nil
}

func (s *TelegramAdapter) DeleteFile(fileID string) error {
	msgID, err := s.FindMessageIDByFileID(fileID)
	if err != nil {
		return fmt.Errorf("failed to find message ID: %w", err)
	}

	_, err = s.bot.Send(tgbotapi.DeleteMessageConfig{ChatID: s.chatID, MessageID: msgID})
	return err
}

func (s *TelegramAdapter) SaveMapping(fileID string, messageID int) error {
	return s.storagePort.SaveMapping(fileID, messageID)
}

func (s *TelegramAdapter) FindMessageIDByFileID(fileID string) (int, error) {
	return s.storagePort.FindMessageIDByFileID(fileID)
}
