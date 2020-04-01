package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	databasePath = "data.sqlite"
	logFilePath  = "log.txt"
)

func main() {

	flags := initializeFlags()

	db, e := sql.Open("sqlite3", databasePath)
	if e != nil {
		panic("Error opening database")
	}

	e = createTables(db)
	if e != nil {
		panic("Could not create database tables")
	}

	f, e := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if e != nil {
		panic("Could not create a log file")
	}
	logger := log.New(f, "", log.Ldate|log.Ltime)

	ie := dispatch(db, flags)
	if ie != nil {
		fmt.Println(ie.Message())
		logger.Println(ie.Error())
	}
}
