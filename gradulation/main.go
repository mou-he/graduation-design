package main

import (
	"fmt"
	"log"

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
		helper, err := manager.GetOrCreateAIHelper(modelType, m.Content, m.UserName, config)
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
