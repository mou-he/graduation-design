package model

import (
	"time"
)

type Message struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID string    `gorm:"index;not null type:char(36)" json:"sessionId"`
	Content   string    `gorm:"type:text" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	IsUser    bool      `gorm:"not null" json:"isUser"`
}
type History struct {
	IsUser  bool   `json:"isUser"`
	Content string `json:"content"`
}
