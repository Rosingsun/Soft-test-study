package service

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

var (
	ErrInvalidFilePath = errors.New("非法文件路径")
	ErrInvalidXmind    = errors.New("xmind 文件格式错误或内容为空")
)

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

// ========================================================================
// 思维导图在线阅读：解析 xmind（ZIP 包内 content.json）为树形结构
// ========================================================================

// xmindTopic xmind content.json 中的主题节点（仅解析所需字段）
type xmindTopic struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Href     string `json:"href"`
	Children *struct {
		Attached []xmindTopic `json:"attached"`
		Detached []xmindTopic `json:"detached"`
	} `json:"children"`
}

type xmindSheet struct {
	ID        string     `json:"id"`
	Class     string     `json:"class"`
	RootTopic xmindTopic `json:"rootTopic"`
}

// Content 返回资料的思维导图树形结构（不改变预览计数）
func (s *StudyMaterialService) Content(id uint) ([]dto.MindMapNode, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if m.Status != 1 {
		return nil, ErrNotFound
	}
	absPath, err := s.resolveFilePath(m.FileURL)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(absPath); err != nil {
		return nil, ErrNotFound
	}

	topics, err := parseXmind(absPath)
	if err != nil {
		return nil, err
	}

	nodes := make([]dto.MindMapNode, 0, len(topics))
	for _, t := range topics {
		if node := convertTopic(t); node != nil {
			nodes = append(nodes, *node)
		}
	}
	if len(nodes) == 0 {
		return nil, ErrInvalidXmind
	}
	return nodes, nil
}

// parseXmind 解压并解析 xmind 文件的 content.json，返回所有画布的根主题
func parseXmind(path string) ([]xmindTopic, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, ErrInvalidXmind
	}
	defer reader.Close()

	var raw []byte
	for _, f := range reader.File {
		if f.Name == "content.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, ErrInvalidXmind
			}
			raw, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, ErrInvalidXmind
			}
			break
		}
	}
	if len(raw) == 0 {
		return nil, ErrInvalidXmind
	}

	var sheets []xmindSheet
	if err := json.Unmarshal(raw, &sheets); err != nil {
		return nil, ErrInvalidXmind
	}
	roots := make([]xmindTopic, 0, len(sheets))
	for _, sheet := range sheets {
		if sheet.RootTopic.Title != "" || len(sheet.RootTopic.Children.Attached) > 0 {
			roots = append(roots, sheet.RootTopic)
		}
	}
	if len(roots) == 0 {
		return nil, ErrInvalidXmind
	}
	return roots, nil
}

// convertTopic 递归将 xmind 主题转换为 DTO 节点，空标题节点被跳过
func convertTopic(t xmindTopic) *dto.MindMapNode {
	title := strings.TrimSpace(t.Title)
	var children []dto.MindMapNode
	if t.Children != nil {
		for _, child := range t.Children.Attached {
			if node := convertTopic(child); node != nil {
				children = append(children, *node)
			}
		}
		for _, child := range t.Children.Detached {
			if node := convertTopic(child); node != nil {
				children = append(children, *node)
			}
		}
	}
	if title == "" && len(children) == 0 {
		return nil
	}
	return &dto.MindMapNode{Title: title, Children: children}
}
