package config

import (
	"exifScan/internal/model"
	"os"

	"gopkg.in/yaml.v3"
)

var AppConfig *model.Config

func LoadConfig(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	AppConfig = &model.Config{}
	err = yaml.Unmarshal(file, AppConfig)
	if err != nil {
		return err
	}
	return nil
}
