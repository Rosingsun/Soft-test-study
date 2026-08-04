<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getStatsOverview, getDailyStats, getSubjectProgress, getCalendarStats, getChapterProgress } from '@/api/stats'
import CalendarHeatmap from '@/components/charts/CalendarHeatmap.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import type { StatsOverviewResp, DailyStatsResp, SubjectProgressResp, CalendarStatsResp, ChapterProgressResp } from '@/types/stats'

const router = useRouter()
const overview = ref<StatsOverviewResp | null>(null)
const dailyStats = ref<DailyStatsResp[]>([])
const subjectProgress = ref<SubjectProgressResp[]>([])
const calendarData = ref<CalendarStatsResp[]>([])
const chapterProgress = ref<ChapterProgressResp[]>([])
const loading = ref(true)
const error = ref('')

const now = new Date()
const currentMonth = ref(`${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`)

const monthTotal = computed(() => calendarData.value.reduce((sum, d) => sum + d.total_count, 0))
const monthCorrect = computed(() => calendarData.value.reduce((sum, d) => sum + d.correct_count, 0))
const monthAccuracy = computed(() => monthTotal.value ? (monthCorrect.value / monthTotal.value * 100) : 0)
const monthDuration = computed(() => calendarData.value.reduce((sum, d) => sum + d.duration, 0))

const maxDailyCount = computed(() => {
  if (!dailyStats.value.length) return 1
  return Math.max(...dailyStats.value.map(d => d.total_count), 1)
})

const chartDays = computed(() => {
  const days: { date: string; count: number; pct: number; label: string }[] = []
  const dataMap = new Map(dailyStats.value.map(d => [d.date, d]))
  for (let i = 29; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    const key = d.toISOString().split('T')[0]
    const count = dataMap.get(key)?.total_count || 0
    days.push({
      date: key,
      count,
      pct: count / maxDailyCount.value,
      label: `${d.getMonth() + 1}/${d.getDate()}`,
    })
  }
  return days
})

function shiftMonth(delta: number) {
  const [y, m] = currentMonth.value.split('-').map(Number)
  const d = new Date(y, m - 1 + delta, 1)
  currentMonth.value = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
  loadCalendar()
}

function accuracyClass(a: number) {
  if (a >= 80) return 'text-emerald-500'
  if (a >= 60) return 'text-indigo-600'
  if (a >= 40) return 'text-yellow-500'
  return 'text-red-500'
}

function barColor(pct: number) {
  if (pct === 0) return 'bg-gray-100'
  if (pct > 0.6) return 'bg-indigo-600'
  if (pct > 0.3) return 'bg-indigo-400'
  return 'bg-indigo-200'
}

async function loadCalendar() {
  try {
    calendarData.value = await getCalendarStats(currentMonth.value)
  } catch {
    calendarData.value = []
  }
}

onMounted(async () => {
  try {
    const [ov, daily, subj, chapter] = await Promise.all([
      getStatsOverview(),
      getDailyStats(30),
      getSubjectProgress(),
      getChapterProgress(),
    ])
    overview.value = ov
    dailyStats.value = daily || []
    subjectProgress.value = subj || []
    chapterProgress.value = chapter || []
    await loadCalendar()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <BasePageHeader title="学习进度" subtitle="掌握你的学习节奏与知识掌握情况" />

    <BaseSkeleton v-if="loading" variant="detail" />

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

      <!-- 学习日历热力图 -->
      <div class="mb-6 rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
        <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-900">学习日历</h2>
          <div class="flex items-center gap-2">
            <button class="cursor-pointer rounded-lg border border-gray-300 bg-white px-2.5 py-1 text-sm text-gray-600 hover:bg-gray-50" @click="shiftMonth(-1)">
              ← 上月
            </button>
            <span class="min-w-24 text-center text-sm font-medium text-gray-700">{{ currentMonth }}</span>
            <button class="cursor-pointer rounded-lg border border-gray-300 bg-white px-2.5 py-1 text-sm text-gray-600 hover:bg-gray-50" @click="shiftMonth(1)">
              下月 →
            </button>
          </div>
        </div>

        <div class="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
          <div class="rounded-lg bg-gray-50 p-3 text-center">
            <p class="text-lg font-bold text-indigo-600">{{ monthTotal }}</p>
            <p class="text-xs text-gray-400">本月答题</p>
          </div>
          <div class="rounded-lg bg-gray-50 p-3 text-center">
            <p class="text-lg font-bold" :class="accuracyClass(monthAccuracy)">{{ monthAccuracy.toFixed(1) }}%</p>
            <p class="text-xs text-gray-400">本月正确率</p>
          </div>
          <div class="rounded-lg bg-gray-50 p-3 text-center">
            <p class="text-lg font-bold text-emerald-600">{{ monthCorrect }}</p>
            <p class="text-xs text-gray-400">本月答对</p>
          </div>
          <div class="rounded-lg bg-gray-50 p-3 text-center">
            <p class="text-lg font-bold text-gray-700">{{ Math.round(monthDuration / 60) }} 分</p>
            <p class="text-xs text-gray-400">本月学习时长</p>
          </div>
        </div>

        <CalendarHeatmap :month="currentMonth" :data="calendarData" />
      </div>

      <!-- 近30天趋势 + 科目掌握度 -->
      <div class="grid gap-6 lg:grid-cols-2">
        <div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
          <h2 class="mb-4 text-base font-semibold text-gray-900">近 30 天答题趋势</h2>
          <div class="flex h-36 items-end gap-[3px]">
            <div v-for="(day, idx) in chartDays" :key="idx" class="group relative flex-1 flex flex-col justify-end">
              <div
                class="w-full cursor-pointer rounded-t transition-all duration-300"
                :class="barColor(day.pct)"
                :style="{ height: Math.max(day.pct * 100, day.count > 0 ? 8 : 4) + '%' }"
              >
                <div class="pointer-events-none absolute bottom-full left-1/2 z-10 mb-1 hidden -translate-x-1/2 whitespace-nowrap rounded bg-gray-800 px-2 py-1 text-xs text-white group-hover:block">
                  {{ day.date }}: {{ day.count }} 题
                </div>
              </div>
            </div>
          </div>
          <div class="mt-2 flex items-center justify-between text-xs text-gray-400">
            <span>{{ chartDays[0]?.label || '' }}</span>
            <div class="flex items-center gap-1">
              <span>少</span>
              <div class="h-3 w-3 rounded-sm bg-gray-100" />
              <div class="h-3 w-3 rounded-sm bg-indigo-200" />
              <div class="h-3 w-3 rounded-sm bg-indigo-400" />
              <div class="h-3 w-3 rounded-sm bg-indigo-600" />
              <span>多</span>
            </div>
            <span>{{ chartDays[chartDays.length - 1]?.label || '' }}</span>
          </div>
        </div>

        <div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
          <h2 class="mb-4 text-base font-semibold text-gray-900">科目掌握度</h2>
          <div v-if="subjectProgress.length === 0" class="py-10 text-center text-sm text-gray-400">
            暂无练习数据
          </div>
          <div v-else class="space-y-4">
            <div v-for="sp in subjectProgress" :key="sp.subject_id">
              <div class="mb-1 flex items-center justify-between text-sm">
                <span class="truncate font-medium text-gray-700">{{ sp.subject_name }}</span>
                <span class="ml-2 shrink-0">
                  <span class="text-gray-400">{{ sp.correct_count }}/{{ sp.total_count }}</span>
                  <span class="ml-1 font-semibold" :class="accuracyClass(sp.accuracy)">{{ sp.accuracy.toFixed(0) }}%</span>
                </span>
              </div>
              <div class="h-2 overflow-hidden rounded-full bg-gray-100">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="accuracyClass(sp.accuracy) === 'text-emerald-500' ? 'bg-emerald-500' : accuracyClass(sp.accuracy) === 'text-indigo-600' ? 'bg-indigo-500' : 'bg-yellow-500'"
                  :style="{ width: Math.min(100, sp.accuracy) + '%' }"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 章节掌握度 -->
      <div v-if="chapterProgress.length" class="mt-6 rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-base font-semibold text-gray-900">章节掌握度</h2>
          <button class="cursor-pointer text-xs text-indigo-600 hover:underline" @click="router.push('/subjects')">浏览章节</button>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div
            v-for="cp in chapterProgress.slice(0, 8)"
            :key="cp.chapter_id"
            class="cursor-pointer rounded-lg border border-gray-100 p-4 transition-colors hover:border-indigo-200 hover:bg-indigo-50/30"
            @click="router.push(`/practice/chapter/${cp.chapter_id}`)"
          >
            <div class="mb-2 flex items-center justify-between">
              <span class="truncate text-sm font-medium text-gray-700">{{ cp.chapter_name }}</span>
              <span class="ml-2 shrink-0 text-sm font-semibold" :class="accuracyClass(cp.accuracy)">{{ cp.accuracy.toFixed(0) }}%</span>
            </div>
            <div class="h-1.5 overflow-hidden rounded-full bg-gray-100">
              <div
                class="h-full rounded-full"
                :class="cp.accuracy >= 60 ? 'bg-emerald-500' : cp.accuracy >= 30 ? 'bg-yellow-500' : 'bg-red-500'"
                :style="{ width: Math.min(100, cp.accuracy) + '%' }"
              />
            </div>
            <p class="mt-1.5 text-xs text-gray-400">已练 {{ cp.total_count }} 题 · 答对 {{ cp.correct_count }} 题</p>
          </div>
        </div>
      </div>

      <!-- 模考统计 -->
      <div v-if="overview.total_exams > 0" class="mt-6 rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-base font-semibold text-gray-900">模拟考试</h2>
          <button class="cursor-pointer text-xs text-indigo-600 hover:underline" @click="router.push('/exam-records')">查看记录</button>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="rounded-lg bg-gray-50 p-4 text-center">
            <p class="text-2xl font-bold text-indigo-600">{{ overview.total_exams }}</p>
            <p class="text-xs text-gray-500">已完成试卷</p>
          </div>
          <div class="rounded-lg bg-gray-50 p-4 text-center">
            <p class="text-2xl font-bold" :class="accuracyClass(overview.avg_exam_score / 100 * 100)">{{ overview.avg_exam_score.toFixed(0) }}</p>
            <p class="text-xs text-gray-500">平均分</p>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
