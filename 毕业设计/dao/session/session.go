package session

import (
	mysql "github.com/mou-he/graduation-design/common/mysql"
	model "github.com/mou-he/graduation-design/model"
)

// Username如果之后添加修改功能的时候可能要改成string类型
func GetSessionByUserName(UserName string) ([]model.SessionInfo, error) {
	var sessions []model.SessionInfo
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
