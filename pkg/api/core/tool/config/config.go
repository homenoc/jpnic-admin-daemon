package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strconv"
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
		return err
	}
	Conf = data
	ParseDatabase()

	return nil
}

func GetEnvConfig() error {
	var data Config
	var databasePort uint = 3306

	data.CaURL = os.Getenv("CA_URL")
	data.Database.Type = os.Getenv("DATABASE_TYPE")
	data.Database.Path = os.Getenv("DATABASE_PATH")
	data.Database.IP = os.Getenv("DATABASE_IP")
	portStr := os.Getenv("DATABASE_PORT")
	if portStr != "" {
		tmpPort, err := strconv.Atoi(portStr)
		if err != nil {
			return err
		}
		databasePort = uint(tmpPort)
	}
	data.Database.Port = databasePort
	data.Database.Name = os.Getenv("DATABASE_NAME")
	data.Database.User = os.Getenv("DATABASE_USER")
	data.Database.Pass = os.Getenv("DATABASE_PASS")

	Conf = data
	return nil
}
