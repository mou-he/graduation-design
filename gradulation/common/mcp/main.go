package main

import (
	"fmt"

	mcp "github.com/kaitai/gopherai-mcp/server"
)

func main() {
	fmt.Println("正在启动MCP服务...")
	if err := mcp.StartServer(":8081"); err != nil {
		fmt.Printf("MCP 服务器启动失败: %v", err)
	}
}
