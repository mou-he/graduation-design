package aihelper

import (
	"context"
	"sync"
)

type AIHelperManager struct {
	helpers map[string]map[string]*AIHelper
	mu      sync.RWMutex
}

func (m *AIHelperManager) GetAIHelper(userName string, sessionID string) (*AIHelper, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return nil, false
	}

	helper, exists := userHelpers[sessionID]
	return helper, exists
}

func NewAIHelperManager() *AIHelperManager {
	return &AIHelperManager{
		helpers: make(map[string]map[string]*AIHelper),
	}
}
func (m *AIHelperManager) GetOrCreateAIHelper(userName, sessionID, modelType string, config map[string]interface{}) (*AIHelper, error) {
	// 检查是否存在
	// 加锁确保并发安全
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取用户的会话映射
	userHelpers, exists := m.helpers[userName]
	if !exists {
		userHelpers = make(map[string]*AIHelper)
		m.helpers[userName] = userHelpers
	}

	// 检查会话是否已存在
	helper, exists := userHelpers[sessionID]
	if exists {
		return helper, nil
	}
	if exists {
		return helper, nil
	}

	// 创建新的AIHelper
	// 获取全局工厂实例
	factory := GetGlobalFactory()
	aihelper, err := factory.CreateAIHelper(context.Background(), modelType, sessionID, config)
	if err != nil {
		return nil, err
	}
	helper = aihelper
	m.helpers[userName][sessionID] = helper
	return helper, nil
}
func (m *AIHelperManager) RemoveAIHelper(userName, sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	userHelpers, exists := m.helpers[userName]
	if !exists {
		return
	}
	delete(userHelpers, sessionID)
	// 如果用户没有会话了，清理用户的会话映射
	if len(userHelpers) == 0 {
		delete(m.helpers, userName)
	}
}
func (m *AIHelperManager) GetUserSessions(userName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	userHelpers, exists := m.helpers[userName]
	if !exists {
		return nil
	}
	sessions := make([]string, 0, len(userHelpers))
	for sessionID := range userHelpers {
		sessions = append(sessions, sessionID)
	}
	return sessions
}

var globalManager *AIHelperManager
var once sync.Once

func GetGlobalManager() *AIHelperManager {
	once.Do(func() {
		globalManager = NewAIHelperManager()
	})
	return globalManager
}
