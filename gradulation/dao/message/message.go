package message

import (
	"errors"

	mysql "github.com/mou-he/graduation-design/common/mysql"
	model "github.com/mou-he/graduation-design/model"
	"gorm.io/gorm"
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

// 删除指定会话ID的所有消息（带事务）
func DeleteBySessionIDWithTx(tx *gorm.DB, sessionID string) error {
	// 前置校验
	if tx == nil {
		return errors.New("事务对象不能为空")
	}
	if sessionID == "" {
		return errors.New("sessionID不能为空")
	}

	// 执行批量删除（利用session_id索引加速）
	result := tx.Where("session_id = ?", sessionID).Delete(&model.Message{})
	if result.Error != nil {
		return errors.New("删除消息失败：" + result.Error.Error())
	}

	// 可选：打印删除行数（便于调试）
	// log.Printf("删除会话[%s]的消息共%d条", sessionID, result.RowsAffected)
	return nil
}
