import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { listNotifications, getUnreadCount, markNotificationRead, markAllNotificationsRead } from '@/api/notification'
import type { Notification } from '@/types/notification'

export const useNotificationStore = defineStore('notification', () => {
  const list = ref<Notification[]>([])
  const unreadCount = ref(0)
  const loading = ref(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const hasUnread = computed(() => unreadCount.value > 0)

  async function fetchList(limit = 20) {
    loading.value = true
    try {
      const resp = await listNotifications(limit)
      list.value = resp.list
      unreadCount.value = resp.unread_count
    } catch (e) {
      // 静默失败：轮询时不应弹错误
      console.warn('[notification] fetchList failed', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchUnread() {
    try {
      const resp = await getUnreadCount()
      unreadCount.value = resp.unread_count
    } catch (e) {
      console.warn('[notification] fetchUnread failed', e)
    }
  }

  async function markRead(id: number) {
    try {
      await markNotificationRead(id)
      // 本地状态更新
      const n = list.value.find(x => x.id === id)
      if (n && !n.read) {
        n.read = true
        unreadCount.value = Math.max(0, unreadCount.value - 1)
      }
    } catch (e) {
      console.warn('[notification] markRead failed', e)
    }
  }

  async function markAllRead() {
    try {
      await markAllNotificationsRead()
      list.value.forEach(n => { n.read = true })
      unreadCount.value = 0
    } catch (e) {
      console.warn('[notification] markAllRead failed', e)
    }
  }

  // 启动 30s 轮询未读数
  function startPolling(intervalMs = 30000) {
    if (pollTimer) return
    fetchUnread()
    pollTimer = setInterval(fetchUnread, intervalMs)
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  return {
    list,
    unreadCount,
    loading,
    hasUnread,
    fetchList,
    fetchUnread,
    markRead,
    markAllRead,
    startPolling,
    stopPolling,
  }
})
