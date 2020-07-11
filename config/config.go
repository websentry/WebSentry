package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ReleaseMode         bool              `json:"releaseMode"`
	Addr                string            `json:"addr"`
	Database            Database          `json:"database"`
	VerificationEmail   VerificationEmail `json:"verificationEmail"`
	FileStoragePath     string            `json:"fileStoragePath"`
	WorkerKey            string            `json:"workerKey"`
	TokenSecretKey      string            `json:"tokenSecretKey"`
	BackendURL          string            `json:"backendUrl"`
	CROSAllowOrigins    []string          `json:"crosAllowOrigins"`
	ForwardedByClientIP bool              `json:"forwardedByClientIP"`
}

type Database struct {
	DataSourceName string `json:"dataSourceName"`
	Type           string `json:"type"`
}

type VerificationEmail struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var config Config

func Load(file string) error {
	configFile, err := os.Open(file)
	if err != nil {
		return err
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		return err
	}

	err = configFile.Close()
	return err
}

func GetConfig() Config {
	return config
}
