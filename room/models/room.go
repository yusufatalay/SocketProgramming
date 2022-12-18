package models

import (
	"SocketProgramming/room/database"
	"errors"
	"log"

	"gorm.io/gorm"
)

// A room only has a unique name on its own
// how-ever it also has a "Has Many" relationship with the Reservations
type Room struct {
	Name         string        `gorm:"primaryKey" json:"room_name"`
	Reservations []Reservation `gorm:"foreignKey:RoomName;References:Name" json:"reservations"`
}

// There is no constraint for room creation, only error would be the situation where user
// tries to create a room with an existing name, and we won't consider that case in here.
func (room *Room) Validate() (err error) {

	if room.Name == "" {
		err = errors.New("room name cannot be empty")
	}
	return
}

func (room *Room) BeforeCreate(tx *gorm.DB) (err error) {

	err = room.Validate()

	return
}

func CreateRoom(room *Room) error {
	err := database.DBConn.Create(&room).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return err
	}
	return nil
}

// RemoveRoom permanently removes a room.
// Instead of making soft delete, to lower the complexity.
func RemoveRoom(name string) error {

	err := database.DBConn.Unscoped().Delete(&Room{}, name).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return err
	}
	return nil
}

func GetRoom(name string) (*Room, error) {
	room := &Room{}

	err := database.DBConn.First(&room, name).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return nil, err
	}
	return room, nil
}

func GetAllRooms() ([]Room, error) {
	rooms := []Room{}

	err := database.DBConn.Find(&rooms).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return nil, err
	}
	return rooms, nil
}
