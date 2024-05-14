package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	TelegramToken  string `json:"telegramToken"`
	TelegramChatID int64  `json:"telegramChatID"`
	MappingFile    string `json:"mappingFile"`
}

func LoadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}
