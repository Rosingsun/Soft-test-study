package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

type BookmarkService struct {
	repo *repository.BookmarkRepo
}

func NewBookmarkService(repo *repository.BookmarkRepo) *BookmarkService {
	return &BookmarkService{repo: repo}
}

func (s *BookmarkService) CreateFolder(userID uint, req dto.CreateFolderReq) (*dto.FolderResp, error) {
	folder := &model.FavoriteFolder{
		UserID: userID,
		Name:   req.Name,
	}
	if err := s.repo.CreateFolder(folder); err != nil {
		return nil, err
	}
	return &dto.FolderResp{ID: folder.ID, Name: folder.Name}, nil
}

func (s *BookmarkService) ListFolders(userID uint) ([]dto.FolderResp, error) {
	folders, err := s.repo.FindFoldersByUser(userID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.FolderResp, len(folders))
	for i, f := range folders {
		resp[i] = dto.FolderResp{ID: f.ID, Name: f.Name}
	}
	return resp, nil
}

func (s *BookmarkService) AddFavorite(userID, questionID, folderID uint) error {
	// folderID > 0 时必须校验归属：
	//   - folderID == 0：表示"未分类"，无需校验
	//   - folderID > 0：必须存在且属于当前用户，防止越权写入他人收藏夹
	if folderID > 0 {
		if _, err := s.repo.FindFolderByID(folderID, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 不区分"不存在"与"归属不符"，统一返回 ErrForbidden
				// 避免通过接口探测他人 folderID
				return ErrForbidden
			}
			return err
		}
	}
	fav := &model.QuestionFavorite{
		UserID:     userID,
		QuestionID: questionID,
		FolderID:   folderID,
	}
	return s.repo.AddFavorite(fav)
}

func (s *BookmarkService) RemoveFavorite(userID, questionID uint) error {
	return s.repo.RemoveFavorite(userID, questionID)
}

func (s *BookmarkService) ListFavorites(userID uint) ([]dto.QuestionFavoriteResp, error) {
	list, err := s.repo.FindFavoritesByUser(userID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.QuestionFavoriteResp, len(list))
	for i, v := range list {
		resp[i] = dto.QuestionFavoriteResp{
			ID:         v.ID,
			QuestionID: v.QuestionID,
			FolderID:   v.FolderID,
			CreatedAt:  v.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return resp, nil
}

func (s *BookmarkService) ListFavoritesByFolder(userID, folderID uint) ([]dto.QuestionFavoriteResp, error) {
	list, err := s.repo.FindFavoritesByFolder(userID, folderID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.QuestionFavoriteResp, len(list))
	for i, v := range list {
		resp[i] = dto.QuestionFavoriteResp{
			ID:         v.ID,
			QuestionID: v.QuestionID,
			FolderID:   v.FolderID,
			CreatedAt:  v.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return resp, nil
}

func (s *BookmarkService) IsFavorited(userID, questionID uint) (bool, error) {
	return s.repo.IsFavorited(userID, questionID)
}
