package models

import (
	"gorm.io/gorm"
)

type Customer struct {
	ID          uint   `gorm:"primary key; autoIncrement" json:"id"`
	FirstName   string `json:"first_name" validate:"required,max=100"`
	LastName    string `json:"last_name" validate:"required,max=100"`
	DateOfBirth string `json:"date_of_birth" validate:"required"`
	Gender      string `json:"gender" validate:"required,oneof=Male Female"`
	Email       string `json:"e_mail" validate:"required,email"`
	Address     string `json:"address" validate:"omitempty,max=200"`
}

func MigrateCustomers(db *gorm.DB) error {
	err := db.AutoMigrate(&Customer{})
	return err
}
