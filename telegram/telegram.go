package telegram

import (
	"bufio"
	"fmt"
	"github.com/dlc-01/config"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramService struct {
	Bot    *tgbotapi.BotAPI
	Config config.Config
}

func NewTelegramService(cfg config.Config) (*TelegramService, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}
	return &TelegramService{Bot: bot, Config: cfg}, nil
}

func (s *TelegramService) DeleteMessage(fileID string) error {
	messageID := s.FindMessageIDByFileID(fileID)
	if messageID == 0 {
		return fmt.Errorf("message with file ID %s not found", fileID)
	}

	deleteMsg := tgbotapi.NewDeleteMessage(s.Config.TelegramChatID, messageID)
	_, err := s.Bot.Send(deleteMsg)
	return err
}

func (s *TelegramService) FindMessageIDByFileID(fileID string) int {
	file, err := os.Open(s.Config.MappingFile)
	if err != nil {
		log.Fatalf("Failed to open mapping file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) == 2 && parts[0] == fileID {
			var messageID int
			fmt.Sscanf(parts[1], "%d", &messageID)
			return messageID
		}
	}
	return 0
}

func (s *TelegramService) SaveMapping(fileID string, messageID int) {
	file, err := os.OpenFile(s.Config.MappingFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open mapping file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(fmt.Sprintf("%s:%d\n", fileID, messageID))
	if err != nil {
		log.Fatalf("Failed to write to mapping file: %v", err)
	}
	writer.Flush()
}

func (s *TelegramService) RemoveMapping(fileID string) {
	input, err := ioutil.ReadFile(s.Config.MappingFile)
	if err != nil {
		log.Fatalf("Failed to read mapping file: %v", err)
	}

	lines := strings.Split(string(input), "\n")
	var output []string
	for _, line := range lines {
		if !strings.HasPrefix(line, fileID+":") {
			output = append(output, line)
		}
	}

	err = ioutil.WriteFile(s.Config.MappingFile, []byte(strings.Join(output, "\n")), 0644)
	if err != nil {
		log.Fatalf("Failed to write mapping file: %v", err)
	}
}

func (s *TelegramService) FetchFile(fileID string) ([]byte, error) {
	fileConfig := tgbotapi.FileConfig{FileID: fileID}
	tgFile, err := s.Bot.GetFile(fileConfig)
	if err != nil {
		return nil, err
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", s.Config.TelegramToken, tgFile.FilePath)
	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *TelegramService) UploadFile(fileName string, fileBytes []byte, tag string) (string, int, error) {
	msg := tgbotapi.NewDocument(s.Config.TelegramChatID, tgbotapi.FileBytes{Name: fileName, Bytes: fileBytes})
	message, err := s.Bot.Send(msg)
	if err != nil {
		return "", 0, err
	}

	caption := fmt.Sprintf("#%s", tag)
	editMsg := tgbotapi.NewEditMessageCaption(s.Config.TelegramChatID, message.MessageID, caption)
	_, err = s.Bot.Send(editMsg)
	if err != nil {
		return "", 0, err
	}

	return message.Document.FileID, message.MessageID, nil
}
