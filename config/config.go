package config

import (
    "encoding/json"
    "os"
    "log"
    "fmt"
)

type configStruct struct {
    Addr string `json:"addr"`
}

var config configStruct

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

    fmt.Println("aaa")
    fmt.Println(config.Addr)
}

func GetAddr() string {
    return config.Addr
}
