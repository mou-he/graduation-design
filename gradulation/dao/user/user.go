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

func GetUserByUsername(username string) (*model.User, error) {
	user := new(model.User)
	err := mysql.DB.Where("username = ?", username).First(user).Error
	return user, err
}

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
