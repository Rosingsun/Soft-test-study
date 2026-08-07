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
		&model.CheckIn{},
		&model.CheckInQuestion{},
		&model.StudyPlan{},
		&model.ReviewCard{},
	); err != nil {
		return nil, err
	}

	// practice_records 复合索引，支撑打卡/计划/排行的每日与区间聚合
	var idxExists int64
	db.Raw(`SELECT COUNT(*) FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = 'practice_records' AND index_name = 'idx_user_created_at'`).
		Scan(&idxExists)
	if idxExists == 0 {
		if err := db.Exec(
			"ALTER TABLE practice_records ADD INDEX idx_user_created_at (user_id, created_at)",
		).Error; err != nil {
			return nil, err
		}
	}

	// 回填 check_ins 历史数据的首次打卡快照（rank_*）：
	// 历史行 rank_* 为 0，将其回填为当日 correct_count/accuracy/duration，避免排名与个人平均被清零。
	if err := db.Exec(
		"UPDATE check_ins SET rank_correct_count = correct_count, rank_accuracy = accuracy, rank_duration = duration WHERE rank_correct_count = 0 AND rank_accuracy = 0",
	).Error; err != nil {
		return nil, err
	}

	return db, nil
}
