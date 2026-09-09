<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubjectStore } from '@/stores/subject'
import { getSubjectsByLevel } from '@/api/subject'
import type { Subject } from '@/types/subject'
import NotificationBell from '@/components/layout/NotificationBell.vue'

const auth = useAuthStore()
const subjectStore = useSubjectStore()
const route = useRoute()
const router = useRouter()

const mobileOpen = ref(false)
const userOpen = ref(false)
const userMenuRef = ref<HTMLElement | null>(null)

// 科目导航下拉：把原来的「等级 + 科目」两个独立下拉合并为一个层级下拉
const subjectNavOpen = ref(false)
const subjectNavMenuRef = ref<HTMLElement | null>(null)
// 按等级缓存科目列表，避免重复请求
const subjectsByLevel = ref<Record<number, Subject[]>>({})
const loadingLevels = ref<Set<number>>(new Set())

const navGroups = computed(() => {
  const groups = [
    {
      title: '学习',
      items: [
        { label: '数据看板', icon: 'home', to: '/' },
        { label: '科目导航', icon: 'book', to: '/subjects' },
        { label: '每日打卡', icon: 'checkin', to: '/check-in' },
        { label: '学习资料', icon: 'materials', to: '/materials' },
        { label: '知识点', icon: 'knowledge', to: '/knowledge' },
      ],
    },
    {
      title: '练习',
      items: [
        { label: 'AI 学习', icon: 'ai', to: '/ai/practice' },
        { label: '专项练习', icon: 'target', to: '/practice/special' },
        { label: '今日复习', icon: 'review', to: '/review' },
        { label: '错题本', icon: 'wrong', to: '/wrong-questions' },
        { label: '我的收藏', icon: 'star', to: '/favorites' },
      ],
    },
    {
      title: '考试',
      items: [
        { label: '模拟考试', icon: 'exam', to: '/exam' },
        { label: '考试记录', icon: 'history', to: '/exam-records' },
      ],
    },
    {
      title: '我的',
      items: [
        { label: '学习进度', icon: 'chart', to: '/progress' },
        { label: '成绩排行', icon: 'trophy', to: '/rankings' },
        { label: '学习计划', icon: 'plan', to: '/plan' },
        { label: '个人中心', icon: 'user', to: '/profile' },
      ],
    },
    {
      title: '社区',
      items: [{ label: '经验分享', icon: 'chat', to: '/community' }],
    },
  ]
  if (auth.user?.role === 'admin') {
    groups.push({
      title: '管理',
      items: [
        { label: '邀请码管理', icon: 'ticket', to: '/admin/invite-codes' },
        { label: '用户管理', icon: 'users', to: '/admin/users' },
        { label: '科目管理', icon: 'layers', to: '/admin/subjects' },
        { label: '题库管理', icon: 'database', to: '/admin/questions' },
      ],
    })
  }
  return groups
})

const currentLevelName = computed(() => {
  if (!auth.user) return ''
  return subjectStore.levels.find(l => l.id === auth.selectedLevelId)?.name || auth.user.level_name || ''
})

const currentSubjectName = computed(() => {
  if (!auth.user) return ''
  return subjectStore.subjects.find(s => s.id === auth.selectedSubjectId)?.name || auth.user.subject_name || ''
})

function isActive(to: string) {
  if (to === '/') return route.path === '/'
  return route.path === to || route.path.startsWith(to + '/')
}

function handleClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (userOpen.value && userMenuRef.value && !userMenuRef.value.contains(target)) userOpen.value = false
  if (subjectNavOpen.value && subjectNavMenuRef.value && !subjectNavMenuRef.value.contains(target)) subjectNavOpen.value = false
}

onMounted(async () => {
  document.addEventListener('click', handleClickOutside)
  if (subjectStore.levels.length === 0) await subjectStore.fetchLevels()
  // 预加载当前等级的科目到 store，供其他页面使用
  if (auth.selectedLevelId) await subjectStore.ensureSubjects(auth.selectedLevelId)
  // 同时把当前等级的科目缓存到本地，供下拉使用
  if (auth.selectedLevelId && !subjectsByLevel.value[auth.selectedLevelId]) {
    await loadLevelSubjects(auth.selectedLevelId)
  }
})

// 预加载指定等级的科目（带缓存）
async function loadLevelSubjects(levelId: number) {
  if (subjectsByLevel.value[levelId] || loadingLevels.value.has(levelId)) return
  loadingLevels.value.add(levelId)
  try {
    subjectsByLevel.value[levelId] = await getSubjectsByLevel(levelId)
  } catch {
    subjectsByLevel.value[levelId] = []
  } finally {
  loadingLevels.value.delete(levelId)
  }
}

// 打开科目导航下拉时按需预加载所有等级科目
async function openSubjectNav() {
  subjectNavOpen.value = !subjectNavOpen.value
  if (subjectNavOpen.value && subjectStore.levels.length) {
    await Promise.allSettled(
      subjectStore.levels.map(l => loadLevelSubjects(l.id)),
    )
  }
}

function closeSubjectNav() {
  subjectNavOpen.value = false
}

// 在下拉中选择某个科目
async function pickSubject(levelId: number, subjectId: number) {
  closeSubjectNav()
  if (levelId !== auth.selectedLevelId) {
    // 跨等级切换：先把 store 切到目标等级的科目列表
    await subjectStore.fetchSubjectsByLevel(levelId)
  }
  await auth.updateSubjectLevel(levelId, subjectId)
  router.push(`/subjects/${subjectId}`)
}

function closeMobile() {
  mobileOpen.value = false
}

function navigate(to: string) {
  mobileOpen.value = false
  router.push(to)
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- 桌面侧边栏 -->
    <aside class="fixed inset-y-0 left-0 z-40 hidden w-64 flex-col border-r border-gray-200 bg-white lg:flex">
      <div class="flex h-16 items-center gap-2.5 border-b border-gray-100 px-5">
        <div class="bg-brand-gradient flex h-9 w-9 items-center justify-center rounded-xl text-white shadow-md shadow-indigo-600/25">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
          </svg>
        </div>
        <span class="text-lg font-bold tracking-tight text-gray-900">软考学系</span>
      </div>

      <nav class="flex-1 space-y-5 overflow-y-auto px-3 py-5">
        <div v-for="group in navGroups" :key="group.title">
          <p class="mb-1.5 px-3 text-xs font-semibold uppercase tracking-wider text-gray-400">{{ group.title }}</p>
          <div class="space-y-0.5">
            <button
              v-for="item in group.items"
              :key="item.to"
              class="flex w-full cursor-pointer items-center gap-3 rounded-xl px-3 py-2 text-sm font-medium transition-all duration-200"
              :class="isActive(item.to)
                ? 'bg-brand-gradient text-white shadow-md shadow-indigo-600/20'
                : 'text-gray-600 hover:bg-indigo-50/60 hover:text-indigo-700'"
              @click="navigate(item.to)"
            >
              <svg class="h-4.5 w-4.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <template v-if="item.icon === 'home'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </template>
                <template v-else-if="item.icon === 'book'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                </template>
                <template v-else-if="item.icon === 'materials'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                </template>
                <template v-else-if="item.icon === 'random'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.77 4.769L4 5.5m0 0H4m0 0V4m10.5 15.5h-.582m-15.356-2a8.001 8.001 0 0015.356 1.731L20 18.5m0 0V19m0 0h.5m0 0V14" />
                </template>
                <template v-else-if="item.icon === 'ai'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5" />
                </template>
                <template v-else-if="item.icon === 'target'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </template>
                <template v-else-if="item.icon === 'wrong'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </template>
                <template v-else-if="item.icon === 'star'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.562.562 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z" />
                </template>
                <template v-else-if="item.icon === 'exam'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </template>
                <template v-else-if="item.icon === 'history'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </template>
                <template v-else-if="item.icon === 'chart'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
                </template>
                <template v-else-if="item.icon === 'user'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
                </template>
                <template v-else-if="item.icon === 'chat'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 20.25c4.97 0 9-3.694 9-8.25s-4.03-8.25-9-8.25S3 7.444 3 12c0 2.104.859 4.023 2.273 5.48.432.447.74 1.04.586 1.641a4.483 4.483 0 01-.923 1.785A5.969 5.969 0 006 21c1.282 0 2.47-.113 3.401-.662a8.25 8.25 0 002.599.912z" />
                </template>
                <template v-else-if="item.icon === 'ticket'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 6v.75m0 3v.75m0 3v.75m0 3V18m-9-5.25h5.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H7.5c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125zM21 8.25V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v.75m18 0v8.25A2.25 2.25 0 0118.75 18H5.25A2.25 2.25 0 013 15.75V8.25" />
                </template>
                <template v-else-if="item.icon === 'users'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
                </template>
                <template v-else-if="item.icon === 'layers'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0l4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0l-5.571 3-5.571-3" />
                </template>
                <template v-else-if="item.icon === 'database'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 5.625c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" />
                </template>
                <template v-else-if="item.icon === 'checkin'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5M9 16.125v-1.5m3 1.5v-1.5m3 1.5v-1.5" />
                </template>
                <template v-else-if="item.icon === 'review'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
                </template>
                <template v-else-if="item.icon === 'trophy'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 18.75h-9m9 0a3 3 0 013 3h-15a3 3 0 013-3m9 0v-3.375c0-.621-.503-1.125-1.125-1.125h-.871M7.5 18.75v-3.375c0-.621.504-1.125 1.125-1.125h.872m5.007 0H9.497m5.007 0a7.454 7.454 0 01-.982-3.172M9.497 14.25a7.454 7.454 0 00.981-3.172M5.25 4.236c-.982.143-1.954.317-2.916.52A6.003 6.003 0 007.73 9.728M5.25 4.236V4.5c0 2.108.966 3.99 2.48 5.228M5.25 4.236V2.721C7.456 2.41 9.71 2.25 12 2.25c2.291 0 4.545.16 6.75.47v1.516M7.73 9.728a6.726 6.726 0 002.748 1.35m8.272-6.842V4.5c0 2.108-.966 3.99-2.48 5.228m2.48-5.492a46.32 46.32 0 012.916.52 6.003 6.003 0 01-5.395 4.972m0 0a6.726 6.726 0 01-2.749 1.35m0 0a6.772 6.772 0 01-3.044 0" />
                </template>
                <template v-else-if="item.icon === 'plan'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </template>
                <template v-else-if="item.icon === 'knowledge'">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 18v-5.25m0 0a6.01 6.01 0 001.5-.189m-1.5.189a6.01 6.01 0 01-1.5-.189m3.75 7.478a12.06 12.06 0 01-4.5 0m3.75 2.383a14.406 14.406 0 01-3 0M14.25 18v-.192c0-.983.658-1.823 1.508-2.316a7.5 7.5 0 10-7.517 0c.85.493 1.509 1.333 1.509 2.316V18" />
                </template>
              </svg>
              <span>{{ item.label }}</span>
            </button>
          </div>
        </div>
      </nav>

      <div class="border-t border-gray-100 p-3">
        <div class="flex items-center gap-2 rounded-xl border border-gray-100 bg-gray-50/80 px-3 py-2.5">
          <div class="bg-brand-gradient flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-sm font-semibold text-white shadow-sm">
            {{ (auth.user?.nickname || auth.user?.username || '?').slice(0, 1).toUpperCase() }}
          </div>
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-gray-800">{{ auth.user?.nickname || auth.user?.username }}</p>
            <p class="truncate text-xs text-gray-400">{{ currentSubjectName || currentLevelName || auth.user?.email }}</p>
          </div>
          <button class="cursor-pointer text-gray-400 transition-colors hover:text-red-500" title="退出登录" @click="auth.logout()">
            <svg class="h-4.5 w-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75" />
            </svg>
          </button>
        </div>
      </div>
    </aside>

    <!-- 移动端抽屉 -->
    <Transition name="drawer-fade">
      <div v-if="mobileOpen" class="fixed inset-0 z-50 bg-gray-900/50 backdrop-blur-sm lg:hidden" @click="closeMobile" />
    </Transition>
    <Transition name="drawer-slide">
      <aside v-if="mobileOpen" class="fixed inset-y-0 left-0 z-50 flex w-72 flex-col border-r border-gray-200 bg-white lg:hidden">
        <div class="flex h-16 items-center justify-between border-b border-gray-100 px-5">
          <div class="flex items-center gap-2">
            <div class="bg-brand-gradient flex h-8 w-8 items-center justify-center rounded-lg text-white shadow-sm">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </svg>
            </div>
            <span class="text-lg font-bold tracking-tight text-gray-900">软考学系</span>
          </div>
          <button class="cursor-pointer text-gray-400 hover:text-gray-600" @click="closeMobile">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <nav class="flex-1 space-y-5 overflow-y-auto px-3 py-5">
          <div v-for="group in navGroups" :key="group.title">
            <p class="mb-1.5 px-3 text-xs font-semibold uppercase tracking-wider text-gray-400">{{ group.title }}</p>
            <div class="space-y-0.5">
              <button
                v-for="item in group.items"
                :key="item.to"
                class="flex w-full cursor-pointer items-center gap-3 rounded-xl px-3 py-2 text-sm font-medium transition-all duration-200"
                :class="isActive(item.to)
                  ? 'bg-brand-gradient text-white shadow-md shadow-indigo-600/20'
                  : 'text-gray-600 hover:bg-indigo-50/60 hover:text-indigo-700'"
                @click="navigate(item.to)"
              >
                <span>{{ item.label }}</span>
              </button>
            </div>
          </div>
        </nav>
        <div class="border-t border-gray-100 p-3">
          <button class="flex w-full cursor-pointer items-center justify-center gap-2 rounded-lg bg-red-50 px-3 py-2 text-sm font-medium text-red-600" @click="auth.logout()">
            退出登录
          </button>
        </div>
      </aside>
    </Transition>

    <!-- 主内容 -->
    <div class="lg:pl-64">
      <header class="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-gray-200 bg-white/90 px-4 backdrop-blur sm:px-6">
        <div class="flex items-center gap-3">
          <button class="cursor-pointer rounded-lg p-2 text-gray-500 hover:bg-gray-100 lg:hidden" @click="mobileOpen = true">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>

          <div ref="subjectNavMenuRef" class="relative">
            <button
              class="group flex cursor-pointer items-center gap-2 rounded-xl border bg-white py-1.5 pl-2.5 pr-2 text-sm transition-all duration-200"
              :class="subjectNavOpen
                ? 'border-indigo-300 shadow-sm ring-2 ring-indigo-500/15'
                : 'border-gray-200 hover:border-indigo-200 hover:shadow-sm'"
              @click.stop="openSubjectNav"
            >
              <span class="flex h-6 w-6 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500 to-violet-500 text-white shadow-sm">
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4.26 10.147a60.438 60.438 0 0 0-.491 6.347A48.62 48.62 0 0 1 12 20.904a48.62 48.62 0 0 1 8.232-4.41 60.46 60.46 0 0 0-.491-6.347m-15.482 0a50.636 50.636 0 0 0-2.658-.813A59.906 59.906 0 0 1 12 3.493a59.903 59.903 0 0 1 10.399 5.84c-.896.248-1.783.52-2.658.814m-15.482 0A50.717 50.717 0 0 1 12 13.489a50.702 50.702 0 0 1 7.74-3.342" />
                </svg>
              </span>
              <span v-if="currentLevelName" class="font-medium text-gray-600">{{ currentLevelName }}</span>
              <span v-if="currentLevelName" class="text-gray-300">/</span>
              <span v-if="currentSubjectName" class="max-w-[10rem] truncate font-semibold text-gray-900">{{ currentSubjectName }}</span>
              <span v-else class="text-sm text-gray-400">选择科目</span>
              <svg
                class="h-3.5 w-3.5 text-gray-400 transition-transform duration-200"
                :class="subjectNavOpen ? 'rotate-180 text-indigo-500' : ''"
                fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
              </svg>
            </button>

            <Transition name="nav-drop">
              <div
                v-if="subjectNavOpen"
                class="absolute left-0 top-full z-50 mt-2 w-80 overflow-hidden rounded-2xl border border-gray-100 bg-white shadow-2xl ring-1 ring-black/5"
              >
                <!-- 头部 -->
                <div class="flex items-center justify-between border-b border-gray-100 bg-gradient-to-r from-indigo-50/70 via-white to-violet-50/50 px-4 py-3">
                  <div>
                    <p class="text-xs font-semibold text-indigo-600">科目导航</p>
                    <p class="mt-0.5 text-[11px] text-gray-500">选择其他等级 / 科目后自动跳转</p>
                  </div>
                  <span class="rounded-full bg-white px-2 py-0.5 text-[10px] font-medium text-gray-500 ring-1 ring-gray-200">
                    共 {{ subjectStore.levels.length }} 个等级
                  </span>
                </div>

                <!-- 等级 + 科目列表 -->
                <div class="max-h-[28rem] overflow-y-auto py-2">
                  <div v-if="!subjectStore.levels.length" class="px-4 py-8 text-center text-sm text-gray-400">
                    暂无可选等级
                  </div>
                  <div v-for="level in subjectStore.levels" :key="level.id" class="px-2 pb-1">
                    <div class="sticky top-0 z-10 mb-0.5 flex items-center gap-2 bg-white/95 px-2 py-1.5 backdrop-blur">
                      <span class="rounded-md bg-indigo-50 px-1.5 py-0.5 text-[11px] font-semibold tracking-wide text-indigo-600 ring-1 ring-inset ring-indigo-200">
                        {{ level.name }}
                      </span>
                      <span v-if="subjectsByLevel[level.id]" class="text-[10px] text-gray-400">
                        {{ subjectsByLevel[level.id].length }} 个科目
                      </span>
                    </div>
                    <div v-if="loadingLevels.has(level.id)" class="space-y-1 px-2 py-1">
                      <div class="h-7 rounded-md bg-gray-50"></div>
                      <div class="h-7 w-4/5 rounded-md bg-gray-50"></div>
                    </div>
                    <div
                      v-else-if="!subjectsByLevel[level.id] || subjectsByLevel[level.id].length === 0"
                      class="px-2 py-2 text-[11px] text-gray-400"
                    >
                      暂无科目
                    </div>
                    <div v-else class="space-y-0.5">
                      <button
                        v-for="sub in subjectsByLevel[level.id]"
                        :key="sub.id"
                        class="group flex w-full cursor-pointer items-center justify-between gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors"
                        :class="sub.id === auth.selectedSubjectId
                          ? 'bg-gradient-to-r from-indigo-50 to-violet-50/60 font-semibold text-indigo-700 ring-1 ring-inset ring-indigo-200'
                          : 'text-gray-700 hover:bg-indigo-50/60 hover:text-indigo-700'"
                        @click="pickSubject(level.id, sub.id)"
                      >
                        <span class="flex min-w-0 items-center gap-2">
                          <span
                            class="flex h-5 w-5 shrink-0 items-center justify-center rounded-md text-[10px] font-bold"
                            :class="sub.id === auth.selectedSubjectId ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-500 group-hover:bg-white group-hover:text-indigo-500'"
                          >
                            {{ (sub.short_name || sub.name).slice(0, 1) }}
                          </span>
                          <span class="truncate">{{ sub.short_name || sub.name }}</span>
                        </span>
                        <svg
                          v-if="sub.id === auth.selectedSubjectId"
                          class="h-4 w-4 shrink-0 text-indigo-600"
                          fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"
                        >
                          <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </Transition>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <router-link
            to="/practice/random"
            class="bg-brand-gradient hidden items-center gap-1.5 rounded-xl px-3.5 py-1.5 text-sm font-medium text-white shadow-sm transition-all duration-200 hover:shadow-md hover:brightness-110 sm:flex"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            开始练习
          </router-link>

          <NotificationBell />

          <div ref="userMenuRef" class="relative">
            <button
              class="flex cursor-pointer items-center gap-2 rounded-xl py-1 pl-1 pr-2 transition-colors hover:bg-gray-100"
              @click.stop="userOpen = !userOpen"
            >
              <span class="bg-brand-gradient flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold text-white shadow-sm">
                {{ (auth.user?.nickname || auth.user?.username || '?').slice(0, 1).toUpperCase() }}
              </span>
              <span class="hidden text-sm font-medium text-gray-700 sm:block">{{ auth.user?.nickname || auth.user?.username }}</span>
              <svg class="h-3 w-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
            </button>
            <div v-if="userOpen" class="absolute right-0 top-full z-50 mt-1 w-44 overflow-hidden rounded-xl border border-gray-200 bg-white py-1 shadow-lg">
              <router-link
                to="/profile"
                class="block px-4 py-2 text-sm text-gray-600 hover:bg-gray-50"
                @click="userOpen = false"
              >个人中心</router-link>
              <router-link
                to="/progress"
                class="block px-4 py-2 text-sm text-gray-600 hover:bg-gray-50"
                @click="userOpen = false"
              >学习进度</router-link>
              <button
                class="block w-full cursor-pointer px-4 py-2 text-left text-sm text-red-500 hover:bg-red-50"
                @click="auth.logout()"
              >退出登录</button>
            </div>
          </div>
        </div>
      </header>

      <main class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
        <router-view v-slot="{ Component }">
          <component :is="Component" :key="route.path" class="animate-page-in" />
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped>
.drawer-fade-enter-active,
.drawer-fade-leave-active {
  transition: opacity 0.2s ease;
}
.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}
.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition: transform 0.25s ease;
}
.drawer-slide-enter-from,
.drawer-slide-leave-to {
  transform: translateX(-100%);
}
.nav-drop-enter-active,
.nav-drop-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.nav-drop-enter-from,
.nav-drop-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
