<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'

const props = withDefaults(defineProps<{
  label?: string
  placeholder?: string
  error?: string
  disabled?: boolean
  min?: string
  max?: string
  clearable?: boolean
}>(), {
  placeholder: '请选择日期',
  disabled: false,
  clearable: true,
})

const model = defineModel<string>({ default: '' })

const WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']
const PANEL_WIDTH = 288
const PANEL_HEIGHT = 332

const wrapperRef = ref<HTMLElement | null>(null)
const panelRef = ref<HTMLElement | null>(null)
const open = ref(false)
const today = new Date()

const viewYear = ref(today.getFullYear())
const viewMonth = ref(today.getMonth())
const panelStyle = ref<{ top: string; left: string }>({ top: '0px', left: '0px' })

function parseISO(s: string): Date | null {
  if (!s) return null
  const m = /^(\d{4})-(\d{1,2})-(\d{1,2})$/.exec(s)
  if (!m) return null
  const d = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  return Number.isNaN(d.getTime()) ? null : d
}

function formatYMD(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function isSameDay(a: Date, b: Date | null) {
  if (!b) return false
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

function isDisabled(d: Date) {
  const s = formatYMD(d)
  if (props.min && s < props.min) return true
  if (props.max && s > props.max) return true
  return false
}

watch(model, (v) => {
  const d = parseISO(v)
  if (d) {
    viewYear.value = d.getFullYear()
    viewMonth.value = d.getMonth()
  }
}, { immediate: true })

const displayText = computed(() => model.value || '')

const calendarDays = computed(() => {
  const firstDay = new Date(viewYear.value, viewMonth.value, 1)
  const startWeekday = firstDay.getDay()
  const daysInMonth = new Date(viewYear.value, viewMonth.value + 1, 0).getDate()
  const cells: Array<{ date: Date; current: boolean } | null> = []
  for (let i = 0; i < startWeekday; i++) cells.push(null)
  for (let d = 1; d <= daysInMonth; d++) {
    cells.push({ date: new Date(viewYear.value, viewMonth.value, d), current: true })
  }
  while (cells.length % 7 !== 0) cells.push(null)
  return cells
})

const monthLabel = computed(() => `${viewYear.value} 年 ${viewMonth.value + 1} 月`)
const selectedDate = computed(() => parseISO(model.value))

function prevMonth() {
  if (viewMonth.value === 0) {
    viewYear.value -= 1
    viewMonth.value = 11
  } else {
    viewMonth.value -= 1
  }
}

function nextMonth() {
  if (viewMonth.value === 11) {
    viewYear.value += 1
    viewMonth.value = 0
  } else {
    viewMonth.value += 1
  }
}

function selectDate(d: Date) {
  if (isDisabled(d)) return
  model.value = formatYMD(d)
  open.value = false
}

function goToday() {
  if (isDisabled(today)) return
  selectDate(today)
}

function clear() {
  model.value = ''
}

function toggle() {
  if (props.disabled) return
  open.value = !open.value
}

function updatePosition() {
  const rect = wrapperRef.value?.getBoundingClientRect()
  if (!rect) return
  const margin = 8
  let top = rect.bottom + 6
  let left = rect.left
  if (top + PANEL_HEIGHT > window.innerHeight - margin && rect.top - PANEL_HEIGHT - 6 > margin) {
    top = rect.top - PANEL_HEIGHT - 6
  }
  if (left + PANEL_WIDTH > window.innerWidth - margin) {
    left = window.innerWidth - PANEL_WIDTH - margin
  }
  if (left < margin) left = margin
  panelStyle.value = { top: `${top}px`, left: `${left}px` }
}

watch(open, async (v) => {
  if (v) {
    await nextTick()
    updatePosition()
  }
})

function onDocClick(e: MouseEvent) {
  if (!open.value) return
  const target = e.target as Node | null
  if (!target) return
  if (wrapperRef.value?.contains(target)) return
  if (panelRef.value?.contains(target)) return
  open.value = false
}

function onKeydown(e: KeyboardEvent) {
  if (!open.value) return
  if (e.key === 'Escape') {
    open.value = false
  }
}

function onWindowScroll() {
  if (open.value) open.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocClick, true)
  document.addEventListener('keydown', onKeydown)
  window.addEventListener('resize', onWindowScroll)
  window.addEventListener('scroll', onWindowScroll, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocClick, true)
  document.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', onWindowScroll)
  window.removeEventListener('scroll', onWindowScroll, true)
})
</script>

<template>
  <div ref="wrapperRef" class="relative">
    <label v-if="label" class="mb-1.5 block text-sm font-medium text-gray-700">{{ label }}</label>
    <div
      :class="[
        'flex h-9 w-full items-center rounded-lg border bg-white px-3 text-sm text-gray-800 shadow-sm transition-colors',
        disabled
          ? 'cursor-not-allowed border-gray-200 bg-gray-100 text-gray-400'
          : open
            ? 'cursor-pointer border-indigo-500 ring-2 ring-indigo-500/20'
            : 'cursor-pointer border-gray-300 hover:border-gray-400',
      ]"
    >
      <svg class="mr-2 h-4 w-4 shrink-0 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
      </svg>
      <input
        :value="displayText"
        :placeholder="placeholder"
        :disabled="disabled"
        readonly
        class="min-w-0 flex-1 cursor-pointer bg-transparent text-sm text-gray-800 outline-none placeholder:text-gray-400 disabled:cursor-not-allowed"
        @click="toggle"
      />
      <button
        v-if="clearable && model && !disabled"
        type="button"
        class="ml-1 cursor-pointer rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
        aria-label="清除"
        @click.stop="clear"
      >
        <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
    <p v-if="error" class="mt-1 text-xs text-red-500">{{ error }}</p>

    <Teleport to="body">
      <Transition name="picker-fade">
        <div
          v-if="open"
          ref="panelRef"
          class="fixed z-[110] w-72 rounded-xl border border-gray-100 bg-white p-3 shadow-2xl"
          :style="panelStyle"
        >
          <div class="mb-2 flex items-center justify-between">
            <button type="button" class="cursor-pointer rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700" aria-label="上个月" @click="prevMonth">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" /></svg>
            </button>
            <span class="text-sm font-semibold text-gray-800">{{ monthLabel }}</span>
            <button type="button" class="cursor-pointer rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700" aria-label="下个月" @click="nextMonth">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" /></svg>
            </button>
          </div>

          <div class="mb-1 grid grid-cols-7 gap-y-1 text-center text-xs text-gray-400">
            <span v-for="w in WEEKDAYS" :key="w">{{ w }}</span>
          </div>

          <div class="grid grid-cols-7 gap-y-1 text-center text-sm">
            <template v-for="(cell, i) in calendarDays" :key="i">
              <span v-if="!cell" class="h-8" />
              <button
                v-else
                type="button"
                :disabled="isDisabled(cell.date)"
                :class="[
                  'mx-auto flex h-8 w-8 items-center justify-center rounded-full text-sm transition-colors',
                  isDisabled(cell.date) && 'cursor-not-allowed text-gray-300',
                  !isDisabled(cell.date) && !isSameDay(cell.date, selectedDate) && !isSameDay(cell.date, today) && 'cursor-pointer text-gray-700 hover:bg-indigo-50 hover:text-indigo-600',
                  !isDisabled(cell.date) && isSameDay(cell.date, today) && !isSameDay(cell.date, selectedDate) && 'cursor-pointer text-indigo-600 ring-1 ring-inset ring-indigo-300 hover:bg-indigo-50',
                  selectedDate && isSameDay(cell.date, selectedDate) && 'cursor-pointer bg-indigo-600 font-semibold text-white shadow-sm hover:bg-indigo-600',
                ]"
                @click="selectDate(cell.date)"
              >{{ cell.date.getDate() }}</button>
            </template>
          </div>

          <div class="mt-2 flex items-center justify-between border-t border-gray-100 pt-2">
            <button
              type="button"
              class="cursor-pointer rounded-lg px-2.5 py-1 text-xs font-medium text-indigo-600 transition-colors hover:bg-indigo-50 disabled:cursor-not-allowed disabled:text-gray-300 disabled:hover:bg-transparent"
              :disabled="isDisabled(today)"
              @click="goToday"
            >今天</button>
            <span class="text-xs text-gray-400">{{ model || '未选择' }}</span>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.picker-fade-enter-active,
.picker-fade-leave-active {
  transition: opacity 0.14s ease, transform 0.14s ease;
}
.picker-fade-enter-from,
.picker-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
}
</style>
