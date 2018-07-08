package config

import (
    "encoding/json"
    "os"
    "log"
)

type Config struct {
    Addr string `json:"addr"`
    Mongodb Mongodb `json:"mongodb"`
}

type Mongodb struct {
        Url string `json:"url"`
        Database string `json:"database"`
}

var config Config

func Init() {
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
