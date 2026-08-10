import { get, post } from './request'
import type { NotificationListResp, UnreadCountResp } from '@/types/notification'

export function listNotifications(limit = 20) {
  return get<NotificationListResp>('/notifications', { limit })
}

export function getUnreadCount() {
  return get<UnreadCountResp>('/notifications/unread-count')
}

export function markNotificationRead(id: number) {
  return post(`/notifications/${id}/read`)
}

export function markAllNotificationsRead() {
  return post('/notifications/read-all')
}
