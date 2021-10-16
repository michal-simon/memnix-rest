package models

import (
	"gorm.io/gorm"
)

// Access structure
type Answer struct {
	gorm.Model
	CardID uint `json:"card_id" example:"1"`
	Card   Card
	Answer string `json:"answer" example:"1"`
}