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
		if os.IsNotExist(err) {
			// Create default config
			AppConfig = &model.Config{}
			AppConfig.Server.Port = 8080
			AppConfig.Json.Enabled = true
			AppConfig.Json.Output = "scan_results.json"
			// Ensure initialized
			AppConfig.Scan.Path = ""
			AppConfig.Scan.Extensions = []string{".jpg", ".jpeg", ".png"}

			data, err := yaml.Marshal(AppConfig)
			if err != nil {
				return err
			}
			return os.WriteFile(path, data, 0644)
		}
		return err
	}
	AppConfig = &model.Config{}
	err = yaml.Unmarshal(file, AppConfig)
	if err != nil {
		return err
	}
	return nil
}
