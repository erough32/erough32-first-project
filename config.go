package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type sqlConfigType struct {
	Host           string
	Port           string
	User           string
	Pass           string
	Database       string
	UserTable      string
	BookTable      string
	FavoritesTable string
}

type sessionsConfigType struct {
	Host     string
	Port     string
	Password string
}

var sqlConf sqlConfigType
var sessionConf sessionsConfigType
var commConfig []string

func init() {
	file, _ := os.Open("config/sql.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&sqlConf)
	if err != nil {
		fmt.Println("Error:", err)
	}

	file2, _ := os.Open("config/session.json")
	defer file2.Close()
	decoder2 := json.NewDecoder(file2)
	err2 := decoder2.Decode(&sessionConf)
	if err2 != nil {
		fmt.Println("Error:", err)
	}

	commFile, _ := ioutil.ReadFile("./config/comm.json")
	json.Unmarshal(commFile, &commConfig)
}

func sqlConfig(info string) string {
	if info == "use" {
		return sqlConf.User + ":" + sqlConf.Pass + "@tcp(" + sqlConf.Host + ":" + sqlConf.Port + ")/" + sqlConf.Database + "?charset=utf8mb4"
	} else if info == "database" {
		return sqlConf.Database
	} else if info == "userTable" {
		return sqlConf.UserTable
	} else if info == "bookTable" {
		return sqlConf.BookTable
	} else if info == "favoritesTable" {
		return sqlConf.FavoritesTable
	} else {
		return ""
	}
}

func sessionsConfig(info string) string {
	if info == "server" {
		return sessionConf.Host + ":" + sessionConf.Port
	} else if info == "password" {
		return sessionConf.Password
	} else {
		return ""
	}
}
