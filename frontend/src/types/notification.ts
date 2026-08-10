export interface Notification {
  id: number
  type: string
  title: string
  content: string
  link: string
  read: boolean
  created_at: string
}

export interface NotificationListResp {
  list: Notification[]
  unread_count: number
}

export interface UnreadCountResp {
  unread_count: number
}
