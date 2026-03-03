package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/mou-he/graduation-design/config"
)

type TTSService struct{}

func NewTTSService() *TTSService {
	return &TTSService{}
}

// 请求
type TTSRequest struct {
	Text           string `json:"text"`
	Format         string `json:"format"`
	Voice          int    `json:"voice"`
	Lang           string `json:"lang"`
	Speed          int    `json:"speed"`
	Pitch          int    `json:"pitch"`
	Volume         int    `json:"volume"`
	EnableSubtitle int    `json:"enable_subtitle"`
}

// 响应
type TTSCreateResponse struct {
	TaskID string `json:"task_id"`
}

func (s *TTSService) GetAccessToken() string {
	config := config.GetConfig()
	// 构建POST请求参数
	url := "https://aip.baidubce.com/oauth/2.0/token"
	postData := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		config.VoiceServiceConfig.VoiceServiceApiKey,
		config.VoiceServiceConfig.VoiceServiceSecretKey)
	// 发送POST请求
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(postData))
	if err != nil {
		log.Println("Post fail err is : ", err)
		return ""
	}
	defer resp.Body.Close()
	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("ReadAll fail err is : ", err)
		return ""
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	// 解析响应体
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Println("Unmarshal fail err is : ", err)
		return ""
	}
	return tokenResp.AccessToken
}

func (s *TTSService) CreateTTS(ctx context.Context, text string) (string, error) {
	accessToken := s.GetAccessToken()
	if accessToken == "" {
		log.Println("GetAccessToken fail")
		return "", fmt.Errorf("get access token fail")
	}
	payload := TTSRequest{
		Text:           text,
		Format:         "mp3-16k",
		Voice:          4200,
		Lang:           "zh",
		Speed:          5,
		Pitch:          5,
		Volume:         5,
		EnableSubtitle: 0,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Marshal fail err is : ", err)
		return "", err
	}
	// 构建post请求
	url := "https://aip.baidubce.com/rpc/2.0/tts/v1/create?access_token=" + accessToken
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err

	}
	log.Println("[TTS Create] raw:", string(body))
	// 解析响应体
	var ttsResp TTSCreateResponse
	if err := json.Unmarshal(body, &ttsResp); err != nil {
		log.Println("Unmarshal fail err is : ", err)
		return "", err
	}
	return ttsResp.TaskID, nil

}

// 查询
// 查询任务状态
type TTSTaskResult struct {
	SpeechURL string `json:"speech_url,omitempty"`
}

type TTSTask struct {
	TaskID     string         `json:"task_id"`
	TaskStatus string         `json:"task_status"`
	TaskResult *TTSTaskResult `json:"task_result,omitempty"`
}

type TTSQueryResponse struct {
	LogID     string    `json:"log_id"`
	TasksInfo []TTSTask `json:"tasks_info"`
}

func (s *TTSService) QueryTTSFull(ctx context.Context, taskID string) (*TTSQueryResponse, error) {
	accessToken := s.GetAccessToken()
	if accessToken == "" {
		log.Println("GetAccessToken fail")
		return nil, fmt.Errorf("get access token fail")
	}
	reqBody := map[string][]string{
		"task_ids": {taskID},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Println("Marshal fail err is : ", err)
		return nil, err
	}
	// 构建post请求
	url := "https://aip.baidubce.com/rpc/2.0/tts/v1/query?access_token=" + accessToken
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Println("[TTS Query] raw:", string(respBody))
	// 官方返回原始 JSON
	var rawResp struct {
		LogID     json.Number `json:"log_id"`
		TasksInfo []struct {
			TaskID     string          `json:"task_id"`
			TaskStatus string          `json:"task_status"`
			TaskResult json.RawMessage `json:"task_result,omitempty"`
		} `json:"tasks_info"`
	}

	if err := json.Unmarshal(respBody, &rawResp); err != nil {
		return nil, err
	}

	result := &TTSQueryResponse{
		LogID:     rawResp.LogID.String(),
		TasksInfo: make([]TTSTask, 0, len(rawResp.TasksInfo)),
	}

	for _, t := range rawResp.TasksInfo {
		task := TTSTask{
			TaskID:     t.TaskID,
			TaskStatus: t.TaskStatus,
			TaskResult: nil, // 默认 nil
		}

		if t.TaskStatus == "Success" && len(t.TaskResult) > 0 {
			var r TTSTaskResult
			if err := json.Unmarshal(t.TaskResult, &r); err != nil {
				log.Println("parse task_result error:", err)
				return nil, fmt.Errorf("failed to parse task result: %v", err)
			}
			task.TaskResult = &r
		}

		result.TasksInfo = append(result.TasksInfo, task)
	}

	return result, nil

}
