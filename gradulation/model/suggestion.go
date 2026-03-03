package model

import "time"

// Suggestion 留言数据库模型
type Suggestion struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	RootID          uint64 `gorm:"not null;default:0" json:"root_id"`   // 根留言ID（主留言=自身ID，回复=主留言ID）
	ParentID        uint64 `gorm:"not null;default:0" json:"parent_id"` // 父级ID（主留言=0，回复=目标留言ID）
	UserID          string `gorm:"not null" json:"user_id"`             // 发布者ID
	Username        string `gorm:"not null" json:"username"`            // 发布者昵称
	ReplyToUserID   string `gorm:"default:''" json:"reply_to_user_id"`  // 被回复者ID
	ReplyToUsername string `gorm:"default:''" json:"reply_to_username"` // 被回复者昵称
	Content         string `gorm:"not null;size:500" json:"content"`    // 内容
	// 重点：删除LikeCount字段，新增ReplyCount字段（根评论专用）
	ReplyCount int       `gorm:"not null;default:0" json:"reply_count"` // 回复数（根评论专用）
	IsDeleted  int       `gorm:"not null;default:0" json:"is_deleted"`  // 是否删除（0-未删，1-已删）
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`     // 创建时间
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`     // 更新时间
}

// PublishRequest 发布主留言请求（Service层用）
type PublishRequest struct {
	Content string `json:"content"`
}

// UpdateRequest 更新留言请求（Service层用）
type UpdateRequest struct {
	ID      uint64 `json:"id"`
	Content string `json:"content"`
}

// ReplyRequest 回复留言请求（Service层用）
type ReplyRequest struct {
	ParentID        uint64 `json:"parent_id"`
	ReplyToUserID   string `json:"reply_to_user_id"`
	ReplyToUsername string `json:"reply_to_username"`
	Content         string `json:"content"`
}

// RootSuggestionVO 根留言展示VO（前端用）
type RootSuggestionVO struct {
	ID         uint64    `json:"id"`
	Content    string    `json:"content"`
	Username   string    `json:"username"`
	CreateTime time.Time `json:"create_time"`
	ReplyCount int64     `json:"reply_count"` // 下属回复数（删除LikeCount字段）
}

// Suggestion 留言数据库模型表名映射
func (Suggestion) TableName() string {
	return "suggestion_board"
}
