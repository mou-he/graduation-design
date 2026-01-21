package redis

import (
	"fmt"

	"github.com/mou-he/graduation-design/config"
)

// 生成验证码的key
func GenerateCaptcha(email string) string {
	return fmt.Sprintf("%s", config.DefaultRedisKeyConfig.CaptchaPrefix)
}
