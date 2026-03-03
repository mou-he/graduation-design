package session

import (
	"errors"
	"log"

	mysql "github.com/mou-he/graduation-design/common/mysql"
	message "github.com/mou-he/graduation-design/dao/message"
	model "github.com/mou-he/graduation-design/model"
	"gorm.io/gorm"
)

// Username如果之后添加修改功能的时候可能要改成string类型
func GetSessionByUserName(UserName string) ([]model.Session, error) {
	var sessions []model.Session
	err := mysql.DB.Where("user_name = ?", UserName).Find(&sessions).Error
	return sessions, err
}

// 创建会话
func CreateSession(session *model.Session) (*model.Session, error) {
	err := mysql.DB.Create(session).Error
	return session, err
}

func GetSessionByID(sessionID string) (*model.Session, error) {
	var session model.Session
	err := mysql.DB.Where("id = ?", sessionID).First(&session).Error
	return &session, err
}

func TransDeleteSessionAndMessages(sessionID string) error {
	// 前置校验：避免空参数
	if sessionID == "" {
		return errors.New("sessionID不能为空")
	}

	// 校验会话是否存在（避免删除不存在的会话）
	_, err := GetSessionByID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("会话不存在，无需删除")
		}
		return errors.New("查询会话失败：" + err.Error())
	}

	// 启动GORM事务
	err = mysql.DB.Transaction(func(tx *gorm.DB) error {
		// 步骤1：调用message DAO层的事务方法删除关联消息
		// 需在message DAO层实现DeleteBySessionIDWithTx方法（见下方）
		if err := message.DeleteBySessionIDWithTx(tx, sessionID); err != nil {
			log.Printf("事务内删除消息失败 [sessionID:%s]：%v", sessionID, err)
			return err // 返回错误触发事务回滚
		}

		// 步骤2：删除会话本身（使用session_id字段匹配）
		result := tx.Where("id = ?", sessionID).Delete(&model.Session{})
		if result.Error != nil {
			log.Printf("事务内删除会话失败 [sessionID:%s]：%v", sessionID, result.Error)
			return result.Error // 返回错误触发事务回滚
		}

		// 校验是否真的删除了会话（防止未匹配到数据）
		if result.RowsAffected == 0 {
			return errors.New("会话删除失败：未匹配到数据")
		}
		return nil
	})
	return err
}
func DeleteSessionAndMessages(sessionID string) error {
	// 直接调用事务方法
	err := TransDeleteSessionAndMessages(sessionID)
	if err != nil {
		return errors.New("删除会话及消息失败：" + err.Error())
	}
	return nil
}
