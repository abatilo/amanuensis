package db

import "gorm.io/gorm"

type ValidatedStatus string

const (
	NotValidated ValidatedStatus = "not_validated"
	Valid        ValidatedStatus = "valid"
	Invalid      ValidatedStatus = "invalid"
)

type Validated struct {
	Status ValidatedStatus `gorm:"default:not_validated"`
	gorm.Model
}
