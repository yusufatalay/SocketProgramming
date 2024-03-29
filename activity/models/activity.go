package models

import (
	"errors"
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
	if activity.Name == "" {
		err = errors.New("activity name cannot be empty")
	}
	return
}

func (activity *Activity) BeforeCreate(tx *gorm.DB) (err error) {
	err = activity.Validate()

	var exists bool
	err = database.DBConn.Model(Activity{}).Select("count(*) > 0").Where("name = ?", activity.Name).Find(&exists).Error
	if err != nil {
		log.Printf("Error %+v", err)
		return err
	}
	if exists {
		return errors.New("Activity already exists")
	}
	return nil

}

func CreateActivity(activity *Activity) error {
	err := database.DBConn.Create(&activity).Error
	if err != nil {
		return err
	}
	return nil
}

func RemoveActivity(name string) error {

	var exists bool
	err := database.DBConn.Model(Activity{}).Select("count(*) > 0").Where("name = ?", name).Find(&exists).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return err
	}
	if !exists {
		return errors.New("activity does not exists")
	}

	err = database.DBConn.Delete(&Activity{Name: name}).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return err
	}
	return nil

}

func CheckActivity(name string) (bool, error) {

	var exists bool
	err := database.DBConn.Model(Activity{}).Select("count(*) > 0").Where("name = ?", name).Find(&exists).Error
	if err != nil {
		log.Printf("Error: %+v", err)
		return false, err
	}

	return exists, nil
}