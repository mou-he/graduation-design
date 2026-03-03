package rabbitmq

import (
	"encoding/json"
	"log"
	"time"

	"github.com/mou-he/graduation-design/common/redis"           // 假设你的redis包路径
	"github.com/mou-he/graduation-design/model"                  // 模型层
	video "github.com/mou-he/graduation-design/service/ai_video" // 视频服务层

	"github.com/streadway/amqp"
)

// ================== 数据结构定义 ==================

// VideoSubmitParam 提交任务的参数
type VideoSubmitParam struct {
	LocalID string `json:"local_id"` // 本地UUID
	Prompt  string `json:"prompt"`   // 提示词
}

// VideoCheckParam 检查任务的参数
type VideoCheckParam struct {
	LocalID string `json:"local_id"`
	ZhipuID string `json:"zhipu_id"` // 智谱返回的任务ID
}

// ================== 消费者回调逻辑 ==================

// MQVideoSubmit 消费者：处理提交任务
func MQVideoSubmit(msg *amqp.Delivery) error {
	var param model.VideoMQParam
	if err := json.Unmarshal(msg.Body, &param); err != nil {
		return err
	}

	log.Printf("[MQ Submit] Prompt: %s, Duration: %d", param.Prompt, param.Duration)
	redis.SetVideoTaskState(param.LocalID, model.VideoTaskState{Status: "PROCESSING"})

	// 1. 组装 AI 调用参数
	opts := video.SubmitOptions{
		Prompt:    param.Prompt,
		Duration:  param.Duration, // 传入MQ带过来的时长
		Quality:   param.Quality,  // 传入MQ带过来的画质
		WithAudio: param.WithAudio,
	}

	// 2. 调用智谱
	zhipuService := video.NewZhipuService()
	zhipuID, err := zhipuService.SubmitTask(opts)

	if err != nil {
		log.Printf("[MQ Submit] Fail: %v", err)
		redis.SetVideoTaskState(param.LocalID, model.VideoTaskState{Status: "FAIL", FailError: err.Error()})
		return nil
	}

	// 3. 成功流程 (保持不变)
	redis.SetVideoTaskState(param.LocalID, model.VideoTaskState{Status: "PROCESSING", ZhipuID: zhipuID})

	checkParam := model.VideoMQParam{LocalID: param.LocalID, ZhipuID: zhipuID}
	checkData, _ := json.Marshal(checkParam)
	return RMQVideoCheck.Publish(checkData)
}

// MQVideoCheck 消费：轮询检查进度
func MQVideoCheck(msg *amqp.Delivery) error {
	var param model.VideoMQParam
	if err := json.Unmarshal(msg.Body, &param); err != nil {
		return err
	}

	zhipuService := video.NewZhipuService()
	status, videoURL, err := zhipuService.QueryResult(param.ZhipuID)

	// 网络错误重试
	if err != nil {
		go delayedRepublish(param, 5*time.Second)
		return nil
	}

	if status == "SUCCESS" {
		log.Printf("[MQ Check] Success: %s", param.LocalID)
		redis.SetVideoTaskState(param.LocalID, model.VideoTaskState{Status: "SUCCESS", VideoURL: videoURL})
	} else if status == "FAIL" {
		log.Printf("[MQ Check] Fail: %s", param.LocalID)
		redis.SetVideoTaskState(param.LocalID, model.VideoTaskState{Status: "FAIL", FailError: "Generation Failed"})
	} else {
		// 生成中 -> 延迟10秒后重发到队列
		// log.Printf("[MQ Check] Processing: %s, retry in 10s", param.LocalID)
		go delayedRepublish(param, 10*time.Second)
	}
	return nil
}

func delayedRepublish(param model.VideoMQParam, delay time.Duration) {
	time.Sleep(delay)
	data, _ := json.Marshal(param)
	if RMQVideoCheck != nil {
		RMQVideoCheck.Publish(data)
	}
}
