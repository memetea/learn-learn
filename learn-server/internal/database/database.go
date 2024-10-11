package database

import (
	"learn/internal/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	// 自动迁移表结构
	return db.AutoMigrate(
		&models.QuestionBank{},
		&models.Question{},
		&models.QuestionAttempt{},
		&models.Tag{},
		&models.AnswerOption{},
		&models.TrueFalseAnswer{},
		&models.WrittenAnswer{},
		&models.FillInTheBlankAnswer{},
		&models.RelatedQuestion{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
	)
}
