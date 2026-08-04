<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getStatsOverview, getDailyStats, getSubjectProgress } from '@/api/stats'
import BarChart from '@/components/charts/BarChart.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import type { StatsOverviewResp, DailyStatsResp, SubjectProgressResp } from '@/types/stats'

const router = useRouter()
const auth = useAuthStore()
const overview = ref<StatsOverviewResp | null>(null)
const dailyStats = ref<DailyStatsResp[]>([])
const subjectProgress = ref<SubjectProgressResp[]>([])
const loading = ref(true)
const error = ref('')

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '夜深了'
  if (hour < 12) return '早上好'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const chartData = computed(() => {
  const now = new Date()
  const map = new Map(dailyStats.value.map(d => [d.date, d]))
  const days: { label: string; value: number; hint: string }[] = []
  for (let i = 13; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const key = d.toISOString().split('T')[0]
    const entry = map.get(key)
    const count = entry?.total_count || 0
    days.push({
      label: `${d.getMonth() + 1}/${d.getDate()}`,
      value: count,
      hint: '题',
    })
  }
  return days
})

const quickActions = [
  { label: '随机练习', desc: '综合检测', to: '/practice/random', icon: 'bg-indigo-50 text-indigo-600', path: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.77 4.769L4 5.5m0 0H4m0 0V4m10.5 15.5h-.582m-15.356-2a8.001 8.001 0 0015.356 1.731L20 18.5m0 0V19m0 0h.5m0 0V14' },
  { label: '专项练习', desc: '题型突破', to: '/practice/special', icon: 'bg-emerald-50 text-emerald-600', path: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' },
  { label: '模拟考试', desc: '全真模拟', to: '/exam', icon: 'bg-purple-50 text-purple-600', path: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
  { label: '错题本', desc: '巩固弱项', to: '/wrong-questions', icon: 'bg-red-50 text-red-600', path: 'M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
]

onMounted(async () => {
  try {
    const [ov, daily, subj] = await Promise.all([
      getStatsOverview(),
      getDailyStats(14),
      getSubjectProgress(),
    ])
    overview.value = ov
    dailyStats.value = daily || []
    subjectProgress.value = subj || []
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

function accuracyClass(a: number) {
  if (a >= 80) return 'text-emerald-500'
  if (a >= 60) return 'text-indigo-600'
  return 'text-yellow-500'
}

function accuracyBarClass(a: number) {
  if (a >= 80) return 'bg-emerald-500'
  if (a >= 60) return 'bg-indigo-500'
  return 'bg-yellow-500'
}
</script>

<template>
  <div>
    <!-- 欢迎横幅 -->
    <div class="bg-brand-gradient relative mb-6 overflow-hidden rounded-2xl p-6 text-white shadow-lg shadow-indigo-600/20 sm:p-8">
      <div class="absolute -right-10 -top-10 h-40 w-40 rounded-full bg-white/10" />
      <div class="absolute -bottom-16 right-24 h-48 w-48 rounded-full bg-white/10" />
      <div class="absolute -left-8 top-1/2 h-24 w-24 -translate-y-1/2 rounded-full bg-white/5" />
      <div class="relative">
        <h1 class="text-xl font-bold tracking-tight sm:text-2xl">{{ greeting }}，{{ auth.user?.nickname || auth.user?.username }} 👋</h1>
        <p class="mt-1.5 flex items-center gap-2 text-sm text-indigo-100">
          <span class="inline-flex items-center gap-1 rounded-full bg-white/15 px-2.5 py-0.5 text-xs font-medium">
            {{ auth.user?.level_name || '备考中' }}
          </span>
          <span>{{ auth.user?.subject_name || '未选择科目' }}</span>
        </p>
        <div class="mt-5 flex flex-wrap gap-2.5">
          <button
            class="cursor-pointer rounded-xl bg-white px-4 py-2 text-sm font-semibold text-indigo-600 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:shadow-md"
            @click="router.push('/practice/random')"
          >
            开始练习
          </button>
          <button
            class="cursor-pointer rounded-xl border border-white/40 px-4 py-2 text-sm font-medium text-white transition-all duration-200 hover:bg-white/15"
            @click="router.push('/exam')"
          >
            模拟考试
          </button>
        </div>
      </div>
    </div>

    <div v-if="loading">
      <BaseLoading />
    </div>

    <div v-else-if="error" class="rounded-xl border border-red-100 bg-red-50 p-6">
      <p class="text-sm text-red-600">{{ error }}</p>
    </div>

    <template v-else-if="overview">
      <!-- 概览卡片 -->
      <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <div class="card-lift rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-50">
            <svg class="h-5 w-5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
            </svg>
          </div>
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ overview.total_practiced }}</p>
          <p class="mt-1 text-xs text-gray-400">累计答题</p>
        </div>
        <div class="card-lift rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-50">
            <svg class="h-5 w-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <p class="text-2xl font-bold tracking-tight" :class="accuracyClass(overview.accuracy)">{{ overview.accuracy.toFixed(1) }}%</p>
          <p class="mt-1 text-xs text-gray-400">总体正确率</p>
        </div>
        <div class="card-lift rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-violet-50">
            <svg class="h-5 w-5 text-violet-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
            </svg>
          </div>
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ overview.study_days }}</p>
          <p class="mt-1 text-xs text-gray-400">累计学习天数</p>
        </div>
        <div class="card-lift rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-red-50">
            <svg class="h-5 w-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
            </svg>
          </div>
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ overview.wrong_count }}</p>
          <p class="mt-1 text-xs text-gray-400">待复习错题</p>
        </div>
      </div>

      <!-- 快速入口 -->
      <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <button
          v-for="action in quickActions"
          :key="action.to"
          class="card-lift group flex cursor-pointer items-center gap-3 rounded-xl border border-gray-100 bg-white p-4 text-left shadow-sm"
          @click="router.push(action.to)"
        >
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl transition-transform duration-200 group-hover:scale-110" :class="action.icon">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" :d="action.path" />
            </svg>
          </div>
          <div class="min-w-0">
            <p class="text-sm font-semibold text-gray-800">{{ action.label }}</p>
            <p class="text-xs text-gray-400">{{ action.desc }}</p>
          </div>
        </button>
      </div>

      <div class="grid gap-6 lg:grid-cols-5">
        <!-- 近14天趋势 -->
        <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm lg:col-span-3">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-base font-semibold tracking-tight text-gray-900">近 14 天答题趋势</h2>
            <span class="text-xs text-gray-400">每日练习量</span>
          </div>
          <BarChart :data="chartData" :height="160" />
        </div>

        <!-- 科目掌握度 -->
        <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm lg:col-span-2">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-base font-semibold text-gray-900">科目掌握度</h2>
            <button class="cursor-pointer text-xs text-indigo-600 hover:underline" @click="router.push('/progress')">查看详情</button>
          </div>
          <div v-if="subjectProgress.length === 0" class="py-8 text-center text-sm text-gray-400">
            暂无练习数据，快去刷题吧
          </div>
          <div v-else class="space-y-4">
            <div v-for="sp in subjectProgress.slice(0, 4)" :key="sp.subject_id">
              <div class="mb-1 flex items-center justify-between text-sm">
                <span class="truncate font-medium text-gray-700">{{ sp.subject_name }}</span>
                <span class="ml-2 shrink-0 font-semibold" :class="accuracyClass(sp.accuracy)">{{ sp.accuracy.toFixed(0) }}%</span>
              </div>
              <div class="h-2 overflow-hidden rounded-full bg-gray-100">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="accuracyBarClass(sp.accuracy)"
                  :style="{ width: Math.min(100, sp.accuracy) + '%' }"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 模考统计 -->
      <div v-if="overview.total_exams > 0" class="mt-6 grid gap-4 sm:grid-cols-2">
        <div class="card-lift flex items-center gap-4 rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-indigo-50">
            <svg class="h-6 w-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div>
            <p class="text-2xl font-bold text-gray-900">{{ overview.total_exams }}</p>
            <p class="text-xs text-gray-400">已完成模拟考试</p>
          </div>
        </div>
        <div class="card-lift flex items-center gap-4 rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-emerald-50">
            <svg class="h-6 w-6 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625z" />
            </svg>
          </div>
          <div>
            <p class="text-2xl font-bold text-gray-900">{{ overview.avg_exam_score.toFixed(1) }}</p>
            <p class="text-xs text-gray-400">平均分</p>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
