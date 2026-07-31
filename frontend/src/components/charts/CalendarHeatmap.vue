<script setup lang="ts">
import { computed } from 'vue'
import type { CalendarStatsResp } from '@/types/stats'

const props = withDefaults(defineProps<{
  month: string
  data: CalendarStatsResp[]
  year?: number
}>(), {
  year: new Date().getFullYear(),
})

interface Cell {
  date: string
  day: number
  count: number
  duration: number
  level: number
}

const dataMap = computed(() => {
  const map = new Map<string, CalendarStatsResp>()
  for (const d of props.data) map.set(d.date, d)
  return map
})

const maxCount = computed(() => {
  if (!props.data.length) return 1
  return Math.max(...props.data.map(d => d.total_count), 1)
})

const cells = computed<Cell[]>(() => {
  const [yearStr, monthStr] = props.month.split('-')
  const y = Number(yearStr)
  const m = Number(monthStr)
  const firstDay = new Date(y, m - 1, 1)
  // 周一为一周开始
  const offset = (firstDay.getDay() + 6) % 7
  const daysInMonth = new Date(y, m, 0).getDate()

  const result: Cell[] = []
  for (let day = 1; day <= daysInMonth; day++) {
    const date = new Date(y, m - 1, day)
    const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    const entry = dataMap.value.get(key)
    const count = entry?.total_count || 0
    const level = count === 0 ? 0 : Math.min(4, Math.ceil((count / maxCount.value) * 4))
    result.push({
      date: key,
      day,
      count,
      duration: entry?.duration || 0,
      level,
    })
  }
  // 前置空白格子
  const blanks = Array.from({ length: offset }, () => ({
    date: '', day: 0, count: 0, duration: 0, level: -1,
  }))
  return [...blanks, ...result]
})

const weekRows = computed(() => {
  const rows: Cell[][] = []
  for (let i = 0; i < cells.value.length; i += 7) {
    rows.push(cells.value.slice(i, i + 7))
  }
  return rows
})

const monthLabel = computed(() => `${props.month.split('-')[1]} 月`)

const colorClasses = [
  'bg-gray-100',
  'bg-emerald-100',
  'bg-emerald-300',
  'bg-emerald-500',
  'bg-emerald-600',
]

function cellTitle(cell: Cell) {
  if (!cell.count) return `${cell.date}: 未练习`
  return `${cell.date}: ${cell.count} 题 · 用时 ${(cell.duration / 60).toFixed(0)} 分钟`
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between text-xs text-gray-500">
      <span>{{ monthLabel }} · 学习日历</span>
      <div class="flex items-center gap-1">
        <span>少</span>
        <span v-for="(c, i) in colorClasses" :key="i" class="h-3 w-3 rounded-sm" :class="c" />
        <span>多</span>
      </div>
    </div>
    <div class="mt-3 overflow-x-auto">
      <div class="inline-block min-w-full">
        <div class="flex gap-1.5 text-[10px] text-gray-400">
          <div class="w-7" />
          <div class="flex flex-1 justify-between px-1">
            <span>一</span><span>二</span><span>三</span><span>四</span><span>五</span><span>六</span><span>日</span>
          </div>
        </div>
        <table class="mt-1 w-full border-separate border-spacing-1">
          <tr v-for="(row, ri) in weekRows" :key="ri">
            <td class="w-7 pr-1 text-right text-[10px] text-gray-400">
              <template v-if="row.find(c => c.day === 1)">第 {{ ri + 1 }} 周</template>
            </td>
            <td v-for="(cell, ci) in row" :key="ci" class="h-8 w-[calc(100%/7)]">
              <div
                v-if="cell.day"
                :title="cellTitle(cell)"
                class="group relative mx-auto flex h-8 w-full max-w-9 cursor-pointer items-center justify-center rounded-md text-[10px] font-medium transition-transform hover:scale-110"
                :class="[colorClasses[cell.level]]"
                :style="cell.day ? {} : undefined"
              >
                <span :class="cell.level >= 3 ? 'text-white' : 'text-gray-700'">{{ cell.day }}</span>
                <div class="pointer-events-none absolute -top-1 left-1/2 z-10 hidden -translate-x-1/2 whitespace-nowrap rounded-md bg-gray-800 px-2 py-1 text-xs text-white group-hover:block">
                  {{ cellTitle(cell) }}
                </div>
              </div>
              <div v-else class="h-8 w-full rounded-md bg-transparent" />
            </td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>
