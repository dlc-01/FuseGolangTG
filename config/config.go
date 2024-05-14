package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	TelegramToken  string `json:"telegramToken"`
	TelegramChatID int64  `json:"telegramChatID""`
	PostgresURL    string `json:"postgresUrl"`
	MigrationDir   string `json:"migrationDir"`
}

func LoadConfig(filepath string) (*Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
