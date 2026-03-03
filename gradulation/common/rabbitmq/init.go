package rabbitmq

var (
	RMQMessage       *RabbitMQ // 用于消息队列
	RMQVideoSubmit   *RabbitMQ // 新增：用于提交视频生成任务
	RMQVideoCheck    *RabbitMQ // 新增：用于轮询检查进度
	RMQDeleteSession *RabbitMQ // 删除会话队列

)

func InitRabbitMQ() {
	// 创建MQ并启动消费者
	// 无论调用多少次 NewWorkRabbitMQ，只会创建一次连接
	// 不同队列共用一个连接，可以保持不同队列消费消息的顺序
	RMQMessage = NewWorkRabbitMQ("Message")
	go RMQMessage.Consume(MQMessage)

	// 2. 新增：视频提交队列
	RMQVideoSubmit = NewWorkRabbitMQ("VideoSubmit")
	go RMQVideoSubmit.Consume(MQVideoSubmit)

	// 3. 新增：视频检查队列
	RMQVideoCheck = NewWorkRabbitMQ("VideoCheck")
	go RMQVideoCheck.Consume(MQVideoCheck)
	// 4. 删除会话队列
	RMQDeleteSession = NewWorkRabbitMQ("DeleteSession")
	go RMQDeleteSession.Consume(MQDeleteSession)

}

// DestroyRabbitMQ 销毁RabbitMQ
func DestroyRabbitMQ() {
	RMQMessage.Destory()
}
