package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// ZhipuService 结构体
type ZhipuService struct {
	BaseURL string
	APIKey  string
}

// GenerateReq 请求参数
type CogVideoX3Request struct {
	Model     string `json:"model"` // 固定 cogvideox-3
	Prompt    string `json:"prompt"`
	Quality   string `json:"quality,omitempty"`  // speed / quality
	WithAudio bool   `json:"with_audio"`         // 文档默认 false，我们设为 true
	Duration  int    `json:"duration,omitempty"` // 5 / 10
	Size      string `json:"size,omitempty"`     // 1920x1080
	FPS       int    `json:"fps,omitempty"`      // 30 / 60
}

// SubmitOptions 提交参数
type SubmitOptions struct {
	Prompt    string
	Duration  int    // 视频时长 (5 或 10)
	Quality   string // 画质 (quality 或 speed)
	WithAudio bool   // 是否开启音效
}

// NewZhipuService 初始化方法
func NewZhipuService() *ZhipuService {
	return &ZhipuService{
		BaseURL: "https://open.bigmodel.cn/api/paas/v4/videos",
		APIKey:  os.Getenv("ZHIPU_API_KEY"),
	}
}

// SubmitTask 提交任务 (HTTP调用)
func (s *ZhipuService) SubmitTask(opts SubmitOptions) (string, error) {
	// 1. 校验 APIKey 格式
	token, err := s.generateToken()
	if err != nil {
		return "", err
	}

	// 1. 参数校验与默认值处理 (依据文档)
	duration := 5
	if opts.Duration == 10 {
		duration = 10
	}

	quality := "quality" // 默认高质量
	if opts.Quality == "speed" {
		quality = "speed"
	}

	// 2. 构造请求体
	reqBody := CogVideoX3Request{
		Model:     "cogvideox-3",
		Prompt:    opts.Prompt,
		Quality:   quality,
		WithAudio: opts.WithAudio, // 布尔值
		Duration:  duration,
		Size:      "1920x1080",
		FPS:       30,
	}

	jsonData, _ := json.Marshal(reqBody)

	// 3. 发送请求
	req, _ := http.NewRequest("POST", s.BaseURL+"/generations", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// 读取错误信息方便调试
		// buf := new(bytes.Buffer)
		// buf.ReadFrom(resp.Body)
		// fmt.Println(buf.String())
		return "", fmt.Errorf("api error code: %d", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ID, nil
}

// QueryResult 查询结果 (HTTP调用)
func (s *ZhipuService) QueryResult(taskID string) (status string, videoURL string, err error) {
	token, _ := s.generateToken()

	url := fmt.Sprintf("https://open.bigmodel.cn/api/paas/v4/async-result/%s", taskID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		TaskStatus  string `json:"task_status"`
		VideoResult []struct {
			URL      string `json:"url"`
			CoverURL string `json:"cover_image_url"`
		} `json:"video_result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	videoURL = ""
	if len(result.VideoResult) > 0 {
		videoURL = result.VideoResult[0].URL
	}

	return result.TaskStatus, videoURL, nil
}

// generateToken JWT生成逻辑
func (s *ZhipuService) generateToken() (string, error) {
	// 1. 校验 APIKey 格式
	parts := strings.Split(s.APIKey, ".")
	// 2. 校验 APIKey 格式
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid api key format")
	}
	// 智谱ai要求把这里的apikey使用jwt来隔开
	id, secret := parts[0], parts[1]
	// 3. 校验 APIKey 格式
	payload := jwt.MapClaims{
		"api_key":   id,
		"exp":       time.Now().Add(time.Minute * 30).UnixMilli(),
		"timestamp": time.Now().UnixMilli(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token.Header["sign_type"] = "SIGN"

	return token.SignedString([]byte(secret))
}

func VideoPlayProxy(localTaskID, videoURL string) (*http.Response, error) {
	// 1. 初始化智谱服务（复用原有鉴权逻辑）
	zhipuService := NewZhipuService()
	// 2. 生成智谱JWT Token（核心鉴权）
	token, err := zhipuService.generateToken()
	if err != nil {
		return nil, fmt.Errorf("生成鉴权Token失败: %w", err)
	}

	// 3. 构造请求（携带JWT + 支持分片）
	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构造视频请求失败: %w", err)
	}

	// 核心：添加智谱鉴权头
	req.Header.Set("Authorization", "Bearer "+token)
	// 设置超时时间
	client := &http.Client{Timeout: 30 * time.Second}

	// 4. 发送请求到智谱存储
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求智谱视频失败: %w", err)
	}

	// 5. 校验响应状态（200=完整播放，206=分片播放）
	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		resp.Body.Close() // 关闭响应体，避免内存泄漏
		return nil, fmt.Errorf("智谱鉴权失败，状态码: %d", resp.StatusCode)
	}

	return resp, nil
}
