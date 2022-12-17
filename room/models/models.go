package models

import (
	"SocketProgramming/room/database"
	"errors"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// A room only has a unique name on its own
// how-ever it also has a "Has Many" relationship with the Reservations
type Room struct {
	Name         string        `gorm:"primaryKey" json:"room_name"`
	Reservations []Reservation `gorm:"foreignKey:RoomName;References:Name" json:"reservations"`
}

// I have use an external packet just for validating the structs (will acts as table entries)
type Reservation struct {
	ID       uint `gorm:"primaryKey:auto_increment" json:"id"`
	RoomName string
	Day      int `json:"day"`
	Hour     int `json:"hour"`
	Duration int `json:"duration"`
}

// here are the validation "methods" for these models

// There is no constraint for room creation, only error would be the situation where user
// tries to create a room with an existing name, and we won't consider that case in here.
func (room *Room) Validate() (err error) {

	if room.Name == "" {
		err = errors.New("room name cannot be empty")
	}
	return
}

func (reservation *Reservation) Validate() (errs []error) {

	if reservation.Day < 1 || reservation.Day > 7 {
		errs = append(errs, errors.New("day value of reservation should be 1 to 7"))
	}

	if reservation.Hour < 9 || reservation.Hour > 17 {
		errs = append(errs, errors.New("hour value of reservation should be 9 to 17"))
	}
	return nil
}

// Database methods

func (room *Room) BeforeCreate(tx *gorm.DB) (err error) {

	err = room.Validate()

	return
}

func (reservation *Reservation) BeforeCreate(tx *gorm.DB) (err error) {

	errs := reservation.Validate()
	if len(errs) == 0 {
		return nil
	}
	// concatenate error's strings
	builder := strings.Builder{}
	for _, e := range errs {
		builder.WriteString(e.Error())
		builder.WriteString(fmt.Sprintf("\n"))
	}

	return errors.New(builder.String())
}

func CreateRoom(room *Room) error {
	err := database.DBConn.Create(&room).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return err
	}
	return nil
}

func CreateReservation(reservation *Reservation) error {
	err := database.DBConn.Create(&reservation).Error
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

func GetAllReservations() ([]Reservation, error) {
	reservations := []Reservation{}

	err := database.DBConn.Find(&reservations).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return nil, err
	}
	return reservations, nil
}