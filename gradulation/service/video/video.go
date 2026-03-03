package video

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/mou-he/graduation-design/common/rabbitmq"
	"github.com/mou-he/graduation-design/common/redis"
	"github.com/mou-he/graduation-design/model"
)

// SubmitVideoRequest 处理用户提交的生成请求
func SubmitVideoRequest(prompt string, duration int, quality string, withAudio bool) (string, error) {
	localID := uuid.New().String()

	// 1. Redis
	err := redis.SetVideoTaskState(localID, model.VideoTaskState{Status: "PENDING"})
	if err != nil {
		return "", err
	}

	// 2. MQ Param 赋值
	param := model.VideoMQParam{
		LocalID:   localID,
		Prompt:    prompt,
		Duration:  duration,  // 传递时长
		Quality:   quality,   // 传递画质
		WithAudio: withAudio, // 传递音频开关
	}
	data, _ := json.Marshal(param)

	// 3. Publish
	err = rabbitmq.RMQVideoSubmit.Publish(data)
	if err != nil {
		return "", err
	}

	return localID, nil
}

// GetVideoStatus 查询视频生成状态
func GetVideoStatus(localID string) (*model.VideoTaskState, error) {
	// 直接查 Redis
	return redis.GetVideoTaskState(localID)
}
