package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Addr              string            `json:"addr"`
	Mongodb           Mongodb           `json:"mongodb"`
	VerificationEmail VerificationEmail `json:"verificationEmail"`
	FileStoragePath   string            `json:"fileStoragePath"`
	SlaveKey          string            `json:"slaveKey"`
	TokenSecretKey    string            `json:"tokenSecretKey"`
}

type Mongodb struct {
	Url      string `json:"url"`
	Database string `json:"database"`
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
	defer configFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

}

func GetAddr() string {
	return config.Addr
}

func GetMongodbConfig() Mongodb {
	return config.Mongodb
}

func GetVerificationEmailConfig() VerificationEmail {
	return config.VerificationEmail
}

func GetFileStoragePath() string {
	return config.FileStoragePath
}

func GetSlaveKey() string {
	return config.SlaveKey
}

func GetTokenSecretKey() string {
	return config.TokenSecretKey
}
