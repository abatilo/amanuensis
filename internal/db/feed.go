package db

import "gorm.io/gorm"

type Feed struct {
	URL   string          `gorm:"unique"`
	Valid ValidatedStatus `gorm:"default:not_validated;not null"`
	gorm.Model
}
