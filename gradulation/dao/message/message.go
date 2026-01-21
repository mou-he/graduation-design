package message

import (
	mysql "github.com/mou-he/graduation-design/common/mysql"
	model "github.com/mou-he/graduation-design/model"
)

// 根据会话ID获取消息
func GetMessageBySessionID(sessionID string) ([]model.Message, error) {
	var msgs []model.Message
	err := mysql.DB.Where("session_id = ?", sessionID).Order("created_at").Find(&msgs).Error
	return msgs, err
}

// 根据会话ID列表获取消息
func GetMessageBySessionIDs(sessionIDs []string) ([]model.Message, error) {
	var msgs []model.Message
	err := mysql.DB.Where("session_id in ?", sessionIDs).Order("created_at").Find(&msgs).Error
	return msgs, err
}

func CreateMessage(message *model.Message) (*model.Message, error) {
	err := mysql.DB.Create(message).Error
	return message, err
}

// 获取所有消息
func GetAllMessages() ([]model.Message, error) {
	var msgs []model.Message
	err := mysql.DB.Order("created_at asc").Find(&msgs).Error
	return msgs, err
}
