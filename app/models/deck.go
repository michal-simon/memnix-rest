package models

import (
	"gorm.io/gorm"
	"memnixrest/pkg/database"
)

// Deck structure
type Deck struct {
	gorm.Model  `swaggerignore:"true"`
	DeckName    string     `json:"deck_name" example:"First Deck"`
	Description string     `json:"deck_description" example:"A simple demo deck"`
	Banner      string     `json:"deck_banner" example:"A banner url"`
	Status      DeckStatus `json:"deck_status" example:"2"` // 1: Draft - 2: Private - 3: Published
}

type DeckStatus int64

const (
	DeckDraft DeckStatus = iota + 1
	DeckPrivate
	DeckPublic
)

func (s DeckStatus) ToString() string {
	switch s {
	case DeckDraft:
		return "Deck Draft"
	case DeckPrivate:
		return "Deck Private"
	case DeckPublic:
		return "Deck Public"
	default:
		return "Unknown"
	}
}

func (deck *Deck) GetOwner() User {
	db := database.DBConn

	access := new(Access)

	if err := db.Joins("User").Joins("Deck").Where("accesses.deck_id =? AND accesses.permission >= ?", deck.ID, AccessOwner).Find(&access).Error; err != nil {
		return access.User
	}

	return access.User
}
