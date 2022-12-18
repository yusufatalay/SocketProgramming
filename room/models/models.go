package models

import (
	"SocketProgramming/room/database"
	"log"
)

func init() {
	err := database.DBConn.AutoMigrate(&Room{}, &Reservation{})
	if err != nil {
		log.Fatalf("Cannot migrate models: %s", err.Error())
	}

}
