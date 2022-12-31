package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DBConn *gorm.DB
)

func init() {
	var err error
	DBConn, err = gorm.Open(sqlite.Open("reservation.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
}
