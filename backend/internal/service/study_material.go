package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

var ErrInvalidFilePath = errors.New("非法文件路径")

type StudyMaterialService struct {
	repo        *repository.StudyMaterialRepo
	subjectRepo *repository.SubjectRepo
}

func NewStudyMaterialService(repo *repository.StudyMaterialRepo, subjectRepo *repository.SubjectRepo) *StudyMaterialService {
	return &StudyMaterialService{repo: repo, subjectRepo: subjectRepo}
}

func (s *StudyMaterialService) List(subjectID uint) ([]dto.StudyMaterialResp, error) {
	list, err := s.repo.FindList(subjectID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.StudyMaterialResp, len(list))
	for i, v := range list {
		resp[i] = s.toResp(v)
	}
	return resp, nil
}

func (s *StudyMaterialService) Detail(id uint) (*dto.StudyMaterialResp, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if m.Status != 1 {
		return nil, ErrNotFound
	}
	_ = s.repo.IncrementViewCount(id)
	m.ViewCount++
	resp := s.toResp(*m)
	return &resp, nil
}

// Download 校验文件存在并递增下载次数，返回文件绝对路径与下载文件名
func (s *StudyMaterialService) Download(id uint) (string, string, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return "", "", ErrNotFound
	}
	if m.Status != 1 {
		return "", "", ErrNotFound
	}
	absPath, err := s.resolveFilePath(m.FileURL)
	if err != nil {
		return "", "", err
	}
	if _, err := os.Stat(absPath); err != nil {
		return "", "", ErrNotFound
	}
	if err := s.repo.IncrementDownloadCount(id); err != nil {
		return "", "", err
	}
	fileName := m.FileName
	if fileName == "" {
		fileName = filepath.Base(m.FileURL)
	}
	return absPath, fileName, nil
}

// resolveFilePath 将 URL 路径（/uploads/materials/1/note.xmind）映射为本地绝对路径
// 静态服务映射：/uploads -> ./data/uploads（相对后端运行目录）
// OPT-03: 加入路径穿越校验——解析后必须仍在 data/uploads 目录下
func (s *StudyMaterialService) resolveFilePath(fileURL string) (string, error) {
	rel := strings.TrimPrefix(fileURL, "/uploads/")
	if rel == fileURL {
		// 已是本地相对路径（未以 /uploads/ 开头），按数据目录解析
		rel = fileURL
	} else {
		rel = filepath.FromSlash(rel)
	}
	cleaned := filepath.Clean(rel)
	baseDir, err := filepath.Abs(filepath.Join("data", "uploads"))
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(filepath.Join(baseDir, cleaned))
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, baseDir+string(os.PathSeparator)) && absPath != baseDir {
		return "", ErrInvalidFilePath
	}
	return absPath, nil
}

func (s *StudyMaterialService) toResp(m model.StudyMaterial) dto.StudyMaterialResp {
	subjectName := ""
	if sub, err := s.subjectRepo.FindByID(m.SubjectID); err == nil {
		subjectName = sub.Name
	}
	return dto.StudyMaterialResp{
		ID:            m.ID,
		Title:         m.Title,
		Description:   m.Description,
		SubjectID:     m.SubjectID,
		SubjectName:   subjectName,
		CoverURL:      m.CoverURL,
		FileURL:       m.FileURL,
		FileName:      m.FileName,
		FileSize:      m.FileSize,
		DownloadCount: m.DownloadCount,
		ViewCount:     m.ViewCount,
		CreatedAt:     m.CreatedAt.Format(time.DateTime),
	}
}
