package database

import (
	"fmt"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/tool/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"strconv"
)

type Base struct {
	DB *gorm.DB
}

func Connect() (*Base, error) {
	var db *gorm.DB
	var err error = nil

	switch config.Conf.Database.Type {
	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(config.Conf.Database.Path), &gorm.Config{})
	case "mysql":
		dsn := config.Conf.Database.User + ":" + config.Conf.Database.Pass +
			"@tcp(" + config.Conf.Database.IP + ":" + strconv.Itoa(int(config.Conf.Database.Port)) + ")/" + config.Conf.Database.Name +
			"?charset=utf8mb4&parseTime=True"
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	default:
		log.Println("Invalid database config.")
		return nil, fmt.Errorf("Invalid database config")
	}

	if err != nil {
		return nil, err
	}
	return &Base{DB: db}, err
}
