package repository

import (
	"group-9/model"
	"log"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)
type Repository struct {
	db *gorm.DB
}

func NewRepository(dsn string) (*Repository, error) {
        db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
        if err != nil {
                return nil, fmt.Errorf("failed to connect database: %w", err)
        }

        // 添加用户模型迁移
        if err = db.Migrator().AutoMigrate(&model.ChatRobot{}, &model.User{}); err != nil {
                return nil, fmt.Errorf("failed to migrate database: %w", err)
        }
        log.Println("database connected successfully")
        return &Repository{db: db}, nil
}
