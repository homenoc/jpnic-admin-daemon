package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	NextTime string   `yaml:"next_time"`
	Port     int      `yaml:"port"`
	Database Database `yaml:"database"`
}

type Database struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
	IP   string `yaml:"ip"`
	Port uint   `yaml:"port"`
	Name string `yaml:"name"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

var Conf Config
var ConfigPath string

func GetConfig(inputConfPath string) error {
	configPath := "./config.json"
	if inputConfPath != "" {
		configPath = inputConfPath
	}
	ConfigPath = configPath
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	var data Config
	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Fatal(err)
	}
	Conf = data
	return nil
}
