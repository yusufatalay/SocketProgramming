package models

import (
	"errors"
	"fmt"
	"log"

	"github.com/yusufatalay/SocketProgramming/activity/database"

	"gorm.io/gorm"
)

// A room only has a unique name on its own
// how-ever it also has a "Has Many" relationship with the Reservations
type Activity struct {
	Name string `gorm:"primaryKey" json:"activity_name"`
}

func (activity *Activity) Validate() (err error) {

	return nil
}
