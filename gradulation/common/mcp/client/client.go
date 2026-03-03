package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// McpClient 是 MCP 客户端的结构体
// 它包含了一个 MCP 客户端的实例
// 它用于发送和接收 MCP 消息
// 它的方法用于发送和接收 MCP 消息
type McpClient struct {
	client *client.Client
}

func NewMcpClient(httpURL string) (*McpClient, error) {
	fmt.Println("正在初始化HTTP客户端")
	// 创建http传输
	httpTransport, err := transport.NewStreamableHTTP(httpURL)
	if err != nil {
		log.Println("创建HTTP传输失败 err is : ", err)
		return nil, fmt.Errorf("创建HTTP传输失败 err is : %w", err)
	}
	// 使用传输创建客户端
	c := client.NewClient(httpTransport)
	return &McpClient{client: c}, nil
}

// 初始化客户端
func (m *McpClient) Init(ctx context.Context) (*mcp.InitializeResult, error) {
	//设置通知处理程序
	m.client.OnNotification(func(notification mcp.JSONRPCNotification) {
		fmt.Println("收到通知:", notification)
	})
	// 初始化客户端
	fmt.Println("正在初始化客户端")
	// 初始化请求
	initRequest := mcp.InitializeRequest{}
	// 初始化请求参数
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "MCP-haiAI Weather Client",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}
	// 初始化客户端
	serverInfo, err := m.client.Initialize(ctx, initRequest)
	if err != nil {
		return nil, fmt.Errorf("初始化失败: %w", err)
	}
	fmt.Println("初始化成功, 服务器信息:", serverInfo)
	return serverInfo, nil
}

// Ping 执行健康检查
func (m *McpClient) Ping(ctx context.Context) error {
	fmt.Println("正在执行健康检查...")
	if err := m.client.Ping(ctx); err != nil {
		return fmt.Errorf("健康检查失败: %w", err)
	}
	fmt.Println("服务器正常运行并响应")
	return nil
}

// CallTool 调用MCP工具
func (m *McpClient) CallTool(ctx context.Context, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	callToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}

	result, err := m.client.CallTool(ctx, callToolRequest)
	if err != nil {
		return nil, fmt.Errorf("调用工具失败: %w", err)
	}

	return result, nil
}

// CallWeatherTool 调用get_weather工具
func (m *McpClient) CallWeatherTool(ctx context.Context, city string) (*mcp.CallToolResult, error) {
	fmt.Printf("正在查询城市 %s 的天气...\n", city)

	callToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_weather",
			Arguments: map[string]any{
				"city": city,
			},
		},
	}

	result, err := m.client.CallTool(ctx, callToolRequest)
	if err != nil {
		return nil, fmt.Errorf("调用工具失败: %w", err)
	}

	return result, nil
}

// GetToolResultText 获取工具结果中的文本内容
func (m *McpClient) GetToolResultText(result *mcp.CallToolResult) string {
	var text string
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			text += textContent.Text + "\n"
		}
	}
	return text
}

func (m *McpClient) Close() {
	if m.client != nil {
		m.client.Close()
	}
}
