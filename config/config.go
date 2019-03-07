package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	ReleaseMode       bool              `json:"releaseMode"`
	Addr              string            `json:"addr"`
	Mongodb           Mongodb           `json:"mongodb"`
	VerificationEmail VerificationEmail `json:"verificationEmail"`
	FileStoragePath   string            `json:"fileStoragePath"`
	SlaveKey          string            `json:"slaveKey"`
	TokenSecretKey    string            `json:"tokenSecretKey"`
	BaseUrl			  string			`json:"baseUrl"`
	AccessControlAllowOrigin string     `json:"accessControlAllowOrigin"`
	ForwardedByClientIP bool			`json:"forwardedByClientIP"`
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

func GetBaseUrl() string {
	return config.BaseUrl
}

func GetAccessControlAllowOrigin() string {
	return config.AccessControlAllowOrigin
}

func IsReleaseMode() bool {
	return config.ReleaseMode
}

func IsForwardedByClientIP() bool {
	return config.ForwardedByClientIP
}