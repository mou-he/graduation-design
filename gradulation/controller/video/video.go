package controller

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/common/code"
	redisCli "github.com/mou-he/graduation-design/common/redis"
	"github.com/mou-he/graduation-design/controller"
	"github.com/mou-he/graduation-design/model"
	ai_video "github.com/mou-he/graduation-design/service/ai_video"
	"github.com/mou-he/graduation-design/service/video"
)

// ================= 结构体定义 =================

// Request 结构体
type GenerateRequest struct {
	Prompt    string `json:"prompt" binding:"required"`
	Duration  int    `json:"duration"`   // 前端传 5 或 10
	Quality   string `json:"quality"`    // "quality" 或 "speed"
	WithAudio bool   `json:"with_audio"` // true 或 false
}

// Response (保持不变)
type GenerateResponse struct {
	TaskID string `json:"task_id,omitempty"`
	controller.Response
}

type VideoStatusResp struct {
	Data *model.VideoTaskState `json:"data,omitempty"`
	controller.Response
}

// ================= 处理函数 =================

// GenerateVideo 处理视频生成请求
func GenerateVideo(c *gin.Context) {
	res := new(GenerateResponse)
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	// 默认值处理（如果前端没传）
	if req.Duration != 10 {
		req.Duration = 5
	} // 只能是 5 或 10
	if req.Quality == "" {
		req.Quality = "quality"
	}

	// 调用 Service，传入所有参数
	taskID, err := video.SubmitVideoRequest(
		req.Prompt,
		req.Duration,
		req.Quality,
		req.WithAudio,
	)

	if err != nil {
		log.Println("Submit error:", err)
		c.JSON(http.StatusOK, res.CodeOf(code.VideoSubmitFail))
		return
	}

	res.Success()
	res.TaskID = taskID
	c.JSON(http.StatusOK, res)
}

// GetVideoStatus 查询视频生成状态
func GetVideoStatus(c *gin.Context) {
	res := new(VideoStatusResp)
	taskID := c.Param("id")

	if taskID == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	state, err := video.GetVideoStatus(taskID)
	if err != nil {
		log.Println("Get status fail", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	if state == nil {
		c.JSON(http.StatusOK, res.CodeOf(code.VideoTaskNotFound))
		return
	}

	res.Success()
	res.Data = state
	c.JSON(http.StatusOK, res)
}

// video_controller.go（路由处理层）

// ProxyVideoPlay 视频播放代理（控制器层）
func ProxyVideoPlay(c *gin.Context) {
	// 1. 接收请求参数
	localTaskID := c.Param("id")
	if localTaskID == "" {
		c.JSON(400, gin.H{
			"status_code": 400,
			"status_msg":  "任务ID不能为空",
		})
		return
	}

	// 2. 数据层：查询视频信息（Redis）
	taskState, err := redisCli.GetVideoTaskState(localTaskID)
	if err != nil || taskState.VideoURL == "" {
		c.JSON(404, gin.H{
			"status_code": 404,
			"status_msg":  "视频不存在或未生成完成",
		})
		return
	}

	// 3. 调用业务层核心逻辑
	resp, err := ai_video.VideoPlayProxy(localTaskID, taskState.VideoURL)
	if err != nil {
		c.JSON(500, gin.H{
			"status_code": 500,
			"status_msg":  fmt.Sprintf("播放代理失败: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	// 4. 透传响应给前端（控制器层只做转发）
	// 设置跨域和播放相关头
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Range")
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Accept-Ranges", resp.Header.Get("Accept-Ranges"))
	c.Header("Content-Length", resp.Header.Get("Content-Length"))
	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		c.Header("Content-Range", contentRange)
	}

	// 5. 流式转发视频流
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
