package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mou-he/graduation-design/config"
	"github.com/redis/go-redis/v9"
)

// 初始化Redis客户端
var Rdb *redis.Client

// 上下文
var ctx = context.Background()

func Init() {
	conf := config.GetConfig()
	host := conf.RedisConfig.RedisHost
	port := conf.RedisConfig.RedisPort
	password := conf.RedisConfig.RedisPassword
	db := conf.RedisConfig.RedisDb
	addr := fmt.Sprintf("%s:%d", host, port)
	// 初始化客户端
	Rdb = redis.NewClient(&redis.Options{
		Password: password,
		DB:       db,
		Addr:     addr,
	})

}

// 设置验证码
func SetCaptchaForEmail(email, captcha string) error {
	// 生成验证码的key
	key := GenerateCaptcha(email)
	expire := 3 * time.Minute
	return Rdb.Set(ctx, key, captcha, expire).Err()
}

// 检查验证码是否匹配
func CheckCaptchaForEmail(email, userInput string) (bool, error) {
	key := GenerateCaptcha(email)
	storedCaptcha, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {

			return false, nil
		}
		return false, err
	}
	// 比较验证码
	if strings.EqualFold(storedCaptcha, userInput) {
		// 验证成功后删除 key
		if err := Rdb.Del(ctx, key).Err(); err != nil {

		} else {

		}
		return true, nil
	}

	return false, nil
}
