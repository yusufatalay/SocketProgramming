package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var (
	DBConn *gorm.DB
)

// init function runs before everything, so we will create our database here
func init() {
	var err error
	DBConn, err = gorm.Open(sqlite.Open("room.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Cannot connect to the database: ", err)
	}
}
