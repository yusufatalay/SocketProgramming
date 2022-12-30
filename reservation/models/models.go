package models

import (
	"github.com/yusufatalay/SocketProgramming/reservation/database"
	"log"
)

func init() {
	err := database.DBConn.AutoMigrate(&RoomReservation{})
	if err != nil {
		log.Fatalf("Cannot migrate models: %s", err.Error())
	}
}