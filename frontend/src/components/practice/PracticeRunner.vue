<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { submitPractice } from '@/api/practice'
import { addFavorite, removeFavorite, checkFavorited } from '@/api/bookmark'
import { sanitizeHtml } from '@/utils/sanitize'
import { typeLabel, parseOptions, isCorrectAnswer } from '@/utils/question'
import type { Question, PracticeRecordResp } from '@/types/question'

const props = withDefaults(defineProps<{
  questions: Question[]
  mode?: string
  title?: string
  submitHandler?: (questionId: number, answer: string) => Promise<PracticeRecordResp>
}>(), {
  mode: 'chapter',
  title: '',
})

const emit = defineEmits<{ (e: 'back'): void }>()

const currentIndex = ref(0)
const answers = ref<Record<number, string>>({})
const submitted = ref<Record<number, boolean>>({})
const marked = ref<Record<number, boolean>>({})
const favorited = ref<Record<number, boolean>>({})
const showCard = ref(false)
const loadingFavorites = ref(true)

const current = computed(() => props.questions[currentIndex.value])
const total = computed(() => props.questions.length)
const progressText = computed(() => `第 ${currentIndex.value + 1} 题 / 共 ${total.value} 题`)
const answeredCount = computed(() => Object.keys(answers.value).length)
const markedCount = computed(() => Object.values(marked.value).filter(Boolean).length)

onMounted(loadFavoriteState)
watch(() => props.questions, loadFavoriteState)

async function loadFavoriteState() {
  loadingFavorites.value = true
  for (const q of props.questions) {
    try {
      const res = await checkFavorited(q.id)
      favorited.value[q.id] = res.favorited
    } catch {
      favorited.value[q.id] = false
    }
  }
  loadingFavorites.value = false
}

function goTo(index: number) {
  if (index >= 0 && index < total.value) currentIndex.value = index
}

function selectAnswer(questionId: number, value: string) {
  const q = props.questions.find(x => x.id === questionId)
  if (!q || submitted.value[questionId]) return
  if (q.type === 'multi') {
    const current = answers.value[questionId] ? answers.value[questionId].split(',') : []
    const idx = current.indexOf(value)
    if (idx >= 0) current.splice(idx, 1)
    else current.push(value)
    current.sort()
    answers.value[questionId] = current.join(',')
  } else {
    answers.value[questionId] = value
  }
}

function toggleMark() {
  if (current.value) marked.value[current.value.id] = !marked.value[current.value.id]
}

async function toggleFavorite() {
  const q = current.value
  if (!q) return
  try {
    if (favorited.value[q.id]) {
      await removeFavorite(q.id)
      favorited.value[q.id] = false
    } else {
      await addFavorite(q.id)
      favorited.value[q.id] = true
    }
  } catch {
    // toast 由 request 层提示
  }
}

async function handleSubmit() {
  const q = current.value
  if (!q || !answers.value[q.id] || submitted.value[q.id]) return
  try {
    const submit = props.submitHandler
      ? props.submitHandler
      : (questionId: number, answer: string) => submitPractice({
        question_id: questionId,
        mode: props.mode,
        answer,
        duration: 0,
      })
    const res = await submit(q.id, answers.value[q.id])
    submitted.value[q.id] = true
    if (res.answer) answers.value[q.id] = res.answer
  } catch {
    // toast 由 request 层提示
  }
}

function isSelected(questionId: number, value: string) {
  const ans = answers.value[questionId] || ''
  return ans.split(',').includes(value)
}

function answeredText(q: Question): string {
  const ans = answers.value[q.id] || ''
  if (!ans) return '未作答'
  if (q.type === 'judge') {
    if (ans === 'A') return '正确'
    if (ans === 'B') return '错误'
    return ans
  }
  if (q.type === 'single' || q.type === 'multi') {
    return ans.split(',').join('、')
  }
  return ans
}

function isCorrect(q: Question): boolean {
  return isCorrectAnswer(q.type, answers.value[q.id], q.answer)
}

function statusClass(questionId: number) {
  if (submitted.value[questionId]) {
    const q = props.questions.find(x => x.id === questionId)
    return q && isCorrect(q) ? 'bg-emerald-500 text-white' : 'bg-red-500 text-white'
  }
  if (answers.value[questionId]) return 'bg-indigo-600 text-white'
  if (marked.value[questionId]) return 'bg-yellow-400 text-white'
  return 'bg-gray-200 text-gray-600'
}

function optionState(q: Question, optValue: string) {
  const isSub = !!submitted.value[q.id]
  const selected = isSelected(q.id, optValue)
  if (!isSub) {
    if (selected) return 'selected'
    return 'idle'
  }
  if (q.type === 'multi') {
    const correctSet = (q.answer || '').split(',').map(s => s.trim())
    if (correctSet.includes(optValue)) return 'correct'
    if (selected) return 'wrong'
    return 'dim'
  }
  if (optValue === q.answer) return 'correct'
  if (selected) return 'wrong'
  return 'dim'
}

function optionClass(state: string) {
  switch (state) {
    case 'selected': return 'border-indigo-500 bg-indigo-50'
    case 'correct': return 'border-emerald-500 bg-emerald-50'
    case 'wrong': return 'border-red-500 bg-red-50'
    case 'dim': return 'border-gray-200 opacity-60'
    default: return 'border-gray-200 hover:border-indigo-300'
  }
}

function optionBadgeClass(q: Question, optValue: string, state: string) {
  const selected = isSelected(q.id, optValue)
  if (!submitted.value[q.id] && selected) return 'border-indigo-500 bg-indigo-500 text-white'
  if (state === 'correct') return 'border-emerald-500 bg-emerald-500 text-white'
  if (state === 'wrong') return 'border-red-500 bg-red-500 text-white'
  return 'border-gray-300'
}
</script>

<template>
  <div class="mx-auto max-w-4xl">
    <div class="mb-4 flex items-center justify-between">
      <button class="flex cursor-pointer items-center gap-1 text-sm text-gray-500 hover:text-indigo-600" @click="emit('back')">
        ← 返回
      </button>
      <div class="flex items-center gap-3">
        <span class="text-sm text-gray-500">{{ progressText }}</span>
        <button
          class="cursor-pointer rounded-lg border px-2.5 py-1 text-xs"
          :class="showCard ? 'border-indigo-300 bg-indigo-50 text-indigo-600' : 'border-gray-300 bg-white text-gray-600'"
          @click="showCard = !showCard"
        >
          答题卡
        </button>
      </div>
    </div>

    <!-- 答题卡面板 -->
    <div v-if="showCard" class="mb-4 rounded-lg bg-white p-4 shadow-sm">
      <div class="mb-2 flex flex-wrap items-center gap-4 text-xs text-gray-500">
        <span><span class="inline-block h-3 w-3 rounded bg-emerald-500 align-middle" /> 正确</span>
        <span><span class="inline-block h-3 w-3 rounded bg-red-500 align-middle" /> 错误</span>
        <span><span class="inline-block h-3 w-3 rounded bg-indigo-500 align-middle" /> 已答</span>
        <span><span class="inline-block h-3 w-3 rounded bg-yellow-400 align-middle" /> 标记</span>
        <span><span class="inline-block h-3 w-3 rounded bg-gray-200 align-middle" /> 未答</span>
      </div>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="(q, idx) in questions"
          :key="q.id"
          class="flex h-8 w-8 cursor-pointer items-center justify-center rounded text-xs font-medium transition-colors"
          :class="[statusClass(q.id), { 'ring-2 ring-indigo-400 ring-offset-1': idx === currentIndex }]"
          @click="goTo(idx)"
        >
          {{ idx + 1 }}
        </button>
      </div>
      <div class="mt-2 text-xs text-gray-400">
        已答 {{ answeredCount }} / {{ total }} 题
        <span v-if="markedCount"> · 标记 {{ markedCount }} 题</span>
      </div>
    </div>

    <!-- 题目卡片 -->
    <div v-if="current" class="rounded-lg bg-white p-6 shadow-sm">
      <div class="mb-4 flex flex-wrap items-center gap-2">
        <span class="rounded-lg bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-600">{{ typeLabel(current.type) }}</span>
        <span class="rounded-lg bg-gray-100 px-2 py-0.5 text-xs text-gray-600">
          {{ { easy: '简单', medium: '中等', hard: '困难' }[current.difficulty] || current.difficulty }}
        </span>
        <span v-if="current.year" class="text-xs text-gray-400">{{ current.year }}年</span>
        <span v-if="marked[current.id]" class="rounded-lg bg-yellow-50 px-2 py-0.5 text-xs text-yellow-600">已标记</span>
      </div>

      <div class="mb-6 text-base leading-relaxed text-gray-800" v-html="sanitizeHtml(current.content)" />

      <!-- 单选 / 多选 -->
      <div v-if="current.type === 'single' || current.type === 'multi'" class="space-y-2">
        <button
          v-for="opt in parseOptions(current.options)"
          :key="opt.label"
          class="flex w-full cursor-pointer items-center gap-3 rounded-lg border px-4 py-3 text-left text-sm transition-colors"
          :class="optionClass(optionState(current, opt.label))"
          @click="selectAnswer(current.id, opt.label)"
        >
          <span
            class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-xs font-medium"
            :class="optionBadgeClass(current, opt.label, optionState(current, opt.label))"
          >{{ opt.label }}</span>
          <span>{{ opt.text }}</span>
        </button>
      </div>

      <!-- 判断题 -->
      <div v-else-if="current.type === 'judge'" class="flex gap-4">
        <button
          v-for="val in ['正确', '错误']"
          :key="val"
          class="flex-1 cursor-pointer rounded-lg border px-6 py-3 text-center text-sm font-medium transition-colors"
          :class="optionClass(optionState(current, val === '正确' ? 'A' : 'B'))"
          @click="selectAnswer(current.id, val === '正确' ? 'A' : 'B')"
        >
          {{ val }}
        </button>
      </div>

      <!-- 填空 / 简答 / 综合 -->
      <div v-else>
        <textarea
          v-model="answers[current.id]"
          :disabled="!!submitted[current.id]"
          rows="3"
          class="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          :class="{
            'border-emerald-500 bg-emerald-50': submitted[current.id] && isCorrect(current),
            'border-red-500 bg-red-50': submitted[current.id] && answers[current.id] && !isCorrect(current),
          }"
          placeholder="请输入答案"
        />
      </div>

      <!-- 解析（提交后显示） -->
      <div v-if="submitted[current.id]" class="mt-6 rounded-lg bg-gray-50 p-4">
        <div class="mb-2 text-sm font-medium" :class="isCorrect(current) ? 'text-emerald-600' : 'text-red-600'">
          {{ isCorrect(current) ? '✓ 回答正确' : '✗ 回答错误' }}
        </div>
        <div class="mb-2 text-sm text-gray-600">
          你的答案：<span class="font-medium text-gray-800">{{ answeredText(current) }}</span>
        </div>
        <div v-if="!isCorrect(current)" class="mb-2 text-sm text-gray-600">
          正确答案：<span class="font-medium text-gray-800">{{ current.answer }}</span>
        </div>
        <div v-if="current.analysis" class="text-sm text-gray-600">
          <span class="font-medium text-gray-800">解析：</span>{{ current.analysis }}
        </div>
      </div>
    </div>

    <!-- 底部操作栏 -->
    <div class="mt-4 flex flex-wrap items-center justify-between gap-2 rounded-lg bg-white p-3 shadow-sm">
      <div class="flex gap-2">
        <button
          class="cursor-pointer rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50 disabled:opacity-30"
          :disabled="currentIndex === 0"
          @click="goTo(currentIndex - 1)"
        >
          ← 上一题
        </button>
        <button
          class="cursor-pointer rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50 disabled:opacity-30"
          :disabled="currentIndex === total - 1"
          @click="goTo(currentIndex + 1)"
        >
          下一题 →
        </button>
      </div>
      <div class="flex flex-wrap gap-2">
        <button
          class="cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors"
          :class="current && marked[current.id] ? 'border-yellow-300 bg-yellow-50 text-yellow-700' : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
          @click="toggleMark"
        >
          {{ current && marked[current.id] ? '已标记' : '标记疑问' }}
        </button>
        <button
          class="cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors"
          :class="current && favorited[current.id] ? 'border-red-300 bg-red-50 text-red-600' : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
          @click="toggleFavorite"
        >
          {{ current && favorited[current.id] ? '♥ 已收藏' : '♡ 收藏' }}
        </button>
        <button
          v-if="current && !submitted[current.id]"
          class="cursor-pointer rounded-lg bg-indigo-600 px-4 py-1.5 text-sm text-white hover:bg-indigo-700 disabled:opacity-50"
          :disabled="!answers[current.id]"
          @click="handleSubmit"
        >
          提交答案
        </button>
      </div>
    </div>
  </div>
</template>
