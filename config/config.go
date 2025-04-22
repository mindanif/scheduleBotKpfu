package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	TelegramToken    string `yaml:"telegram_token"`
	KampusAPIBaseUrl string `yaml:"kampus_api_base_url"`
	TeacherAPIUrl    string `yaml:"teacher_api_url"`
	TeacherAPIToken  string `yaml:"teacher_api_token"`
	KFUApiBaseUrl    string `yaml:"kfu_api_base_url"`
}

func NewConfigFromFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл конфигурации: %v", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл конфигурации: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("не удалось разобрать YAML-конфигурацию: %v", err)
	}

	return &cfg, nil
}
