package models

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
	Day      int `json:"day" validate:"required,min=1,max=7"`
	Hour     int `json:"hour" validate:"required,min=9,max=17"`
	Duration int `json:"duration" validate:"required"`
}

// here are the validation "methods" for these models

func (room *Room) Validate() error {

	return nil
}

func (reservation *Reservation) Validate() error {

	if 
	return nil
}