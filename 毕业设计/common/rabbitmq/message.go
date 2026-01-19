package rabbitmq

import (
	"encoding/json"
	"github.com/mou-he/graduation-design/dao/message"
	"github.com/mou-he/graduation-design/model"
	"github.com/streadway/amqp"
)

type MessageMQParam struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
	UserName  string `json:"user_name"`
	IsUser    bool   `json:"is_user"`
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
