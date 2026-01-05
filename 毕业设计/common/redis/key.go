package redis

import (
	"fmt"

	"github.com/mou-he/graduation-design/config"
)

func GenerateCaptcha(email string) string {
	return fmt.Sprintf("%s", config.DefaultRedisKeyConfig.CaptchaPrefix)
}
