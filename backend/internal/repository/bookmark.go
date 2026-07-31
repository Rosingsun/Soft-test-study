package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type BookmarkRepo struct {
	db *gorm.DB
}

func NewBookmarkRepo(db *gorm.DB) *BookmarkRepo {
	return &BookmarkRepo{db: db}
}

func (r *BookmarkRepo) CreateFolder(folder *model.FavoriteFolder) error {
	return r.db.Create(folder).Error
}

func (r *BookmarkRepo) FindFoldersByUser(userID uint) ([]model.FavoriteFolder, error) {
	var list []model.FavoriteFolder
	err := r.db.Where("user_id = ?", userID).Order("id asc").Find(&list).Error
	return list, err
}

func (r *BookmarkRepo) AddFavorite(fav *model.QuestionFavorite) error {
	return r.db.Create(fav).Error
}

func (r *BookmarkRepo) RemoveFavorite(userID, questionID uint) error {
	return r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&model.QuestionFavorite{}).Error
}

func (r *BookmarkRepo) FindFavoritesByUser(userID uint) ([]model.QuestionFavorite, error) {
	var list []model.QuestionFavorite
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *BookmarkRepo) FindFavoritesByFolder(userID, folderID uint) ([]model.QuestionFavorite, error) {
	var list []model.QuestionFavorite
	err := r.db.Where("user_id = ? AND folder_id = ?", userID, folderID).
		Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *BookmarkRepo) IsFavorited(userID, questionID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.QuestionFavorite{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Count(&count).Error
	return count > 0, err
}
