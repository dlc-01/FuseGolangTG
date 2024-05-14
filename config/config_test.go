package config

//func TestLoadConfig(t *testing.T) {
//	configContent := `{
//		"telegramToken": "test_token",
//		"telegramChatID": 123456789,
//		"mappingFile": "file_mapping.txt"
//	}`
//	configFile := "test_config.json"
//	err := os.WriteFile(configFile, []byte(configContent), 0644)
//	assert.NoError(t, err)
//	defer os.Remove(configFile)
//
//	config, err := LoadConfig(configFile)
//	assert.NoError(t, err)
//	assert.Equal(t, "test_token", config.TelegramToken)
//	assert.Equal(t, int64(123456789), config.TelegramChatID)
//	assert.Equal(t, "file_mapping.txt", config.MappingFile)
//}
