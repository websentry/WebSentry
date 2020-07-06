package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	ReleaseMode         bool              `json:"releaseMode"`
	Addr                string            `json:"addr"`
	Database            Database          `json:"database"`
	VerificationEmail   VerificationEmail `json:"verificationEmail"`
	FileStoragePath     string            `json:"fileStoragePath"`
	SlaveKey            string            `json:"slaveKey"`
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

func init() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	err = configFile.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func GetConfig() Config {
	return config
}
