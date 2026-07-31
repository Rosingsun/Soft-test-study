package service

import "errors"

// 业务层统一错误，handler 层据此映射 HTTP 错误码
var (
	ErrNotFound  = errors.New("资源不存在")
	ErrForbidden = errors.New("无权操作")
	ErrConflict  = errors.New("状态冲突")
)
