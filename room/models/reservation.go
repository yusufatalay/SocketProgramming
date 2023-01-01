package models

import (
	"errors"
	"log"
	"strconv"

	"github.com/yusufatalay/SocketProgramming/room/database"
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

func (reservation *Reservation) Validate() error {

	// check if room exists in database
	var exists bool
	err := database.DBConn.Model(Room{}).Select("count(*) > 0").Where("name = ?", reservation.RoomName).Find(&exists).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return err
	}
	if !exists {
		return errors.New("Room does not exists")
	}

	if reservation.Day < 1 || reservation.Day > 7 {
		return errors.New("day value of reservation should be 1 to 7")
	}

	if reservation.Hour < 9 || reservation.Hour > 17 {
		return errors.New("hour value of reservation should be 9 to 17")
	}

	if reservation.Hour+reservation.Duration > 17 {
		return errors.New("Invalid reservation time slice")
	}
	return nil
}

// Database methods

func CreateReservation(reservation *Reservation) error {
	// first check for constraints
	err := reservation.Validate()
	if err != nil {
		return err
	}

	// now check if this reservation already made in db
	hours, err := GetAvailableHours(reservation.RoomName, reservation.Day)
	if err != nil {
		return errors.New("Could not get available hours: " + err.Error())
	}
	for i := reservation.Hour; i < reservation.Hour+reservation.Duration; i++ {
		if !contains(hours, i) {
			return errors.New("Already Reserved")
		}
	}

	err = database.DBConn.Create(&reservation).Error
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
	var exists bool
	err := database.DBConn.Model(Room{}).Select("count(*) > 0").Where("name = ?", roomname).Find(&exists).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return nil, err
	}
	if !exists {
		return nil, errors.New("Room does not exists")
	}
	availableHours := []string{"9", "10", "11", "12", "13", "14", "15", "16", "17"}
	// get reservations with the given name and day
	reservations := []Reservation{}

	err = database.DBConn.Where("room_name = ? AND day = ?",
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
		if hourIndex == -1 {
			return nil, errors.New("No available hours for reservation")
		}
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

func contains(hours []string, hour int) bool {

	for _, h := range hours {
		k, err := strconv.Atoi(h)
		if err != nil {
			return false
		}

		if k == hour {
			return true
		}
	}
	return false
}
