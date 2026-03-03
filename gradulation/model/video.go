package model

// VideoTaskState 存放在 Redis 中的状态
type VideoTaskState struct {
	Status    string `json:"status"`               // PENDING, PROCESSING, SUCCESS, FAIL
	VideoURL  string `json:"video_url,omitempty"`  // 成功后的视频链接
	CoverURL  string `json:"cover_url,omitempty"`  // 封面图
	FailError string `json:"fail_error,omitempty"` // 失败原因
	ZhipuID   string `json:"zhipu_id,omitempty"`   // 智谱的任务ID
}

// VideoMQParam MQ 消息体 (新增参数)
type VideoMQParam struct {
	LocalID string `json:"local_id"`
	Prompt  string `json:"prompt"`

	// 新增参数，对应 OpenAPI 文档
	Duration  int    `json:"duration,omitempty"`   // 5 或 10
	Quality   string `json:"quality,omitempty"`    // "quality" | "speed"
	WithAudio bool   `json:"with_audio,omitempty"` // true | false
	ZhipuID   string `json:"zhipu_id,omitempty"`
}
