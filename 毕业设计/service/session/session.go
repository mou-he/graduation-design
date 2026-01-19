package session

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/mou-he/graduation-design/common/aihelper"
	"github.com/mou-he/graduation-design/common/code"
	"github.com/mou-he/graduation-design/dao/session"
	"github.com/mou-he/graduation-design/model"
)

var ctx = context.Background()

func GetUserSessionsByUserName(username string) ([]model.SessionInfo, error) {
	// 获取用户所有的会话ID
	manager := aihelper.GetGlobalManager()
	Sessions := manager.GetUserSessions(username)
	var SessionInfos []model.SessionInfo
	for _, session := range Sessions {
		SessionInfos = append(SessionInfos, model.SessionInfo{
			SessionID: session,
			Title:     session, // 暂时用sessionID作为标题，后续重构需要的时候可以更改
		})
	}
	// 返回会话信息列表
	return SessionInfos, nil
}

func CreateSessionAndSendMessage(userName string, userQuestion string, modelType string) (string, string, code.Code) {
	// 创建新会话
	newSession := &model.Session{
		ID:       uuid.New().String(),
		UserName: userName,
		Title:    userQuestion,
	}

	createSession, err := session.CreateSession(newSession)
	if err != nil {
		return "", "", code.CodeServerBusy
	}
	// 获取AIHelper并通过其管理信息
	manager := aihelper.GetGlobalManager()
	config := map[string]interface{}{
		"apiKey": "your-api-key", // TODO: 从配置中获取
	}
	// 创建AIHelper
	helper, err := manager.GetOrCreateAIHelper(userName, createSession.ID, modelType, config)
	if err != nil {
		log.Println("CreateSessionAndSendMessage GetOrCreateAIHelper error:", err)
		return "", "", code.AIModelFail
	}

	// 生成AI回复
	aiResponse, err := helper.GenerateResponse(userName, ctx, userQuestion)
	if err != nil {
		log.Println("CreateSessionAndSendMessage GenerateResponse error:", err)
		return "", "", code.AIModelFail
	}
	return createSession.ID, aiResponse.Content, code.CodeSuccess
}

// 创建流式会话，不发送消息
func CreateStreamSessionOnly(userName string, userQuestion string) (string, code.Code) {
	newSession := &model.Session{
		ID:       uuid.New().String(),
		UserName: userName,
		Title:    userQuestion,
	}
	createdSession, err := session.CreateSession(newSession)
	if err != nil {
		log.Println("CreateStreamSessionOnly CreateSession error:", err)
		return "", code.CodeServerBusy
	}
	return createdSession.ID, code.CodeSuccess
}

// 流式会话发送消息
func StreamMessageToExistingSession(userName string, sessionID string, userQuestion string, modelType string, writer http.ResponseWriter) code.Code {
	// 确保 writer 支持 Flush
	flusher, ok := writer.(http.Flusher)
	if !ok {
		log.Println("StreamMessageToExistingSession: streaming unsupported")
		return code.CodeServerBusy
	}

	manager := aihelper.GetGlobalManager()
	config := map[string]interface{}{
		"apiKey": "your-api-key", // TODO: 从配置中获取
	}
	helper, err := manager.GetOrCreateAIHelper(userName, sessionID, modelType, config)
	if err != nil {
		log.Println("StreamMessageToExistingSession GetOrCreateAIHelper error:", err)
		return code.AIModelFail
	}

	cb := func(msg string) {
		// 直接发送数据，不转义
		// SSE 格式：data: <content>\n\n
		log.Printf("[SSE] Sending chunk: %s (len=%d)\n", msg, len(msg))
		_, err := writer.Write([]byte("data: " + msg + "\n\n"))
		if err != nil {
			log.Println("[SSE] Write error:", err)
			return
		}
		flusher.Flush() //  每次必须 flush
		log.Println("[SSE] Flushed")
	}

	_, err_ := helper.StreamResponse(userName, ctx, cb, userQuestion)
	if err_ != nil {
		log.Println("StreamMessageToExistingSession StreamResponse error:", err_)
		return code.AIModelFail
	}

	_, err = writer.Write([]byte("data: [DONE]\n\n"))
	if err != nil {
		log.Println("StreamMessageToExistingSession write DONE error:", err)
		return code.AIModelFail
	}
	flusher.Flush()

	return code.CodeSuccess
}

// 创建流式会话并发送消息
func CreateStreamSessionAndSendMessage(userName string, userQuestion string, modelType string, writer http.ResponseWriter) (string, code.Code) {

	sessionID, code_ := CreateStreamSessionOnly(userName, userQuestion)
	if code_ != code.CodeSuccess {
		return "", code_
	}

	code_ = StreamMessageToExistingSession(userName, sessionID, userQuestion, modelType, writer)
	if code_ != code.CodeSuccess {

		return sessionID, code_
	}

	return sessionID, code.CodeSuccess
}
func ChatSend(userName, sessionID, userQuestion, modelType string) (string, code.Code) {
	manager := aihelper.GetGlobalManager()
	config := map[string]interface{}{
		"apiKey": "your-api-key", // TODO: 从配置中获取
	}
	helper, err := manager.GetOrCreateAIHelper(userName, sessionID, modelType, config)
	if err != nil {
		log.Println("ChatSend GetOrCreateAIHelper error:", err)
		return "", code.AIModelFail
	}
	//2：生成AI回复
	aiResponse, err_ := helper.GenerateResponse(userName, ctx, userQuestion)
	if err_ != nil {
		log.Println("ChatSend GenerateResponse error:", err_)
		return "", code.AIModelFail
	}

	return aiResponse.Content, code.CodeSuccess

}

func GetChatHistory(userName string, sessionID string) ([]model.History, code.Code) {
	// 获取AIHelper中的消息历史
	manager := aihelper.GetGlobalManager()
	helper, exists := manager.GetAIHelper(userName, sessionID)
	if !exists {
		return nil, code.CodeServerBusy
	}

	messages := helper.GetMessages()
	history := make([]model.History, 0, len(messages))

	// 转换消息为历史格式（根据消息顺序或内容判断用户/AI消息）
	for i, msg := range messages {
		isUser := i%2 == 0
		history = append(history, model.History{
			IsUser:  isUser,
			Content: msg.Content,
		})
	}

	return history, code.CodeSuccess
}
func ChatStreamSend(userName string, sessionID string, userQuestion string, modelType string, writer http.ResponseWriter) code.Code {

	return StreamMessageToExistingSession(userName, sessionID, userQuestion, modelType, writer)
}
