package config

// 统一错误码
const (
	CodeSuccess         = 0     // 成功
	CodeServerError     = 10000 // 服务器内部错误
	CodeBadRequest      = 10001 // 业务/请求处理失败
	CodeParamError      = 10002 // 参数校验失败
	CodeUnauthorized    = 10003 // 未登录 / Token 无效
	CodeForbidden       = 10004 // 无权限
	CodeNotFound        = 10005 // 资源不存在
	// OPT-23: 限流专用错误码（10004 已被 CodeForbidden 占用，避免前端误判）
	CodeTooManyRequests = 10006 // 请求过于频繁
	CodeDuplicate       = 10009 // 资源冲突（如知识点已存在）
)
