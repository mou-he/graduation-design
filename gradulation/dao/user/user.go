package user

import (
	"github.com/mou-he/graduation-design/common/mysql"
	"github.com/mou-he/graduation-design/model"
	"github.com/mou-he/graduation-design/utils"
	"gorm.io/gorm"
)

func InsertUser(user *model.User) (*model.User, error) {
	err := mysql.DB.Create(&user).Error
	return user, err
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*model.User, error) {
	user := new(model.User)
	err := mysql.DB.Where("username = ?", username).First(user).Error
	return user, err
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(email string) (*model.User, error) {
	user := new(model.User)
	err := mysql.DB.Where("email = ?", email).First(user).Error
	return user, err
}

// IsEmailExist 判断邮箱是否存在
func IsEmailExist(email string) (bool, *model.User) {
	user, err := GetUserByEmail(email)
	if err == gorm.ErrRecordNotFound || user == nil {
		return false, nil
	}
	return true, user
}

// IsUserExist 判断用户名是否存在
func IsUserExist(username string) (bool, *model.User) {
	user, err := GetUserByUsername(username)
	if err == gorm.ErrRecordNotFound || user == nil {
		return false, nil
	}
	return true, user
}

func Register(username, email, password string) (*model.User, bool) {
	if user, err := InsertUser(&model.User{
		Email:    email,
		Name:     username,
		Username: username,
		Password: utils.MD5(password),
	}); err != nil {
		return nil, false
	} else {
		return user, true
	}
}

func UpdateUserPassword(userID int64, password string) error {

	return mysql.DB.Model(&model.User{}).Where("id = ?", userID).Update("password", password).Error
}
