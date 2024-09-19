package db

import (
	"learn/internal/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("question_bank.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	// 自动迁移表结构
	DB.AutoMigrate(&models.QuestionBank{}, &models.QuestionType{}, &models.Question{}, &models.AnswerOption{}, &models.RelatedQuestion{})
}
