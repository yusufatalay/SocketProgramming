package models

import (
	"log"

	"github.com/yusufatalay/SocketProgramming/reservation/database"
)

func init() {
	err := database.DBConn.AutoMigrate(&RoomReservation{})
	if err != nil {
		log.Fatalf("Cannot migrate models: %s", err.Error())
	}
}