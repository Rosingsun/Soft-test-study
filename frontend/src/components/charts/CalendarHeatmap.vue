<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { CalendarStatsResp } from '@/types/stats'
import { formatDuration } from '@/utils/format'

const props = defineProps<{
  month: string
  data: CalendarStatsResp[]
}>()

interface Cell {
  date: string
  day: number
  count: number
  correct: number
  incorrect: number
  duration: number
  accuracy: number
  level: number
  isToday: boolean
}

const selectedDate = ref('')

watch(() => props.month, () => {
  selectedDate.value = ''
})

const dataMap = computed(() => {
  const map = new Map<string, CalendarStatsResp>()
  for (const d of props.data) map.set(d.date.slice(0, 10), d)
  return map
})

const maxCount = computed(() => Math.max(1, ...props.data.map(d => d.total_count)))

const cells = computed<Cell[]>(() => {
  const [yearStr, monthStr] = props.month.split('-')
  const y = Number(yearStr)
  const m = Number(monthStr)
  const offset = (new Date(y, m - 1, 1).getDay() + 6) % 7
  const daysInMonth = new Date(y, m, 0).getDate()
  const now = new Date()

  const result: Cell[] = []
  for (let day = 1; day <= daysInMonth; day++) {
    const date = new Date(y, m - 1, day)
    const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    const entry = dataMap.value.get(key)
    const count = entry?.total_count || 0
    const correct = entry?.correct_count || 0
    const duration = entry?.duration || 0
    result.push({
      date: key,
      day,
      count,
      correct,
      incorrect: Math.max(0, count - correct),
      duration,
      accuracy: count ? (correct / count) * 100 : 0,
      level: count === 0 ? 0 : Math.min(4, Math.max(1, Math.round((count / maxCount.value) * 4))),
      isToday:
        date.getFullYear() === now.getFullYear() &&
        date.getMonth() === now.getMonth() &&
        date.getDate() === now.getDate(),
    })
  }
  const blanks: Cell[] = Array.from({ length: offset }, () => ({
    date: '', day: 0, count: 0, correct: 0, incorrect: 0, duration: 0, accuracy: 0, level: -1, isToday: false,
  }))
  return [...blanks, ...result]
})

const weekRows = computed(() => {
  const rows: Cell[][] = []
  for (let i = 0; i < cells.value.length; i += 7) rows.push(cells.value.slice(i, i + 7))
  return rows
})

const monthSummary = computed(() => {
  const list = cells.value.filter(c => c.day > 0)
  const total = list.reduce((s, c) => s + c.count, 0)
  const correct = list.reduce((s, c) => s + c.correct, 0)
  const duration = list.reduce((s, c) => s + c.duration, 0)
  return {
    activeDays: list.filter(c => c.count > 0).length,
    total,
    correct,
    incorrect: Math.max(0, total - correct),
    accuracy: total ? (correct / total) * 100 : 0,
    duration,
  }
})

const selected = computed<Cell | null>(() => cells.value.find(c => c.date === selectedDate.value) || null)

const colorClasses = [
  'bg-gray-100',
  'bg-indigo-100',
  'bg-indigo-300',
  'bg-indigo-500',
  'bg-indigo-600',
]

function cellClass(cell: Cell) {
  if (!cell.day) return 'pointer-events-none border-transparent'
  const base = `${colorClasses[cell.level]} border-gray-100`
  if (selectedDate.value === cell.date) return `${base} ring-2 ring-indigo-500 ring-offset-1`
  if (cell.isToday) return `${base} ring-1 ring-inset ring-indigo-400`
  return base
}

function selectCell(cell: Cell) {
  if (!cell.day) return
  selectedDate.value = selectedDate.value === cell.date ? '' : cell.date
}

function accuracyClass(a: number) {
  if (a >= 80) return 'text-emerald-600'
  if (a >= 60) return 'text-indigo-600'
  if (a >= 40) return 'text-amber-500'
  return 'text-red-500'
}

function cellTextClass(cell: Cell, onLight: string) {
  if (cell.level >= 3) return 'text-white'
  return onLight
}

function accuracyBar(a: number) {
  if (a >= 80) return 'bg-emerald-500'
  if (a >= 60) return 'bg-indigo-500'
  if (a >= 40) return 'bg-amber-500'
  return 'bg-red-500'
}

function cellTitle(cell: Cell) {
  if (!cell.count) return `${cell.date}：未练习`
  return `${cell.date}：${cell.count} 题 · 答对 ${cell.correct} 题 · 正确率 ${cell.accuracy.toFixed(1)}% · 用时 ${formatDuration(cell.duration)}`
}
</script>

<template>
  <div>
    <!-- 月份汇总 -->
    <div class="mb-5 grid grid-cols-2 gap-3 sm:grid-cols-5">
      <div class="rounded-xl border border-gray-100 bg-gradient-to-br from-white to-gray-50 p-3 shadow-sm">
        <p class="text-xl font-bold text-indigo-600">{{ monthSummary.total }}</p>
        <p class="mt-0.5 text-xs text-gray-500">本月答题</p>
      </div>
      <div class="rounded-xl border border-gray-100 bg-gradient-to-br from-white to-gray-50 p-3 shadow-sm">
        <p class="text-xl font-bold" :class="accuracyClass(monthSummary.accuracy)">{{ monthSummary.accuracy.toFixed(1) }}%</p>
        <p class="mt-0.5 text-xs text-gray-500">本月正确率</p>
      </div>
      <div class="rounded-xl border border-gray-100 bg-gradient-to-br from-white to-gray-50 p-3 shadow-sm">
        <p class="text-xl font-bold text-emerald-600">{{ monthSummary.correct }}</p>
        <p class="mt-0.5 text-xs text-gray-500">本月答对</p>
      </div>
      <div class="rounded-xl border border-gray-100 bg-gradient-to-br from-white to-gray-50 p-3 shadow-sm">
        <p class="text-xl font-bold text-red-500">{{ monthSummary.incorrect }}</p>
        <p class="mt-0.5 text-xs text-gray-500">本月答错</p>
      </div>
      <div class="rounded-xl border border-gray-100 bg-gradient-to-br from-white to-gray-50 p-3 shadow-sm">
        <p class="text-xl font-bold text-gray-700">{{ formatDuration(monthSummary.duration) }}</p>
        <p class="mt-0.5 text-xs text-gray-500">学习时长</p>
      </div>
    </div>

    <!-- 图例与活跃天数 -->
    <div class="mb-3 flex items-center justify-between">
      <span class="inline-flex items-center gap-1.5 text-xs text-gray-400">
        本月活跃 <span class="font-semibold text-indigo-600">{{ monthSummary.activeDays }}</span> 天
      </span>
      <div class="flex items-center gap-1 text-xs text-gray-400">
        <span>少</span>
        <span v-for="(c, i) in colorClasses" :key="i" class="h-3 w-3 rounded" :class="c" />
        <span>多</span>
      </div>
    </div>

    <!-- 星期表头 -->
    <div class="grid grid-cols-7 gap-1.5 text-center text-[11px] font-medium text-gray-400">
      <span v-for="w in ['一', '二', '三', '四', '五', '六', '日']" :key="w">{{ w }}</span>
    </div>

    <!-- 日历格子 -->
    <div class="mt-1.5 space-y-1.5">
      <div v-for="(row, ri) in weekRows" :key="ri" class="grid grid-cols-7 gap-1.5">
        <div
          v-for="(cell, ci) in row"
          :key="ci"
          class="relative aspect-[7/5] cursor-pointer overflow-hidden rounded-lg border transition-all duration-150 hover:-translate-y-0.5 hover:shadow-sm"
          :class="cellClass(cell)"
          :title="cell.day ? cellTitle(cell) : ''"
          @click="selectCell(cell)"
        >
          <template v-if="cell.day">
            <span
              class="absolute left-1.5 top-1 text-[10px] font-semibold leading-none"
              :class="cell.level >= 3 ? 'text-white/90' : cell.count > 0 ? 'text-gray-500' : 'text-gray-400'"
            >{{ cell.day }}</span>

            <!-- 正确率（居中主数字） -->
            <div v-if="cell.count > 0" class="absolute inset-0 flex items-center justify-center">
              <span
                class="whitespace-nowrap text-base font-bold leading-none tracking-tight sm:text-lg lg:text-xl"
                :class="cellTextClass(cell, accuracyClass(cell.accuracy))"
              >{{ cell.accuracy.toFixed(0) }}<span class="text-[0.62em] font-semibold">%</span></span>
            </div>

            <!-- 答对 / 答错（底部分隔两角） -->
            <div v-if="cell.count > 0" class="absolute inset-x-1 bottom-1 hidden items-center justify-between sm:flex">
              <span
                class="flex items-center gap-1 rounded-md px-1.5 py-1 text-[11px] font-semibold leading-none lg:text-xs"
                :class="cell.level >= 3 ? 'bg-white/15 text-white' : 'bg-emerald-50 text-emerald-700'"
              >
                <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                {{ cell.correct }}
              </span>
              <span
                class="flex items-center gap-1 rounded-md px-1.5 py-1 text-[11px] font-semibold leading-none lg:text-xs"
                :class="cell.level >= 3 ? 'bg-white/15 text-white' : 'bg-red-50 text-red-600'"
              >
                <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
                {{ cell.incorrect }}
              </span>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- 选中日期详情 -->
    <div v-if="selected" class="mt-5 rounded-xl border border-indigo-100 bg-gradient-to-br from-indigo-50/80 via-white to-white p-4 shadow-sm">
      <div class="flex items-center justify-between">
        <p class="text-sm font-semibold text-gray-800">{{ selected.date }} · 学习详情</p>
        <button class="cursor-pointer text-xs text-gray-400 transition-colors hover:text-indigo-600" @click="selectedDate = ''">
          关闭 ✕
        </button>
      </div>
      <div class="mt-3 grid grid-cols-2 gap-2.5 sm:grid-cols-5">
        <div class="rounded-lg bg-white p-2.5 text-center shadow-sm">
          <p class="text-base font-bold text-indigo-600">{{ selected.count }}</p>
          <p class="text-[11px] text-gray-500">答题</p>
        </div>
        <div class="rounded-lg bg-white p-2.5 text-center shadow-sm">
          <p class="text-base font-bold text-emerald-600">{{ selected.correct }}</p>
          <p class="text-[11px] text-gray-500">答对</p>
        </div>
        <div class="rounded-lg bg-white p-2.5 text-center shadow-sm">
          <p class="text-base font-bold text-red-500">{{ selected.incorrect }}</p>
          <p class="text-[11px] text-gray-500">答错</p>
        </div>
        <div class="rounded-lg bg-white p-2.5 text-center shadow-sm">
          <p class="text-base font-bold" :class="accuracyClass(selected.accuracy)">{{ selected.accuracy.toFixed(1) }}%</p>
          <p class="text-[11px] text-gray-500">正确率</p>
        </div>
        <div class="rounded-lg bg-white p-2.5 text-center shadow-sm">
          <p class="text-base font-bold text-gray-700">{{ formatDuration(selected.duration) }}</p>
          <p class="text-[11px] text-gray-500">学习时长</p>
        </div>
      </div>
      <div v-if="selected.count" class="mt-3">
        <div class="mb-1 flex items-center justify-between text-xs">
          <span class="text-gray-500">当日正确率</span>
          <span class="font-medium" :class="accuracyClass(selected.accuracy)">{{ selected.accuracy.toFixed(1) }}%</span>
        </div>
        <div class="h-1.5 overflow-hidden rounded-full bg-gray-200/70">
          <div class="h-full rounded-full transition-all duration-500" :class="accuracyBar(selected.accuracy)" :style="{ width: selected.accuracy + '%' }" />
        </div>
      </div>
    </div>
    <div v-else class="mt-5 rounded-xl border border-dashed border-gray-200 bg-gray-50/60 p-4 text-center text-xs text-gray-400">
      点击日历中的日期，查看当天答题、正确率与用时详情
    </div>
  </div>
</template>
