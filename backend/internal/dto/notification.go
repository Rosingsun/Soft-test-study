package dto

// NotificationResp 单条通知响应
type NotificationResp struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Link      string `json:"link"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

// NotificationListResp 通知列表响应
type NotificationListResp struct {
	List       []NotificationResp `json:"list"`
	UnreadCount int64             `json:"unread_count"`
}

// UnreadCountResp 未读数响应
type UnreadCountResp struct {
	UnreadCount int64 `json:"unread_count"`
}
