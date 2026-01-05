package mysql

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/config"
	"github.com/mou-he/graduation-design/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMysql() error {
	// 从配置文件中获取数据库连接信息
	host := config.GetConfig().MysqlHost
	port := config.GetConfig().MysqlPort
	dbname := config.GetConfig().MysqlDatabaseName
	username := config.GetConfig().MysqlUser
	password := config.GetConfig().MysqlPassword
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)
	var log logger.Interface
	if gin.Mode() == "debug" {
		log = logger.Default.LogMode(logger.Info)
	} else {
		log = logger.Default
	}
	// 初始化数据库链接
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: log,
	})
	if err != nil {
		return err
	}
	// 配置数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	// 设置数据库连接池参数
	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(100)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(10)
	// 全局变量DB赋值
	DB = db
	return migration()
}
func migration() error {
	// 自动迁移数据库表
	err := DB.AutoMigrate(new(model.User), new(model.Message), new(model.Session))
	if err != nil {
		return err
	}
	return nil
}

func InsertUser(user *model.User) error {
	err := DB.Create(user).Error
	if err != nil {
		return err
	}
	return nil
}
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}
