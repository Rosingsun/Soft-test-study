<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '@/stores/notification'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const notifyStore = useNotificationStore()
const auth = useAuthStore()

const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

const items = computed(() => notifyStore.list)
const unread = computed(() => notifyStore.unreadCount)

function toggle() {
  open.value = !open.value
  if (open.value) {
    notifyStore.fetchList(20)
  }
}

function close() {
  open.value = false
}

async function clickItem(id: number, link: string) {
  await notifyStore.markRead(id)
  close()
  if (link) {
    // 站内链接：与当前路由不同时跳转
    if (router.currentRoute.value.fullPath !== link) {
      router.push(link)
    }
  }
}

async function markAll() {
  await notifyStore.markAllRead()
}

function handleClickOutside(e: MouseEvent) {
  if (!open.value) return
  const target = e.target as Node
  if (rootRef.value && !rootRef.value.contains(target)) {
    close()
  }
}

function formatTime(s: string): string {
  if (!s) return ''
  // 服务端返回 "2006-01-02 15:04:05" 格式
  const d = new Date(s.replace(/-/g, '/'))
  if (isNaN(d.getTime())) return s
  const now = new Date()
  const diff = (now.getTime() - d.getTime()) / 1000
  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
  if (diff < 86400 * 3) return `${Math.floor(diff / 86400)} 天前`
  return s.slice(0, 10)
}

function iconFor(type: string): string {
  if (type === 'ai_generate_done') return 'check'
  if (type === 'ai_generate_failed') return 'x'
  return 'bell'
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  if (auth.user) notifyStore.startPolling(30000)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  // 关键：定时器在全局 store 中，组件卸载必须显式停止，否则退出登录后仍会轮询受保护接口
  notifyStore.stopPolling()
})
</script>

<template>
  <div ref="rootRef" class="relative">
    <button
      class="relative flex h-9 w-9 cursor-pointer items-center justify-center rounded-xl text-gray-500 transition-colors hover:bg-gray-100"
      :class="{ 'bg-indigo-50 text-indigo-600': open }"
      @click.stop="toggle"
    >
      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
        <path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
      </svg>
      <span
        v-if="unread > 0"
        class="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-semibold leading-none text-white ring-2 ring-white"
      >{{ unread > 99 ? '99+' : unread }}</span>
    </button>

    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 -translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div
        v-if="open"
        class="absolute right-0 top-full z-50 mt-2 w-80 overflow-hidden rounded-xl border border-gray-200 bg-white shadow-lg"
      >
        <!-- 头部 -->
        <div class="flex items-center justify-between border-b border-gray-100 px-4 py-3">
          <h3 class="text-sm font-semibold text-gray-900">消息中心</h3>
          <button
            v-if="unread > 0"
            class="cursor-pointer text-xs text-indigo-600 transition-colors hover:text-indigo-700"
            @click="markAll"
          >全部已读</button>
        </div>

        <!-- 列表 -->
        <div class="max-h-96 overflow-y-auto">
          <div v-if="items.length === 0" class="px-4 py-12 text-center">
            <svg class="mx-auto mb-2 h-10 w-10 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
            </svg>
            <p class="text-sm text-gray-400">暂无消息</p>
          </div>

          <button
            v-for="n in items"
            :key="n.id"
            class="flex w-full cursor-pointer items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-gray-50"
            :class="{ 'bg-indigo-50/50': !n.read }"
            @click="clickItem(n.id, n.link)"
          >
            <div
              class="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full"
              :class="n.type === 'ai_generate_done' ? 'bg-emerald-50 text-emerald-600' : n.type === 'ai_generate_failed' ? 'bg-red-50 text-red-600' : 'bg-indigo-50 text-indigo-600'"
            >
              <svg v-if="iconFor(n.type) === 'check'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
              </svg>
              <svg v-else-if="iconFor(n.type) === 'x'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
              <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
              </svg>
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <p class="truncate text-sm font-medium" :class="n.read ? 'text-gray-700' : 'text-gray-900'">{{ n.title }}</p>
                <span v-if="!n.read" class="h-2 w-2 shrink-0 rounded-full bg-indigo-500"></span>
              </div>
              <p class="mt-0.5 line-clamp-2 text-xs text-gray-500">{{ n.content }}</p>
              <p class="mt-1 text-[11px] text-gray-400">{{ formatTime(n.created_at) }}</p>
            </div>
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>
