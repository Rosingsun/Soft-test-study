package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type BookmarkHandler struct {
	svc *service.BookmarkService
}

func NewBookmarkHandler(svc *service.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{svc: svc}
}

func (h *BookmarkHandler) CreateFolder(c *gin.Context) {
	var req dto.CreateFolderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, "参数错误")
		return
	}
	userID, _ := c.Get("user_id")
	folder, err := h.svc.CreateFolder(userID.(uint), req)
	if err != nil {
		response.Error(c, 10001, "创建收藏夹失败")
		return
	}
	response.Success(c, folder)
}

func (h *BookmarkHandler) ListFolders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	folders, err := h.svc.ListFolders(userID.(uint))
	if err != nil {
		response.Error(c, 10001, "获取收藏夹失败")
		return
	}
	response.Success(c, folders)
}

func (h *BookmarkHandler) AddFavorite(c *gin.Context) {
	userID, _ := c.Get("user_id")
	questionIDStr := c.Param("id")
	questionID, _ := strconv.ParseUint(questionIDStr, 10, 64)

	folderIDStr := c.Query("folder_id")
	var folderID uint64
	if folderIDStr != "" {
		folderID, _ = strconv.ParseUint(folderIDStr, 10, 64)
	}

	if err := h.svc.AddFavorite(userID.(uint), uint(questionID), uint(folderID)); err != nil {
		response.Error(c, 10001, "收藏失败")
		return
	}
	response.Success(c, gin.H{"favorited": true})
}

func (h *BookmarkHandler) RemoveFavorite(c *gin.Context) {
	userID, _ := c.Get("user_id")
	questionIDStr := c.Param("id")
	questionID, _ := strconv.ParseUint(questionIDStr, 10, 64)

	if err := h.svc.RemoveFavorite(userID.(uint), uint(questionID)); err != nil {
		response.Error(c, 10001, "取消收藏失败")
		return
	}
	response.Success(c, gin.H{"favorited": false})
}

func (h *BookmarkHandler) ListFavorites(c *gin.Context) {
	userID, _ := c.Get("user_id")
	folderIDStr := c.Query("folder_id")

	var list interface{}
	var err error
	if folderIDStr != "" {
		folderID, _ := strconv.ParseUint(folderIDStr, 10, 64)
		list, err = h.svc.ListFavoritesByFolder(userID.(uint), uint(folderID))
	} else {
		list, err = h.svc.ListFavorites(userID.(uint))
	}
	if err != nil {
		response.Error(c, 10001, "获取收藏列表失败")
		return
	}
	response.Success(c, list)
}

func (h *BookmarkHandler) CheckFavorited(c *gin.Context) {
	userID, _ := c.Get("user_id")
	questionIDStr := c.Param("id")
	questionID, _ := strconv.ParseUint(questionIDStr, 10, 64)

	favorited, err := h.svc.IsFavorited(userID.(uint), uint(questionID))
	if err != nil {
		response.Error(c, 10001, "查询失败")
		return
	}
	response.Success(c, gin.H{"favorited": favorited})
}
