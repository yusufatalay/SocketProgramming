package models

import (
	"github.com/yusufatalay/SocketProgramming/activity/database"
	"log"
)

func init() {
	err := database.DBConn.AutoMigrate(&Activity{})
	if err != nil {
		log.Fatalf("Cannot migrate models: %s", err.Error())
	}

}
