package config

import (
	"log"
	"strconv"
)

type SqlDatabase struct {
	Driver string
	Option string
}

var ConfDatabase SqlDatabase

func ParseDatabase() {
	switch Conf.Database.Type {
	case "sqlite3":
		ConfDatabase.Driver = "sqlite3"
		ConfDatabase.Option = "file:" + Conf.Database.Path + "?cache=shared&mode=rwc&_journal_mode=WAL"
	case "mysql":
		ConfDatabase.Driver = "mysql"
		ConfDatabase.Option = Conf.Database.User + ":" + Conf.Database.Pass + "@(" +
			Conf.Database.Name + ":" + strconv.Itoa(int(Conf.Database.Port)) + ")/" + Conf.Database.Name
	default:
		log.Fatal("Invalid database config.")
		return
	}
}
