package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type Config struct {
	CaURL    string   `yaml:"ca_url"`
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
	configPath := "./config.yaml"
	if inputConfPath != "" {
		configPath = inputConfPath
	}
	ConfigPath = configPath
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	var data Config
	err = yaml.Unmarshal(file, &data)
	if err != nil {
		log.Fatal(err)
	}
	Conf = data
	log.Println(Conf)
	ParseDatabase()

	return nil
}
