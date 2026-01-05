package model

import (
	"gorm.io/gorm"
	"time"
)

type Session struct {
	ID        string         `gorm:"primaryKey;type:char(36)" json:"id"`
	UserName  string         `gorm:"index;not null" json:"username"`
	Title     string         `gorm:"type:varchar(100)" json:"title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 支持软删除
}
type SessionInfo struct {
	SessionID string `json:"sessionId"`
	Title     string `json:"title"`
}
