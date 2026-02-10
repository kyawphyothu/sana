package config

import (
	"os"
	"path/filepath"
)

const dbFileName = "sana.db"

type Config struct {
	DBType string
	DBName string
	DBPath string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBType: "sqlite3",
		DBName: dbFileName,
		DBPath: getDBPath(),
	}
	return cfg, nil
}

func getDBPath() string {
	env := os.Getenv("SANA_ENV")
	var configDir string
	var err error
	if env == "development" {
		configDir = "./data"
	} else {
		configDir, err = os.UserConfigDir()
		if err != nil {
			return filepath.Join(".", dbFileName)
		}
	}

	appName := "sana"
	appDir := filepath.Join(configDir, appName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		os.MkdirAll(appDir, 0755)
	}

	return filepath.Join(appDir, dbFileName)
}
