package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mou-he/graduation-design/common/aihelper"
	"github.com/mou-he/graduation-design/common/mysql"
	"github.com/mou-he/graduation-design/common/rabbitmq"
	"github.com/mou-he/graduation-design/common/redis"
	"github.com/mou-he/graduation-design/config"
	"github.com/mou-he/graduation-design/dao/message"
	"github.com/mou-he/graduation-design/router"
)

func StartServer(addr string, port int) error {
	r := router.InitRouter()
	return r.Run(fmt.Sprintf("%s:%d", addr, port))
}
func readDataFromDB() error {
	manager := aihelper.GetGlobalManager()
	// 从数据库中读取数据
	msgs, err := message.GetAllMessages()
	if err != nil {
		return err
	}
	for i := range msgs {
		//
		m := msgs[i]
		// 默认openAI模型
		modelType := "1"
		config := make(map[string]interface{})
		// 创建对应的AIhelper
		helper, err := manager.GetOrCreateAIHelper(m.UserName, m.SessionID, modelType, config)
		if err != nil {
			log.Printf("[readDataFromDB] failed to create helper for user=%s session=%s: %v", m.UserName, m.SessionID, err)
			continue
		}
		log.Println("readDataFromDB init:  ", helper.SessionID)
		// 添加消息到内存中(不开启存储功能)
		helper.AddMessage(m.Content, m.UserName, m.IsUser, false)
	}
	log.Println("readDataFromDB success  ")
	return nil
}

func main() {
	os.Setenv("ZHIPU_API_KEY", "602358133cad480da7d43d4d0e5a3b0f.vMxr0PqkWmXAL2FT")
	os.Setenv("RAG_API_KEY", "sk-97ffa234a0724d1190c5dae2a6b287ec")
	os.Setenv("OPENAI_API_KEY", "sk-6e277f0001ca4714b3f96125b5e3910c")
	os.Setenv("OPENAI_MODEL_NAME", "qwen-plus")
	os.Setenv("OPENAI_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	conf := config.GetConfig()
	host := conf.MainConfig.Host
	port := conf.MainConfig.Port
	//初始化mysql
	if err := mysql.InitMysql(); err != nil {
		log.Println("InitMysql error , " + err.Error())
		return
	}
	//初始化AIHelperManager
	readDataFromDB()
	//初始化redis
	redis.Init()
	log.Println("redis init success  ")
	rabbitmq.InitRabbitMQ()
	log.Println("rabbitmq init success  ")

	err := StartServer(host, port) // 启动 HTTP 服务
	if err != nil {
		panic(err)
	}
}
