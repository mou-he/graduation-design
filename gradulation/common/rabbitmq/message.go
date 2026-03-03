package rabbitmq

import (
	"encoding/json"
	"errors"

	"github.com/mou-he/graduation-design/dao/message"
	"github.com/mou-he/graduation-design/dao/session"
	"github.com/mou-he/graduation-design/model"
	"github.com/streadway/amqp"
)

type MessageMQParam struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
	UserName  string `json:"user_name"`
	IsUser    bool   `json:"is_user"`
}
type DeleteSessionMsg struct {
	SessionID string `json:"session_id"` // 会话ID
	UserName  string `json:"user_name"`  // 用户名（关联AIHelper）

}

func GenerateMessageMQParam(sessionID string, content string, userName string, IsUser bool) []byte {
	msg := &MessageMQParam{
		SessionID: sessionID,
		Content:   content,
		UserName:  userName,
		IsUser:    IsUser,
	}
	// 序列化为JSON
	data, _ := json.Marshal(msg)
	return data
}

func MQMessage(msg *amqp.Delivery) error {
	var param MessageMQParam
	err := json.Unmarshal(msg.Body, &param)
	if err != nil {
		return err
	}
	newMsg := &model.Message{
		SessionID: param.SessionID,
		Content:   param.Content,
		UserName:  param.UserName,
		IsUser:    param.IsUser,
	}
	//消费者异步插入到数据库中
	message.CreateMessage(newMsg)
	return nil
}

func MQDeleteSession(msg *amqp.Delivery) error {
	// 1. 反序列化消息（复用你MQMessage的反序列化逻辑）
	var deleteMsg DeleteSessionMsg
	err := json.Unmarshal(msg.Body, &deleteMsg)
	if err != nil {
		return err
	}

	// 2. 校验必要参数（和你MQMessage的参数校验逻辑一致）
	if deleteMsg.SessionID == "" {
		return errors.New("session_id不能为空")
	}
	// 3. 核心：删除会话+关联消息（复用DAO层，保证事务原子性）
	err = session.DeleteSessionAndMessages(deleteMsg.SessionID)

	if err != nil {
		return errors.New("删除会话及关联消息失败")
	}
	return nil
}
