package database

import (
	"fmt"

	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Init(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.ExamLevel{},
		&model.Subject{},
		&model.SubSubject{},
		&model.Chapter{},
		&model.Question{},
		&model.PracticeRecord{},
		&model.QuestionKnowledge{},
		&model.FavoriteFolder{},
		&model.QuestionFavorite{},
		&model.QuestionMark{},
		&model.WrongQuestion{},
		&model.ExamTemplate{},
		&model.ExamTemplateQuestion{},
		&model.ExamRecord{},
		&model.ExamRecordAnswer{},
		&model.AiGeneratedQuestion{},
		&model.EssayScore{},
		&model.StudyMaterial{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
