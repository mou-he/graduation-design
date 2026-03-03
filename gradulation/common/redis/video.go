package redis

import (
	"encoding/json"
	"time"

	"github.com/mou-he/graduation-design/model"
	redisCli "github.com/redis/go-redis/v9"
)

const (
	VideoTaskPrefix = "video_task:"
	TaskExpire      = 24 * time.Hour
)

// SetVideoTaskState 更新任务状态 (支持增量更新)
func SetVideoTaskState(localID string, newState model.VideoTaskState) error {
	// 2. 构建完整的 Redis 键
	key := VideoTaskPrefix + localID

	// 1. 尝试获取旧状态以保留字段 (如 ZhipuID)
	oldState, _ := GetVideoTaskState(localID)
	if oldState != nil {
		if newState.Status == "" {
			newState.Status = oldState.Status
		}
		if newState.VideoURL == "" {
			newState.VideoURL = oldState.VideoURL
		}
		if newState.ZhipuID == "" {
			newState.ZhipuID = oldState.ZhipuID
		}
		if newState.FailError == "" {
			newState.FailError = oldState.FailError
		}
	}

	data, err := json.Marshal(newState)
	if err != nil {
		return err
	}

	// 使用你包里全局的 Rdb 变量
	return Rdb.Set(ctx, key, data, TaskExpire).Err()
}

// GetVideoTaskState 获取任务详情
func GetVideoTaskState(localID string) (*model.VideoTaskState, error) {
	key := VideoTaskPrefix + localID
	val, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redisCli.Nil {
			return nil, nil
		}
		return nil, err
	}

	var state model.VideoTaskState
	err = json.Unmarshal([]byte(val), &state)
	return &state, err
}
