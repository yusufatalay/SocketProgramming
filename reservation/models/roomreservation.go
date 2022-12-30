package models

import (
	"log"

	"github.com/yusufatalay/SocketProgramming/reservation/database"
)

type RoomReservation struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	ActivityName string `json:"activity_name"`
	RoomName     string `json:"room_name"`
	Day          int    `json:"day"`
	Hour         int    `json:"hour"`
	Duration     int    `json:"duration"`
}

func CreateRoomReservation(roomreservation *RoomReservation) error {
	err := database.DBConn.Create(&roomreservation).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return err
	}

	return nil

}

func GetReservationByID(id uint) (*RoomReservation, error) {
	reservation := RoomReservation{}
	err := database.DBConn.Find(&reservation, id).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return nil, err
	}
	return &reservation, nil
}
