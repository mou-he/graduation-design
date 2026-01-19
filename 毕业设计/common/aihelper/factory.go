package aihelper

import (
	"context"
	"fmt"
	"sync"
)

// 定义模型创建函数类型
type ModelCreator func(ctx context.Context, config map[string]interface{}) (AIModel, error)

// AI模型工厂
type AIModelFactory struct {
	creators map[string]ModelCreator
	mu       sync.RWMutex
}

// 全局工厂实例
var (
	globalFactory *AIModelFactory
	factoryOnce   sync.Once
)

func GetGlobalFactory() *AIModelFactory {
	factoryOnce.Do(func() {
		globalFactory = &AIModelFactory{
			creators: make(map[string]ModelCreator),
		}
		globalFactory.registerCreators()
	})
	return globalFactory
}
func (f *AIModelFactory) registerCreators() {
	// openai
	f.creators["1"] = func(ctx context.Context, config map[string]interface{}) (AIModel, error) {
		return NewOpenAIModel(ctx)
	}
	f.creators["2"] = func(ctx context.Context, config map[string]interface{}) (AIModel, error) {
		baseURL, ok := config["baseURL"].(string)
		if !ok {
			return nil, fmt.Errorf("baseURL is required")
		}
		modelName, ok := config["modelName"].(string)
		if !ok {
			return nil, fmt.Errorf("modelName is required")
		}
		return NewOllamaModel(ctx, baseURL, modelName)
	}
}

func (f *AIModelFactory) CreateAIModel(ctx context.Context, modeType string, config map[string]interface{}) (AIModel, error) {
	f.mu.RLock()
	creator, exists := f.creators[modeType]
	defer f.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("model type %s not registered", modeType)
	}
	return creator(ctx, config)
}

// 注册创建者
func (f *AIModelFactory) RegisterModel(modelType string, creator ModelCreator) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[modelType] = creator
}

// 创建AI助手
func (f *AIModelFactory) CreateAIHelper(ctx context.Context, modelType string, SessionID string, config map[string]interface{}) (*AIHelper, error) {
	model, err := f.CreateAIModel(ctx, modelType, config)
	if err != nil {
		return nil, err
	}
	return NewAIHelper(model, SessionID), nil
}
