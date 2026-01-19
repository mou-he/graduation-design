package aihelper

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// 流回调函数
type StreamCallback func(msg string)

// 模型接口
type AIModel interface {
	// 生成响应
	GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
	// 流式响应
	StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error)
	//获取模型类型
	GetModelType() string
}
type OpenAIModel struct {
	// 模型
	llm model.ToolCallingChatModel
}

// 创建OpenAI模型
func NewOpenAIModel(ctx context.Context) (*OpenAIModel, error) {
	key := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  key,
	})
	if err != nil {
		return nil, fmt.Errorf("create openai model failed: %v", err)
	}
	return &OpenAIModel{
		llm: llm,
	}, nil
}
func (o *OpenAIModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	resp, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("generate openai response failed: %v", err)
	}
	return resp, nil
}

// 流式响应
func (o *OpenAIModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	stream, err := o.llm.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("stream openai response failed: %v", err)
	}
	defer stream.Close()
	// 聚合响应
	var fullResp strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("openai stream recv failed: %v", err)
		}
		if len(msg.Content) > 0 {
			fullResp.WriteString(msg.Content) // 聚合

			cb(msg.Content) // 实时调用cb函数，方便主动发送给前端
		}
	}
	return fullResp.String(), nil
}
func (o *OpenAIModel) GetModelType() string {
	return "openai"
}

// Ollama模型实现
type OllamaModel struct {
	// 模型
	llm model.ToolCallingChatModel
}

func NewOllamaModel(ctx context.Context, baseURL, modelName string) (*OllamaModel, error) {
	llm, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("create ollama model failed: %v", err)
	}
	return &OllamaModel{
		llm: llm,
	}, nil
}

func (o *OllamaModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	resp, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("generate ollama response failed: %v", err)
	}
	return resp, nil
}

func (o *OllamaModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	stream, err := o.llm.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("stream ollama response failed: %v", err)
	}
	// 关闭流资源
	defer stream.Close()
	// 聚合相应
	var fullResp strings.Builder
	// 读取数据流

	for {
		msg, err := stream.Recv()
		// 流数据读取
		if err == io.EOF {
			break
		}
		// 处理流数据
		if len(msg.Content) > 0 {
			fullResp.WriteString(msg.Content) // 聚合
			cb(msg.Content)                   // 实时调用cb函数，方便主动发送给前端
		}
	}
	return fullResp.String(), nil
}

func (o *OllamaModel) GetModelType() string { return "ollama" }
