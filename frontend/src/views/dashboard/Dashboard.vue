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
    <div class="relative mb-6 overflow-hidden rounded-lg bg-indigo-600 p-6 text-white shadow-sm sm:p-8">
      <div class="absolute -right-10 -top-10 h-40 w-40 rounded-full bg-white/10" />
      <div class="absolute -bottom-16 right-24 h-48 w-48 rounded-full bg-white/10" />
      <div class="relative">
        <h1 class="text-xl font-bold sm:text-2xl">{{ greeting }}，{{ auth.user?.nickname || auth.user?.username }} 👋</h1>
        <p class="mt-1 text-sm text-indigo-200">
          {{ auth.user?.level_name || '备考中' }} · {{ auth.user?.subject_name || '未选择科目' }}
        </p>
        <div class="mt-4 flex flex-wrap gap-2">
          <button
            class="cursor-pointer rounded-lg bg-white px-4 py-2 text-sm font-medium text-indigo-600 transition-colors hover:bg-indigo-50"
            @click="router.push('/practice/random')"
          >
            开始练习
          </button>
          <button
            class="cursor-pointer rounded-lg border border-white/40 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-white/10"
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

    <div v-else-if="error" class="rounded-lg bg-white p-6 shadow-sm">
      <p class="text-red-500">{{ error }}</p>
    </div>

    <template v-else-if="overview">
      <!-- 概览卡片 -->
      <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <div class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <p class="text-2xl font-bold text-indigo-600">{{ overview.total_practiced }}</p>
          <p class="mt-1 text-xs text-gray-500">累计答题</p>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <p class="text-2xl font-bold" :class="accuracyClass(overview.accuracy)">{{ overview.accuracy.toFixed(1) }}%</p>
          <p class="mt-1 text-xs text-gray-500">总体正确率</p>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <p class="text-2xl font-bold text-emerald-600">{{ overview.study_days }}</p>
          <p class="mt-1 text-xs text-gray-500">累计学习天数</p>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <p class="text-2xl font-bold text-red-500">{{ overview.wrong_count }}</p>
          <p class="mt-1 text-xs text-gray-500">待复习错题</p>
        </div>
      </div>

      <!-- 快速入口 -->
      <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <button
          v-for="action in quickActions"
          :key="action.to"
          class="group flex cursor-pointer items-center gap-3 rounded-lg border border-gray-200 bg-white p-4 text-left shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md"
          @click="router.push(action.to)"
        >
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg" :class="action.icon">
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
        <div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm lg:col-span-3">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-base font-semibold text-gray-900">近 14 天答题趋势</h2>
            <span class="text-xs text-gray-400">每日练习量</span>
          </div>
          <BarChart :data="chartData" :height="160" />
        </div>

        <!-- 科目掌握度 -->
        <div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm lg:col-span-2">
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
        <div class="flex items-center gap-4 rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <div class="flex h-11 w-11 items-center justify-center rounded-lg bg-indigo-50">
            <svg class="h-6 w-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div>
            <p class="text-2xl font-bold text-gray-900">{{ overview.total_exams }}</p>
            <p class="text-xs text-gray-400">已完成模拟考试</p>
          </div>
        </div>
        <div class="flex items-center gap-4 rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <div class="flex h-11 w-11 items-center justify-center rounded-lg bg-emerald-50">
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
