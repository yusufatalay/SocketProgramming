package models

import (
	"SocketProgramming/room/database"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// I have use an external packet just for validating the structs (will acts as table entries)
type Reservation struct {
	ID       uint   `gorm:"primaryKey:auto_increment" json:"id"`
	RoomName string `json:"room_name"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Duration int    `json:"duration"`
}

// here are the validation "methods" for these models

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
func CreateReservation(reservation *Reservation) error {
	err := database.DBConn.Create(&reservation).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return err
	}
	return nil
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

// GetAvailableHours returns list of hours for the given room if succesfful
// Returns error if room not found , returns empty list if there is no available hours
func GetAvailableHours(roomname string, day int) ([]string, error) {

	availableHours := []string{"9", "10", "11", "12", "13", "14", "15", "16", "17"}
	// get reservations with the given name and day
	reservations := []Reservation{}

	err := database.DBConn.Where("room_name = ? AND day = ?",
		roomname, day).Find(&reservations).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return nil, err
	}

	// for every reservation:
	// block starting from it's hour, till it's hour+duration
	// rest is available
	for _, res := range reservations {
		hourIndex := getHourIndex(availableHours, strconv.Itoa(res.Hour))
		availableHours = append(availableHours[:hourIndex], availableHours[hourIndex+res.Duration:]...)
	}
	return availableHours, nil
}

func getHourIndex(hourlist []string, hour string) int {

	for i, v := range hourlist {
		if v == hour {
			return i
		}
	}
	return -1
}
