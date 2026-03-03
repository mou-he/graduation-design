package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/mou-he/graduation-design/model"
)

func GetRandomNumbers(num int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	code := ""
	for i := 0; i < num; i++ {
		// 0~9随机数
		digit := r.Intn(10)
		code += strconv.Itoa(digit)
	}
	return code
}

// MD5 MD5加密
func MD5(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	return hex.EncodeToString(m.Sum(nil))
}
func GenerateUUID() string {
	return uuid.New().String()
}

func ConvertToModelMessage(sessionID string, userName string, msg *schema.Message) *model.Message {
	return &model.Message{
		SessionID: sessionID,
		UserName:  userName,
		Content:   msg.Content,
	}
}

func ConvertToSchemaMessages(msgs []*model.Message) []*schema.Message {
	schemaMsgs := make([]*schema.Message, 0, len(msgs))
	for _, m := range msgs {
		role := schema.Assistant
		if m.IsUser == false {
			role = schema.User
		}
		schemaMsgs = append(schemaMsgs, &schema.Message{
			Role:    role,
			Content: m.Content,
		})
	}
	return schemaMsgs
}

func RemoveAllFilesInDir(dir string) error {
	// 1. 校验目录是否存在
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		fmt.Printf("目录 %s 不存在，跳过删除\n", dir)
		return nil // 目录不存在，无需处理，返回nil避免报错
	}
	if err != nil {
		fmt.Printf("检查目录 %s 状态失败：%v\n", dir, err)
		return err
	}

	// 2. 读取目录内容
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("读取目录 %s 失败：%v\n", dir, err)
		return err
	}

	// 3. 遍历删除文件/子目录
	for _, entry := range entries {
		// 核心修复：拼接完整路径（dir + 文件名/目录名）
		fullPath := filepath.Join(dir, entry.Name())
		fmt.Printf("准备删除：%s\n", fullPath)

		if entry.IsDir() {
			// 递归删除子目录（传入完整路径）
			fmt.Printf("删除子目录：%s\n", fullPath)
			if err := RemoveAllFilesInDir(fullPath); err != nil {
				fmt.Printf("删除子目录 %s 失败：%v\n", fullPath, err)
				// 不返回err，继续删除其他文件（避免一个目录失败导致整体中断）
				continue
			}
			// 删除空目录本身
			if err := os.Remove(fullPath); err != nil {
				fmt.Printf("删除空目录 %s 失败：%v\n", fullPath, err)
				continue
			}
		} else {
			// 删除文件：先校验文件是否存在
			_, err := os.Stat(fullPath)
			if os.IsNotExist(err) {
				fmt.Printf("文件 %s 不存在，跳过删除\n", fullPath)
				continue
			}
			if err != nil {
				fmt.Printf("检查文件 %s 状态失败：%v\n", fullPath, err)
				continue
			}

			// 执行删除（使用完整路径）
			fmt.Printf("删除文件：%s\n", fullPath)
			if err := os.Remove(fullPath); err != nil {
				fmt.Printf("删除文件 %s 失败：%v\n", fullPath, err)
				// 不返回err，继续删除其他文件
				continue
			}
		}
	}

	fmt.Printf("目录 %s 下的内容已全部删除\n", dir)
	return nil
}
func ValidateFile(file *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".md" && ext != ".txt" {
		return fmt.Errorf("文件类型不正确，只允许 .md 或 .txt 文件，当前扩展名: %s", ext)
	}
	return nil
}
