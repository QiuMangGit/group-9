package model

import (
	"gorm.io/gorm"
	"time"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

type PostAddRobot struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ChatRobot struct {
	Email       string `json:"email"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ChatRobotSql struct {
	Email       string `gorm:"column:Email;type:varchar(255);not null" json:"email"`
	Id          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:varchar(255);not null" json:"description"`
}

// 添加用户模型
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"type:varchar(100);not null" json:"-"` // 密码哈希
	Email     string         `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
