<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listExamRecords } from '@/api/exam'
import { formatDuration, formatPercent, timeAgo } from '@/utils/format'
import LineChart from '@/components/charts/LineChart.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseTabs from '@/components/common/BaseTabs.vue'
import BaseProgressBar from '@/components/common/BaseProgressBar.vue'
import type { ExamRecordResp } from '@/types/exam'

type StatusFilter = 'all' | 'finished' | 'in_progress'

const router = useRouter()
const records = ref<ExamRecordResp[]>([])
const loading = ref(true)
const error = ref('')
const statusFilter = ref<StatusFilter>('all')

async function load() {
  loading.value = true
  error.value = ''
  try {
    records.value = await listExamRecords()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(load)

const sortedRecords = computed(() =>
  [...records.value].sort((a, b) => {
    const ta = new Date(a.started_at || a.created_at || 0).getTime()
    const tb = new Date(b.started_at || b.created_at || 0).getTime()
    return tb - ta
  }),
)

const finishedRecords = computed(() => sortedRecords.value.filter(r => r.status === 'finished'))
const finishedCount = computed(() => finishedRecords.value.length)
const inProgressCount = computed(() => sortedRecords.value.filter(r => r.status === 'in_progress').length)

const displayedRecords = computed(() => {
  if (statusFilter.value === 'all') return sortedRecords.value
  return sortedRecords.value.filter(r => r.status === statusFilter.value)
})

const latestRecord = computed(() => finishedRecords.value[0] || null)

function percent(score: number, total: number) {
  if (!total) return 0
  return Math.round((score / total) * 100)
}

function shortDate(s: string) {
  if (!s) return ''
  const m = s.match(/\d{4}[-/](\d{1,2})[-/](\d{1,2})/)
  return m ? `${m[1].padStart(2, '0')}-${m[2].padStart(2, '0')}` : s.slice(5)
}

function fullDate(s: string) {
  if (!s) return ''
  const d = new Date(s.replace(' ', 'T'))
  if (isNaN(d.getTime())) return s
  return `${d.getMonth() + 1} 月 ${d.getDate()} 日 ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const trendData = computed(() =>
  finishedRecords.value
    .map(r => ({
      label: shortDate(r.started_at || ''),
      value: percent(r.score, r.total_score),
      hint: `正确率 ${formatPercent(r.accuracy)}`,
    }))
    .reverse(),
)

// 成绩趋势范围筛选：全部 / 近 10 次 / 近 5 次
type TrendRange = 'all' | 10 | 5
const trendRange = ref<TrendRange>('all')
const trendRangeTabs = [
  { label: '全部', value: 'all' as const },
  { label: '近 10 次', value: 10 as const },
  { label: '近 5 次', value: 5 as const },
]

const rangedTrendData = computed(() => {
  const list = trendData.value
  if (trendRange.value === 'all' || list.length <= trendRange.value) return list
  return list.slice(-trendRange.value)
})

const rangedTrendStats = computed(() => {
  const list = rangedTrendData.value
  if (!list.length) return { count: 0, avg: 0, best: 0, worst: 0, lastDelta: 0 }
  const values = list.map(d => d.value)
  const avg = Math.round(values.reduce((s, v) => s + v, 0) / values.length)
  const best = Math.max(...values)
  const worst = Math.min(...values)
  const lastDelta = values.length >= 2 ? values[values.length - 1] - values[values.length - 2] : 0
  return { count: list.length, avg, best, worst, lastDelta }
})

const trendStats = computed(() => {
  const list = trendData.value
  if (!list.length) return { count: 0, avg: 0, best: 0, worst: 0, passRate: 0, lastDelta: 0 }
  const values = list.map(d => d.value)
  const passCount = values.filter(v => v >= 60).length
  return {
    count: list.length,
    avg: Math.round(values.reduce((s, v) => s + v, 0) / values.length),
    best: Math.max(...values),
    worst: Math.min(...values),
    passRate: Math.round((passCount / values.length) * 100),
    lastDelta: values.length >= 2 ? values[values.length - 1] - values[values.length - 2] : 0,
  }
})

const latestPct = computed(() => {
  if (!latestRecord.value) return 0
  return percent(latestRecord.value.score, latestRecord.value.total_score)
})

const latestGrade = computed(() => {
  const p = latestPct.value
  if (p >= 85) return { label: '优秀', ring: '#10b981', bg: 'from-emerald-500 via-teal-500 to-cyan-500' }
  if (p >= 70) return { label: '良好', ring: '#6366f1', bg: 'from-indigo-500 via-violet-500 to-purple-500' }
  if (p >= 60) return { label: '及格', ring: '#f59e0b', bg: 'from-amber-500 via-orange-500 to-rose-400' }
  return { label: '加油', ring: '#ef4444', bg: 'from-rose-500 via-red-500 to-pink-500' }
})

function accuracyTextClass(a: number) {
  if (a >= 80) return 'text-emerald-600'
  if (a >= 60) return 'text-indigo-600'
  return 'text-red-500'
}

function accuracyColor(a: number): 'emerald' | 'indigo' | 'red' {
  if (a >= 80) return 'emerald'
  if (a >= 60) return 'indigo'
  return 'red'
}

function statusBadge(r: ExamRecordResp) {
  if (r.status === 'finished') return { type: 'success' as const, label: '已完成' }
  if (r.status === 'in_progress') return { type: 'warning' as const, label: '进行中' }
  return { type: 'default' as const, label: r.status }
}

function diffWithPrev(r: ExamRecordResp): { value: number; hasPrev: boolean } {
  if (r.status !== 'finished') return { value: 0, hasPrev: false }
  const idx = finishedRecords.value.findIndex(x => x.id === r.id)
  if (idx < 0 || idx >= finishedRecords.value.length - 1) return { value: 0, hasPrev: false }
  const prev = finishedRecords.value[idx + 1]
  return { value: r.accuracy - prev.accuracy, hasPrev: true }
}

function goDetail(r: ExamRecordResp) {
  if (r.status === 'finished') {
    router.push(`/exam/${r.id}/result`)
  } else {
    router.push(`/exam/${r.template_id}`)
  }
}

const SCORE_R = 70
const scoreCircumference = 2 * Math.PI * SCORE_R
const scoreCircleOffset = computed(() => scoreCircumference * (1 - latestPct.value / 100))

const miniRingCircumference = 2 * Math.PI * 11
const miniRingOffset = computed(() => miniRingCircumference * (1 - trendStats.value.passRate / 100))

function buildSparkline(values: number[], w = 100, h = 28) {
  if (values.length < 2) return { line: '', area: '' }
  const max = Math.max(...values)
  const min = Math.min(...values)
  const range = max - min || 1
  const step = w / (values.length - 1)
  const points = values.map((v, i) => {
    const x = i * step
    const y = h - ((v - min) / range) * h * 0.85 - h * 0.075
    return [x, y] as const
  })
  const line = points.map(([x, y], i) => `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`).join(' ')
  const area = `${line} L${w},${h} L0,${h} Z`
  return { line, area }
}

const sparkline = computed(() => buildSparkline(trendData.value.slice(-10).map(d => d.value)))

const weeklyCount = computed(() => {
  const now = Date.now()
  return sortedRecords.value.filter(r => {
    const t = new Date(r.started_at || r.created_at || 0).getTime()
    return now - t <= 7 * 24 * 60 * 60 * 1000
  }).length
})
</script>

<template>
  <div>
    <BasePageHeader title="考试记录" subtitle="回顾每一次模拟考试的成绩，洞察成长轨迹">
      <template #actions>
        <BaseButton @click="router.push('/exam')">去考试</BaseButton>
      </template>
    </BasePageHeader>

    <BaseSkeleton v-if="loading" variant="detail" />

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重新加载</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else-if="records.length === 0" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="暂无考试记录" description="完成一次模拟考试后，记录会展示在这里">
        <template #action>
          <BaseButton size="sm" class="mt-4" @click="router.push('/exam')">去参加考试</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else class="space-y-6">
      <!-- BENTO: Hero + KPI -->
      <div class="grid grid-cols-1 gap-4 lg:grid-cols-12">
        <!-- 最近一次考试 英雄卡 -->
        <BaseCard v-if="latestRecord" class="!p-0 overflow-hidden lg:col-span-7">
          <div :class="['relative overflow-hidden bg-gradient-to-br px-6 py-5', latestGrade.bg]">
            <div class="absolute -right-12 -top-12 h-44 w-44 rounded-full bg-white/10" />
            <div class="absolute -bottom-16 right-28 h-36 w-36 rounded-full bg-white/10" />
            <div class="absolute left-1/2 top-1/2 h-24 w-24 -translate-x-1/2 -translate-y-1/2 rounded-full bg-white/5" />
            <div class="relative flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div class="min-w-0 flex-1">
                <p class="text-[10px] font-semibold tracking-[0.2em] text-white/80 uppercase">最近一次考试 · Latest</p>
                <h2 class="mt-1.5 truncate text-2xl font-bold tracking-tight text-white">{{ latestRecord.template_name }}</h2>
                <p class="mt-1 text-xs text-white/80">{{ fullDate(latestRecord.started_at) }} · 用时 {{ formatDuration(latestRecord.duration) }}</p>
              </div>
              <span class="inline-flex shrink-0 items-center gap-1.5 self-start rounded-full bg-white/25 px-3 py-1 text-sm font-bold text-white ring-1 ring-inset ring-white/40 backdrop-blur-sm">
                <svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.562.562 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z" />
                </svg>
                {{ latestGrade.label }}
              </span>
            </div>
          </div>

          <div class="grid grid-cols-1 items-center gap-6 p-6 sm:grid-cols-[auto_1fr] sm:gap-8">
            <!-- 自定义渐变圆环得分 -->
            <div class="relative mx-auto sm:mx-0">
              <svg viewBox="0 0 160 160" class="h-40 w-40">
                <defs>
                  <linearGradient :id="`scoreGrad-${latestRecord.id}`" x1="0" y1="0" x2="1" y2="1">
                    <stop offset="0%" :stop-color="latestGrade.ring" stop-opacity="0.95" />
                    <stop offset="100%" :stop-color="latestGrade.ring" stop-opacity="0.55" />
                  </linearGradient>
                </defs>
                <circle cx="80" cy="80" :r="SCORE_R" fill="none" stroke="#f1f5f9" stroke-width="12" />
                <circle
                  cx="80" cy="80" :r="SCORE_R" fill="none"
                  :stroke="`url(#scoreGrad-${latestRecord.id})`" stroke-width="12" stroke-linecap="round"
                  :stroke-dasharray="scoreCircumference" :stroke-dashoffset="scoreCircleOffset"
                  transform="rotate(-90 80 80)" class="transition-all duration-1000 ease-out"
                />
              </svg>
              <div class="absolute inset-0 flex flex-col items-center justify-center">
                <span class="text-4xl font-bold tracking-tight" :class="accuracyTextClass(latestPct)">{{ latestPct }}</span>
                <span class="mt-0.5 text-[10px] font-medium tracking-wider text-gray-400 uppercase">得分率 / 100</span>
              </div>
            </div>

            <!-- 4 指标 -->
            <div class="grid grid-cols-2 gap-3">
              <div class="rounded-xl border border-gray-100 bg-white p-3.5">
                <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">得分</p>
                <p class="mt-1 text-xl font-bold text-gray-900">{{ latestRecord.score }}<span class="text-xs font-normal text-gray-400"> / {{ latestRecord.total_score }}</span></p>
              </div>
              <div class="rounded-xl border border-gray-100 bg-white p-3.5">
                <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">正确率</p>
                <p class="mt-1 text-xl font-bold" :class="accuracyTextClass(latestRecord.accuracy)">{{ formatPercent(latestRecord.accuracy) }}</p>
              </div>
              <div class="rounded-xl border border-gray-100 bg-white p-3.5">
                <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">答对题数</p>
                <p class="mt-1 text-xl font-bold text-gray-900">{{ latestRecord.correct_count }}<span class="text-xs font-normal text-gray-400"> 题</span></p>
              </div>
              <div class="rounded-xl border border-gray-100 bg-white p-3.5">
                <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">用时</p>
                <p class="mt-1 text-xl font-bold text-gray-900">{{ formatDuration(latestRecord.duration) }}</p>
              </div>
            </div>
          </div>

          <!-- 底部条：对比 + CTA -->
          <div class="flex flex-col gap-3 border-t border-gray-100 bg-gray-50/60 px-6 py-3.5 sm:flex-row sm:items-center sm:justify-between">
            <div class="text-xs text-gray-600">
              <template v-if="trendStats.count >= 2 && trendStats.lastDelta > 0">
                较上一次 <span class="font-bold text-emerald-600">▲ +{{ trendStats.lastDelta }}%</span>，保持节奏
              </template>
              <template v-else-if="trendStats.count >= 2 && trendStats.lastDelta < 0">
                较上一次 <span class="font-bold text-red-500">▼ {{ trendStats.lastDelta }}%</span>，再接再厉
              </template>
              <template v-else-if="trendStats.count >= 2">
                与上一次持平，继续保持
              </template>
              <template v-else>
                这是你的第一次考试记录
              </template>
            </div>
            <BaseButton size="sm" @click="goDetail(latestRecord)">查看完整报告 →</BaseButton>
          </div>
        </BaseCard>

        <!-- KPI Bento 2x2 -->
        <div :class="['grid grid-cols-2 gap-4', latestRecord ? 'lg:col-span-5' : 'lg:col-span-12']">
          <!-- 考试次数 -->
          <BaseCard class="!p-5">
            <div class="flex items-start justify-between">
              <div>
                <p class="text-[10px] font-semibold uppercase tracking-wider text-indigo-500">考试次数</p>
                <p class="mt-2 text-3xl font-bold tracking-tight text-gray-900">{{ trendStats.count }}</p>
                <p class="mt-0.5 text-xs text-gray-400">累计完成</p>
              </div>
              <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
                </svg>
              </div>
            </div>
            <div class="mt-3 flex items-end justify-between border-t border-gray-50 pt-2">
              <span class="text-[10px] text-gray-400">近 7 天</span>
              <span class="text-xs font-semibold text-indigo-600">+{{ weeklyCount }} 次</span>
            </div>
          </BaseCard>

          <!-- 平均得分 -->
          <BaseCard class="!p-5">
            <div class="flex items-start justify-between">
              <div>
                <p class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">平均得分</p>
                <p class="mt-2 text-3xl font-bold tracking-tight text-gray-900">{{ trendStats.avg }}<span class="ml-0.5 text-base font-medium text-gray-400">%</span></p>
                <p class="mt-0.5 text-xs text-gray-400">所有考试</p>
              </div>
              <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
                </svg>
              </div>
            </div>
            <svg v-if="sparkline.line" viewBox="0 0 100 28" class="mt-2 h-7 w-full" preserveAspectRatio="none">
              <path :d="sparkline.area" :fill="latestGrade.ring" fill-opacity="0.15" />
              <path :d="sparkline.line" :stroke="latestGrade.ring" stroke-width="1.5" fill="none" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
            <div v-else class="mt-2 h-7 rounded-md bg-gray-50" />
          </BaseCard>

          <!-- 最高分 -->
          <BaseCard class="!p-5">
            <div class="flex items-start justify-between">
              <div>
                <p class="text-[10px] font-semibold uppercase tracking-wider text-emerald-500">最高分</p>
                <p class="mt-2 text-3xl font-bold tracking-tight text-emerald-600">{{ trendStats.best }}<span class="ml-0.5 text-base font-medium text-emerald-400">%</span></p>
                <p class="mt-0.5 text-xs text-gray-400">个人最佳</p>
              </div>
              <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18L9 11.25l4.306 4.307a11.95 11.95 0 015.814-5.519l2.74-1.22m0 0l-5.94-2.28m5.94 2.28l-2.28 5.941" />
                </svg>
              </div>
            </div>
            <div class="mt-3 flex items-center gap-1.5 border-t border-gray-50 pt-2 text-[10px] text-gray-400">
              <svg class="h-3.5 w-3.5 text-emerald-500" fill="currentColor" viewBox="0 0 24 24">
                <path d="M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.562.562 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z" />
              </svg>
              <span>继续突破新高</span>
            </div>
          </BaseCard>

          <!-- 通过率 -->
          <BaseCard class="!p-5">
            <div class="flex items-start justify-between">
              <div>
                <p class="text-[10px] font-semibold uppercase tracking-wider text-violet-500">通过率</p>
                <p class="mt-2 text-3xl font-bold tracking-tight text-violet-600">{{ trendStats.passRate }}<span class="ml-0.5 text-base font-medium text-violet-400">%</span></p>
                <p class="mt-0.5 text-xs text-gray-400">≥ 60 分占比</p>
              </div>
              <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-violet-50 text-violet-600">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
            <div class="mt-3 flex items-center gap-2.5 border-t border-gray-50 pt-2">
              <div class="relative h-7 w-7 shrink-0">
                <svg viewBox="0 0 28 28" class="h-7 w-7 -rotate-90">
                  <circle cx="14" cy="14" r="11" fill="none" stroke="#ede9fe" stroke-width="4" />
                  <circle cx="14" cy="14" r="11" fill="none" stroke="#8b5cf6" stroke-width="4" stroke-linecap="round"
                    :stroke-dasharray="miniRingCircumference" :stroke-dashoffset="miniRingOffset"
                    class="transition-all duration-700"
                  />
                </svg>
                <span class="absolute inset-0 flex items-center justify-center text-[8px] font-bold text-violet-600">{{ trendStats.passRate }}</span>
              </div>
              <div class="flex-1">
                <div class="h-1.5 overflow-hidden rounded-full bg-gray-100">
                  <div class="h-full rounded-full bg-gradient-to-r from-violet-500 to-purple-500 transition-all duration-500" :style="{ width: trendStats.passRate + '%' }" />
                </div>
                <p class="mt-1 text-[10px] text-gray-400">已通过 {{ Math.round(trendStats.count * trendStats.passRate / 100) }} / {{ trendStats.count }} 次</p>
              </div>
            </div>
          </BaseCard>
        </div>
      </div>

      <!-- 成绩趋势 -->
      <BaseCard v-if="finishedRecords.length" class="!p-0 overflow-hidden">
        <!-- 头部：标题 + 范围筛选 + 更多链接 -->
        <div class="flex flex-col gap-3 p-5 pb-4 sm:p-6 sm:pb-4 lg:flex-row lg:items-center lg:justify-between">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-500 to-violet-500 text-white shadow-sm shadow-indigo-500/25">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
              </svg>
            </div>
            <div>
              <h2 class="text-base font-semibold tracking-tight text-gray-900">成绩趋势</h2>
              <p class="mt-0.5 text-xs text-gray-500">最近 {{ finishedRecords.length }} 次考试得分率走势</p>
            </div>
          </div>

          <div class="flex flex-wrap items-center gap-3 self-start lg:self-auto">
            <BaseTabs
              v-model="trendRange"
              variant="pill"
              size="sm"
              :tabs="trendRangeTabs"
            />
            <button class="cursor-pointer text-xs font-medium text-indigo-600 transition-colors hover:text-indigo-700" @click="router.push('/progress')">
              查看更多趋势 →
            </button>
          </div>
        </div>

        <!-- 区间指标条 -->
        <div class="grid grid-cols-2 gap-px border-t border-gray-50 bg-gray-50 sm:grid-cols-4">
          <div class="bg-white px-5 py-3.5">
            <div class="flex items-center gap-1.5">
              <svg class="h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75z" />
              </svg>
              <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">平均得分率</p>
            </div>
            <p class="mt-1 text-xl font-bold tracking-tight text-gray-900">{{ rangedTrendStats.avg }}<span class="text-sm font-medium text-gray-400">%</span></p>
          </div>
          <div class="bg-white px-5 py-3.5">
            <div class="flex items-center gap-1.5">
              <svg class="h-3.5 w-3.5 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18L9 11.25l4.306 4.307a11.95 11.95 0 015.814-5.519l2.74-1.22m0 0l-5.94-2.28m5.94 2.28l-2.28 5.941" />
              </svg>
              <p class="text-[10px] font-semibold uppercase tracking-wider text-emerald-500">区间最佳</p>
            </div>
            <p class="mt-1 text-xl font-bold tracking-tight text-emerald-600">{{ rangedTrendStats.best }}<span class="text-sm font-medium text-emerald-400">%</span></p>
          </div>
          <div class="bg-white px-5 py-3.5">
            <div class="flex items-center gap-1.5">
              <svg class="h-3.5 w-3.5 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9.75v3.75m0 3h.008v.008H12v-.008zM21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p class="text-[10px] font-semibold uppercase tracking-wider text-rose-500">区间最低</p>
            </div>
            <p class="mt-1 text-xl font-bold tracking-tight text-rose-500">{{ rangedTrendStats.worst }}<span class="text-sm font-medium text-rose-400">%</span></p>
          </div>
          <div class="bg-white px-5 py-3.5">
            <div class="flex items-center gap-1.5">
              <svg class="h-3.5 w-3.5" :class="rangedTrendStats.lastDelta > 0 ? 'text-emerald-500' : rangedTrendStats.lastDelta < 0 ? 'text-red-500' : 'text-gray-400'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 13.5L12 21m0 0l-7.5-7.5M12 21V3" />
              </svg>
              <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">较上次变化</p>
            </div>
            <p class="mt-1 text-xl font-bold tracking-tight" :class="rangedTrendStats.lastDelta > 0 ? 'text-emerald-600' : rangedTrendStats.lastDelta < 0 ? 'text-red-500' : 'text-gray-500'">
              <template v-if="rangedTrendStats.lastDelta > 0">▲ +{{ rangedTrendStats.lastDelta }}</template>
              <template v-else-if="rangedTrendStats.lastDelta < 0">▼ {{ rangedTrendStats.lastDelta }}</template>
              <template v-else>持平</template>
              <span class="text-sm font-medium text-gray-400">%</span>
            </p>
          </div>
        </div>

        <!-- 图表 -->
        <div class="relative pt-4">
          <LineChart
            :data="rangedTrendData"
            :height="240"
            gradient
            smooth
            show-values
            :value-label-limit="8"
            interactive
            emphasize-last
            threshold-band
            :threshold="60"
            threshold-label="及格线 60%"
          />
        </div>

        <!-- 图例 -->
        <div class="mt-2 flex flex-wrap items-center justify-center gap-x-4 gap-y-1.5 px-2 pb-5 text-[11px] text-gray-500">
          <span class="flex items-center gap-1.5">
            <span class="h-2 w-2 rounded-full bg-gradient-to-r from-indigo-600 to-violet-600"></span>
            成绩曲线
          </span>
          <span class="flex items-center gap-1.5">
            <span class="h-2 w-2 rounded-full bg-emerald-500"></span>
            及格（≥60%）
          </span>
          <span class="flex items-center gap-1.5">
            <span class="h-2 w-2 rounded-full bg-rose-500"></span>
            未达标（&lt;60%）
          </span>
          <span class="flex items-center gap-1.5">
            <span class="relative inline-flex h-2.5 w-2.5">
              <span class="absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-40"></span>
              <span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-amber-500"></span>
            </span>
            最新一次
          </span>
          <span class="flex items-center gap-1.5">
            <svg class="h-3 w-3" fill="none" viewBox="0 0 12 2"><line x1="0" y1="1" x2="12" y2="1" stroke="#10b981" stroke-width="1" stroke-dasharray="2 2" /></svg>
            及格线 60%
          </span>
        </div>
      </BaseCard>

      <!-- 考试历程：时间线 -->
      <div>
        <div class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 class="text-base font-semibold tracking-tight text-gray-900">考试历程</h2>
            <p class="mt-0.5 text-xs text-gray-500">按时间倒序排列 · 共 {{ displayedRecords.length }} 条记录</p>
          </div>
          <BaseTabs
            v-model="statusFilter"
            variant="pill"
            size="sm"
            :tabs="[
              { label: '全部', value: 'all', badge: sortedRecords.length },
              { label: '已完成', value: 'finished', badge: finishedCount },
              { label: '进行中', value: 'in_progress', badge: inProgressCount },
            ]"
          />
        </div>

        <div v-if="displayedRecords.length === 0" class="rounded-xl border border-gray-100 bg-white shadow-sm">
          <BaseEmpty title="暂无符合条件的考试记录" description="切换其他状态筛选试试">
            <template #action>
              <BaseButton type="secondary" size="sm" class="mt-4" @click="statusFilter = 'all'">查看全部</BaseButton>
            </template>
          </BaseEmpty>
        </div>

        <div v-else class="relative">
          <!-- 时间线主轴：渐变竖线 -->
          <div v-if="displayedRecords.length > 1" class="absolute left-5 top-7 bottom-7 w-px bg-gradient-to-b from-indigo-300 via-gray-200 to-transparent" aria-hidden="true" />

          <div class="space-y-3">
            <div
              v-for="(r, idx) in displayedRecords"
              :key="r.id"
              class="relative flex items-stretch gap-4"
            >
              <!-- 时间线节点 -->
              <div class="relative z-10 flex shrink-0 flex-col items-center pt-2">
                <div
                  class="flex h-10 w-10 items-center justify-center rounded-full ring-4 ring-white shadow-sm transition-transform duration-200"
                  :class="r.status === 'finished'
                    ? (r.accuracy >= 60 ? 'bg-emerald-500 text-white' : 'bg-rose-500 text-white')
                    : 'bg-amber-500 text-white'"
                >
                  <svg v-if="r.status === 'finished'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                  </svg>
                  <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <!-- 最新一条小皇冠 -->
                <svg v-if="idx === 0 && r.status === 'finished'" class="absolute -top-1.5 left-1/2 h-3.5 w-3.5 -translate-x-1/2 text-amber-400" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.562.562 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z" />
                </svg>
              </div>

              <!-- 记录卡 -->
              <BaseCard hover class="flex-1 !p-0">
                <div class="p-4 sm:p-5" @click="goDetail(r)">
                  <!-- 头部：标题 + 徽章 + 对比 -->
                  <div class="flex flex-wrap items-start gap-2">
                    <h3 class="min-w-0 flex-1 truncate text-sm font-semibold text-gray-900">{{ r.template_name }}</h3>
                    <BaseBadge :type="statusBadge(r).type">{{ statusBadge(r).label }}</BaseBadge>
                    <div
                      v-if="diffWithPrev(r).hasPrev"
                      class="inline-flex items-center gap-0.5 rounded-full px-2 py-0.5 text-[10px] font-bold tabular-nums"
                      :class="diffWithPrev(r).value > 0 ? 'bg-emerald-50 text-emerald-600' : diffWithPrev(r).value < 0 ? 'bg-red-50 text-red-500' : 'bg-gray-100 text-gray-500'"
                    >
                      <span v-if="diffWithPrev(r).value > 0">▲</span>
                      <span v-else-if="diffWithPrev(r).value < 0">▼</span>
                      <span v-else>—</span>
                      <span>{{ diffWithPrev(r).value > 0 ? '+' : '' }}{{ diffWithPrev(r).value.toFixed(1) }}%</span>
                    </div>
                  </div>

                  <!-- 元信息 -->
                  <p class="mt-1.5 text-xs text-gray-400">
                    {{ fullDate(r.started_at || r.created_at) }} · 用时 {{ formatDuration(r.duration) }}
                  </p>

                  <!-- 主体：得分 + 正确率 + 进度条 -->
                  <template v-if="r.status === 'finished'">
                    <div class="mt-4 flex items-end justify-between gap-4">
                      <div class="flex items-baseline gap-3">
                        <div>
                          <span class="text-2xl font-bold text-gray-900">{{ r.score }}</span>
                          <span class="text-sm text-gray-400"> / {{ r.total_score }}</span>
                          <p class="mt-0.5 text-[10px] uppercase tracking-wider text-gray-400">得分</p>
                        </div>
                        <div class="h-8 w-px bg-gray-100" />
                        <div>
                          <span class="text-lg font-bold" :class="accuracyTextClass(r.accuracy)">{{ formatPercent(r.accuracy) }}</span>
                          <p class="mt-0.5 text-[10px] uppercase tracking-wider text-gray-400">正确率</p>
                        </div>
                      </div>
                      <BaseButton type="secondary" size="sm" @click.stop="goDetail(r)">查看详情</BaseButton>
                    </div>
                    <div class="mt-3">
                      <BaseProgressBar :value="r.accuracy" :color="accuracyColor(r.accuracy)" size="sm" />
                    </div>
                  </template>

                  <template v-else>
                    <div class="mt-4 flex items-center justify-between">
                      <p class="text-xs text-amber-600">已开始 {{ timeAgo(r.started_at || r.created_at) }}</p>
                      <BaseButton size="sm" @click.stop="goDetail(r)">继续考试</BaseButton>
                    </div>
                  </template>
                </div>
              </BaseCard>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
