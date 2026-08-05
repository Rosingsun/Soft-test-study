<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { submitPractice } from '@/api/practice'
import { addFavorite, removeFavorite, checkFavorited } from '@/api/bookmark'
import { addMark, removeMark, checkMarked } from '@/api/mark'
import { essayScore, checkEssayScore } from '@/api/ai'
import { useAiStore } from '@/stores/ai'
import { sanitizeHtml } from '@/utils/sanitize'
import { typeLabel, parseOptions, isCorrectAnswer, aiScoreSpec, aiDimensionScore } from '@/utils/question'
import type { Question, PracticeRecordResp } from '@/types/question'
import type { EssayScoreResp } from '@/types/ai'

const router = useRouter()
const aiStore = useAiStore()

const props = withDefaults(defineProps<{
  questions: Question[]
  mode?: string
  title?: string
  startIndex?: number
  submitHandler?: (questionId: number, answer: string) => Promise<PracticeRecordResp>
}>(), {
  mode: 'chapter',
  title: '',
  startIndex: 0,
})

const emit = defineEmits<{ (e: 'back'): void }>()

const currentIndex = ref(props.startIndex ?? 0)
const answers = ref<Record<number, string>>({})
const submitted = ref<Record<number, boolean>>({})
const marked = ref<Record<number, boolean>>({})
const favorited = ref<Record<number, boolean>>({})
const showCard = ref(false)
const loadingFavorites = ref(true)
const caseExpanded = ref(true)

const current = computed(() => props.questions[currentIndex.value])
const total = computed(() => props.questions.length)
const progressText = computed(() => `第 ${currentIndex.value + 1} 题 / 共 ${total.value} 题`)
const progressPct = computed(() => total.value ? ((currentIndex.value + 1) / total.value) * 100 : 0)
const answeredCount = computed(() => Object.keys(answers.value).length)
const markedCount = computed(() => Object.values(marked.value).filter(Boolean).length)
// 当前主观题（论文/案例分析）的 AI 评分展示规格
const currentAiSpec = computed(() => current.value ? aiScoreSpec(current.value.type) : aiScoreSpec('essay'))

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
    try {
      const res = await checkMarked(q.id)
      marked.value[q.id] = res.marked
    } catch {
      marked.value[q.id] = false
    }
  }
  loadingFavorites.value = false
}

function goTo(index: number) {
  if (index >= 0 && index < total.value) currentIndex.value = index
}

function handleNext() {
  if (currentIndex.value === total.value - 1) {
    emit('back')
    return
  }
  goTo(currentIndex.value + 1)
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
  // 单选 / 判断题：选中后自动提交（立即判分并显示解析）
  if (q.type === 'single' || q.type === 'judge') {
    handleSubmit()
  }
}

async function toggleMark() {
  const q = current.value
  if (!q) return
  try {
    if (marked.value[q.id]) {
      await removeMark(q.id)
      marked.value[q.id] = false
    } else {
      await addMark(q.id)
      marked.value[q.id] = true
    }
  } catch {
    // toast 由 request 层提示
  }
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

// AI 评分相关状态
const aiScoring = ref<Record<number, boolean>>({})
const aiScoreResults = ref<Record<number, EssayScoreResp>>({})
const aiScoreError = ref<Record<number, string>>({})
const practiceRecordIds = ref<Record<number, number>>({})

// 检查是否已有评分
async function checkExistingScore(_questionId: number, recordId: number) {
  try {
    const res = await checkEssayScore({ record_type: 'practice', record_id: recordId })
    return res.has_score ? res.score : null
  } catch {
    return null
  }
}

// AI 评分（论文 / 案例分析）
async function handleAiScore() {
  const q = current.value
  if (!q || !isSubjective(q) || !answers.value[q.id] || !submitted.value[q.id]) return

  // 检查是否已配置 AI
  if (!aiStore.hasConfig) {
    if (confirm('尚未配置 AI API Key，是否前往配置？')) {
      router.push('/ai/config')
    }
    return
  }

  const recordId = practiceRecordIds.value[q.id]
  if (!recordId) return

  // 检查是否已有评分
  const existing = await checkExistingScore(q.id, recordId)
  if (existing) {
    if (!confirm(`AI 已对此${typeLabel(q.type)}评过分，是否还需要再次测评？`)) {
      // 直接显示已有评分
      aiScoreResults.value[q.id] = existing
      return
    }
  }

  aiScoring.value[q.id] = true
  aiScoreError.value[q.id] = ''

  try {
    const res = await essayScore({
      api_config: aiStore.getConfig(),
      question_id: q.id,
      user_answer: answers.value[q.id],
      record_type: 'practice',
      record_id: recordId,
    })
    aiScoreResults.value[q.id] = res
  } catch (e) {
    aiScoreError.value[q.id] = (e as Error).message
  } finally {
    aiScoring.value[q.id] = false
  }
}

async function handleSubmit() {
  const q = current.value
  if (!q || !answers.value[q.id] || submitted.value[q.id]) return
  // 已存在练习记录，无需重复提交
  if (practiceRecordIds.value[q.id]) {
    submitted.value[q.id] = true
    return
  }
  // 本地即时判分：题目查询已带回正确答案与解析，先同步置为已提交以立即渲染解析与对错，
  // 再异步上报练习记录（统计/错题入库/AI 评分 record_id），不阻塞判分展示。
  submitted.value[q.id] = true
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
    if (res.answer) answers.value[q.id] = res.answer
    // 保存练习记录ID，供后续 AI 评分使用
    practiceRecordIds.value[q.id] = res.id
  } catch {
    // 上报失败仅影响统计/错题入库，不打断已完成的本地判分展示
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
    if (q && (q.type === 'essay' || q.type === 'case_study')) return 'bg-indigo-100 text-indigo-700'
    return q && isCorrect(q) ? 'bg-emerald-500 text-white' : 'bg-red-500 text-white'
  }
  if (answers.value[questionId]) return 'bg-indigo-600 text-white'
  if (marked.value[questionId]) return 'bg-amber-400 text-white'
  return 'bg-gray-100 text-gray-500'
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
    case 'selected': return 'border-indigo-500 bg-indigo-50/70 shadow-sm'
    case 'correct': return 'border-emerald-500 bg-emerald-50/70'
    case 'wrong': return 'border-red-400 bg-red-50/70'
    case 'dim': return 'border-gray-200 opacity-50'
    default: return 'border-gray-200 hover:border-indigo-300 hover:bg-indigo-50/40 hover:shadow-sm'
  }
}

function optionBadgeClass(q: Question, optValue: string, state: string) {
  const selected = isSelected(q.id, optValue)
  if (!submitted.value[q.id] && selected) return 'border-indigo-500 bg-indigo-500 text-white'
  if (state === 'correct') return 'border-emerald-500 bg-emerald-500 text-white'
  if (state === 'wrong') return 'border-red-500 bg-red-500 text-white'
  return 'border-gray-300 text-gray-500'
}

// 主观题（论文 / 案例分析）：不自动判分，仅展示参考答案
function isSubjective(q: Question): boolean {
  return q.type === 'essay' || q.type === 'case_study'
}

const difficultyMap: Record<string, { label: string; cls: string }> = {  easy: { label: '简单', cls: 'bg-emerald-50 text-emerald-600 ring-emerald-600/20' },
  medium: { label: '中等', cls: 'bg-amber-50 text-amber-600 ring-amber-600/20' },
  hard: { label: '困难', cls: 'bg-red-50 text-red-600 ring-red-600/20' },
}
</script>

<template>
  <div class="mx-auto max-w-4xl">
    <!-- 顶部工具栏 -->
    <div class="mb-4 flex items-center justify-between gap-3">
      <button
        class="flex cursor-pointer items-center gap-1.5 rounded-lg px-2 py-1 text-sm text-gray-500 transition-colors hover:bg-white hover:text-indigo-600"
        @click="emit('back')"
      >
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
        返回
      </button>
      <div class="flex items-center gap-3">
        <div class="hidden items-center gap-2 sm:flex">
          <div class="h-1.5 w-28 overflow-hidden rounded-full bg-gray-200">
            <div
              class="h-full rounded-full bg-gradient-to-r from-indigo-500 to-violet-500 transition-all duration-300"
              :style="{ width: progressPct + '%' }"
            />
          </div>
          <span class="text-xs text-gray-400">{{ progressText }}</span>
        </div>
        <span class="text-xs text-gray-400 sm:hidden">{{ progressText }}</span>
        <button
          class="cursor-pointer rounded-lg border px-2.5 py-1 text-xs font-medium transition-colors"
          :class="showCard ? 'border-indigo-300 bg-indigo-50 text-indigo-600' : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300'"
          @click="showCard = !showCard"
        >
          答题卡
        </button>
      </div>
    </div>

    <!-- 答题卡面板 -->
    <div v-if="showCard" class="mb-4 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <div class="mb-3 flex flex-wrap items-center gap-x-4 gap-y-1.5 text-xs text-gray-500">
        <span class="flex items-center gap-1.5"><span class="inline-block h-2.5 w-2.5 rounded-full bg-emerald-500" />正确</span>
        <span class="flex items-center gap-1.5"><span class="inline-block h-2.5 w-2.5 rounded-full bg-red-500" />错误</span>
        <span class="flex items-center gap-1.5"><span class="inline-block h-2.5 w-2.5 rounded-full bg-indigo-500" />已答</span>
        <span class="flex items-center gap-1.5"><span class="inline-block h-2.5 w-2.5 rounded-full bg-amber-400" />标记</span>
        <span class="flex items-center gap-1.5"><span class="inline-block h-2.5 w-2.5 rounded-full bg-gray-200" />未答</span>
      </div>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="(q, idx) in questions"
          :key="q.id"
          class="flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg text-xs font-medium transition-all duration-150 hover:scale-105"
          :class="[statusClass(q.id), { 'ring-2 ring-indigo-400 ring-offset-1': idx === currentIndex }]"
          @click="goTo(idx)"
        >
          {{ idx + 1 }}
        </button>
      </div>
      <div class="mt-3 border-t border-gray-50 pt-2 text-xs text-gray-400">
        已答 {{ answeredCount }} / {{ total }} 题
        <span v-if="markedCount"> · 标记 {{ markedCount }} 题</span>
      </div>
    </div>

    <!-- 题目卡片上方操作栏（题目卡片外正上方） -->
    <div class="sticky top-2 z-10 mb-3 flex flex-wrap items-center justify-between gap-2 rounded-xl border border-gray-100 bg-white/95 p-3 shadow-sm backdrop-blur">
      <div class="flex gap-2">
        <button
          class="flex cursor-pointer items-center gap-1 rounded-lg border border-gray-200 bg-white px-3.5 py-2 text-sm font-medium text-gray-600 shadow-sm transition-all duration-200 hover:border-gray-300 hover:bg-gray-50 disabled:opacity-30 disabled:hover:bg-white"
          :disabled="currentIndex === 0"
          @click="goTo(currentIndex - 1)"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
          上一题
        </button>
        <button
          class="flex cursor-pointer items-center gap-1 rounded-lg border border-gray-200 bg-white px-3.5 py-2 text-sm font-medium text-gray-600 shadow-sm transition-all duration-200 hover:border-gray-300 hover:bg-gray-50 disabled:opacity-30 disabled:hover:bg-white"
          :disabled="currentIndex === total - 1 && !(current && submitted[current.id])"
          @click="handleNext"
        >
          {{ currentIndex === total - 1 ? '完成' : '下一题' }}
          <svg v-if="currentIndex < total - 1" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <span
          v-if="current && submitted[current.id]"
          class="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-600 ring-1 ring-inset ring-emerald-600/20"
        >
          <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
          已提交
        </span>
        <button
          class="flex cursor-pointer items-center gap-1.5 rounded-lg border px-3.5 py-2 text-sm font-medium transition-all duration-200"
          :class="current && marked[current.id]
            ? 'border-amber-300 bg-amber-50 text-amber-700'
            : 'border-gray-200 bg-white text-gray-600 shadow-sm hover:border-gray-300 hover:bg-gray-50'"
          @click="toggleMark"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 3v1.5M3 21v-6m0 0l2.77-.693a9 9 0 016.208.682l.108.054a9 9 0 016.086.71l3.114-.732a48.524 48.524 0 01-.005-10.499l-3.11.732a9 9 0 01-6.085-.711l-.108-.054a9 9 0 00-6.208-.682L3 4.5M3 15V4.5" />
          </svg>
          {{ current && marked[current.id] ? '已标记' : '标记疑问' }}
        </button>
        <button
          class="flex cursor-pointer items-center gap-1.5 rounded-lg border px-3.5 py-2 text-sm font-medium transition-all duration-200"
          :class="current && favorited[current.id]
            ? 'border-red-200 bg-red-50 text-red-600'
            : 'border-gray-200 bg-white text-gray-600 shadow-sm hover:border-gray-300 hover:bg-gray-50'"
          @click="toggleFavorite"
        >
          <svg class="h-4 w-4" :fill="current && favorited[current.id] ? 'currentColor' : 'none'" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z" />
          </svg>
          {{ current && favorited[current.id] ? '已收藏' : '收藏' }}
        </button>
        <button
          v-if="current && !submitted[current.id] && current.type !== 'single' && current.type !== 'judge'"
          class="bg-brand-gradient cursor-pointer rounded-lg px-5 py-2 text-sm font-semibold text-white shadow-md shadow-indigo-600/25 transition-all duration-200 hover:shadow-lg hover:brightness-110 active:scale-[0.98] disabled:opacity-40 disabled:shadow-none"
          :disabled="!answers[current.id]"
          @click="handleSubmit"
        >
          提交答案
        </button>
      </div>
    </div>

    <!-- 题目卡片 -->
    <div v-if="current" class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm sm:p-7">
      <!-- 元信息 -->
      <div class="mb-5 flex flex-wrap items-center gap-2">
        <span class="inline-flex items-center rounded-full bg-indigo-50 px-2.5 py-0.5 text-xs font-medium text-indigo-600 ring-1 ring-inset ring-indigo-600/20">
          {{ typeLabel(current.type) }}
        </span>
        <span
          class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset"
          :class="difficultyMap[current.difficulty]?.cls || 'bg-gray-100 text-gray-600 ring-gray-500/10'"
        >
          {{ difficultyMap[current.difficulty]?.label || current.difficulty }}
        </span>
        <span v-if="current.year" class="text-xs text-gray-400">{{ current.year }} 年真题</span>
        <span v-if="marked[current.id]" class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-600 ring-1 ring-inset ring-amber-600/20">
          <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 3v1.5M3 21v-6m0 0l2.77-.693a9 9 0 016.208.682l.108.054a9 9 0 016.086.71l3.114-.732a48.524 48.524 0 01-.005-10.499l-3.11.732a9 9 0 01-6.085-.711l-.108-.054a9 9 0 00-6.208-.682L3 4.5M3 15V4.5" />
          </svg>
          已标记
        </span>
      </div>

      <!-- 案例材料（案例分析题）：可折叠引用块 -->
      <div v-if="current.case_material" class="mb-5 rounded-xl border-l-4 border-violet-400 bg-violet-50/50 p-4">
        <button
          type="button"
          class="flex w-full cursor-pointer items-center justify-between text-left"
          @click="caseExpanded = !caseExpanded"
        >
          <span class="inline-flex items-center gap-1.5 text-sm font-semibold text-violet-700">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
            </svg>
            案例材料
          </span>
          <svg
            class="h-4 w-4 text-violet-500 transition-transform duration-200"
            :class="caseExpanded ? 'rotate-180' : ''"
            fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
          </svg>
        </button>
        <div v-show="caseExpanded" class="mt-3 max-h-80 overflow-y-auto text-sm leading-7 text-gray-700" v-html="sanitizeHtml(current.case_material)" />
      </div>

      <!-- 题干 -->
      <div class="mb-6 text-[15px] leading-7 text-gray-800" v-html="sanitizeHtml(current.content)" />

      <!-- 单选 / 多选 -->
      <div v-if="current.type === 'single' || current.type === 'multi'" class="space-y-2.5">
        <p v-if="current.type === 'multi' && !submitted[current.id]" class="mb-1 text-xs text-gray-400">本题为多选题，可选择多个选项</p>
        <button
          v-for="opt in parseOptions(current.options)"
          :key="opt.label"
          class="flex w-full cursor-pointer items-center gap-3.5 rounded-xl border-2 px-4 py-3.5 text-left text-sm transition-all duration-200"
          :class="optionClass(optionState(current, opt.label))"
          @click="selectAnswer(current.id, opt.label)"
        >
          <span
            class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border-2 text-xs font-semibold transition-colors"
            :class="optionBadgeClass(current, opt.label, optionState(current, opt.label))"
          >{{ opt.label }}</span>
          <span class="leading-6">{{ opt.text }}</span>
        </button>
      </div>

      <!-- 判断题 -->
      <div v-else-if="current.type === 'judge'" class="grid grid-cols-2 gap-3">
        <button
          v-for="val in ['正确', '错误']"
          :key="val"
          class="flex cursor-pointer items-center justify-center gap-2 rounded-xl border-2 px-6 py-4 text-sm font-semibold transition-all duration-200"
          :class="optionClass(optionState(current, val === '正确' ? 'A' : 'B'))"
          @click="selectAnswer(current.id, val === '正确' ? 'A' : 'B')"
        >
          <svg v-if="val === '正确'" class="h-5 w-5" :class="optionState(current, 'A') === 'correct' ? 'text-emerald-500' : optionState(current, 'A') === 'wrong' ? 'text-red-500' : 'text-gray-400'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <svg v-else class="h-5 w-5" :class="optionState(current, 'B') === 'correct' ? 'text-emerald-500' : optionState(current, 'B') === 'wrong' ? 'text-red-500' : 'text-gray-400'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ val }}
        </button>
      </div>

      <!-- 填空 / 简答 / 综合 / 论文 / 案例分析 -->
      <div v-else>
        <textarea
          v-model="answers[current.id]"
          :disabled="!!submitted[current.id]"
          :rows="current.type === 'essay' || current.type === 'case_study' ? 8 : 4"
          class="w-full rounded-xl border border-gray-300 bg-gray-50/50 px-4 py-3 text-sm leading-6 transition-colors focus:border-indigo-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500/20 disabled:bg-gray-100"
          :class="{
            'border-emerald-500 bg-emerald-50/50': submitted[current.id] && current.type !== 'essay' && current.type !== 'case_study' && isCorrect(current),
            'border-red-400 bg-red-50/50': submitted[current.id] && current.type !== 'essay' && current.type !== 'case_study' && answers[current.id] && !isCorrect(current),
          }"
          :placeholder="current.type === 'essay' ? '请在此撰写论文正文…' : current.type === 'case_study' ? '请结合上方案例材料作答…' : '请输入答案'"
        />
        <p v-if="current.type === 'essay' || current.type === 'case_study'" class="mt-2 text-xs text-gray-400">
          {{ current.type === 'essay' ? '论文题无标准答案，作答后仅记录提交内容，不自动判分。' : '案例分析题为主观题，作答后仅展示参考答案与解析，不自动判分。' }}
        </p>
      </div>

      <!-- 解析（提交后显示） -->
      <div
        v-if="submitted[current.id]"
        class="mt-6 rounded-xl border-l-4 p-5"
        :class="isSubjective(current)
          ? 'border-indigo-400 bg-indigo-50/50'
          : isCorrect(current) ? 'border-emerald-400 bg-emerald-50/50' : 'border-red-400 bg-red-50/50'"
      >
        <div class="mb-3 flex items-center gap-2">
          <span
            class="flex h-6 w-6 items-center justify-center rounded-full text-white"
            :class="isSubjective(current) ? 'bg-indigo-500' : isCorrect(current) ? 'bg-emerald-500' : 'bg-red-500'"
          >
            <svg v-if="isSubjective(current) || isCorrect(current)" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
            <svg v-else class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </span>
          <span
            class="text-sm font-semibold"
            :class="isSubjective(current) ? 'text-indigo-700' : isCorrect(current) ? 'text-emerald-700' : 'text-red-700'"
          >
            {{ isSubjective(current) ? '已提交，不判分' : isCorrect(current) ? '回答正确' : '回答错误' }}
          </span>
        </div>

        <div class="space-y-2 text-sm">
          <div class="flex gap-2">
            <span class="shrink-0 text-gray-500">你的答案</span>
            <span class="font-medium text-gray-800">{{ answeredText(current) }}</span>
          </div>
          <div v-if="current.type !== 'essay' && !isCorrect(current)" class="flex gap-2">
            <span class="shrink-0 text-gray-500">正确答案</span>
            <span class="font-medium text-emerald-700">{{ current.answer }}</span>
          </div>
          <div v-if="current.type !== 'essay' && current.analysis" class="rounded-lg bg-white/70 p-3 leading-6 text-gray-600">
            <span class="font-medium text-gray-800">解析：</span>{{ current.analysis }}
          </div>
        </div>

        <!-- AI 评分按钮与结果（论文 / 案例分析） -->
        <div v-if="isSubjective(current)" class="mt-4 border-t border-indigo-100 pt-4">
          <button
            v-if="!aiScoring[current.id] && !aiScoreResults[current.id] && !practiceRecordIds[current.id]"
            class="cursor-pointer rounded-lg px-4 py-2 text-sm font-medium text-gray-400 ring-1 ring-inset ring-gray-200"
            disabled
          >
            记录提交中…
          </button>
          <button
            v-else-if="!aiScoring[current.id] && !aiScoreResults[current.id]"
            class="bg-brand-gradient cursor-pointer rounded-lg px-4 py-2 text-sm font-medium text-white shadow-sm transition-all duration-200 hover:shadow-md hover:brightness-110"
            @click="handleAiScore"
          >
            {{ currentAiSpec.title }}
          </button>
          <span v-if="aiScoring[current.id]" class="inline-flex items-center gap-2 text-sm text-indigo-600">
            <span class="h-4 w-4 animate-spin rounded-full border-2 border-indigo-600 border-t-transparent" />
            AI 正在评分中…
          </span>
          <div v-if="aiScoreError[current.id]" class="mt-2 text-sm text-red-500">{{ aiScoreError[current.id] }}</div>

          <!-- 评分结果展示 -->
          <div v-if="aiScoreResults[current.id]" class="mt-3 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm font-semibold text-gray-800">AI 评分结果</span>
              <button
                class="cursor-pointer text-xs font-medium text-indigo-600 transition-colors hover:text-indigo-800"
                @click="handleAiScore"
                :disabled="aiScoring[current.id]"
              >
                {{ aiScoring[current.id] ? '评分中…' : '重新评分' }}
              </button>
            </div>
            <!-- 总分 -->
            <div class="flex items-center gap-3 rounded-xl bg-white p-4 shadow-sm">
              <span class="text-gradient text-3xl font-bold">{{ aiScoreResults[current.id].total_score }}</span>
              <span class="text-xs text-gray-400">/ {{ currentAiSpec.totalMax }} 分</span>
              <div class="h-2 flex-1 overflow-hidden rounded-full bg-gray-100">
                <div
                  class="h-full rounded-full bg-gradient-to-r from-indigo-500 to-violet-500 transition-all duration-500"
                  :style="{ width: (aiScoreResults[current.id].total_score / currentAiSpec.totalMax * 100) + '%' }"
                />
              </div>
            </div>
            <!-- 分维度评分 -->
            <div class="grid grid-cols-2 gap-2">
              <div v-for="dim in currentAiSpec.dimensions" :key="dim.key" class="rounded-xl bg-white p-3 shadow-sm">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-gray-500">{{ dim.label }}</span>
                  <span class="font-semibold text-indigo-600">{{ aiDimensionScore(aiScoreResults[current.id], dim.key) }}/{{ dim.max }}</span>
                </div>
                <div class="mt-1.5 h-1.5 overflow-hidden rounded-full bg-gray-100">
                  <div
                    class="h-full rounded-full bg-gradient-to-r from-indigo-400 to-violet-400"
                    :style="{ width: (aiDimensionScore(aiScoreResults[current.id], dim.key) / dim.max * 100) + '%' }"
                  />
                </div>
              </div>
            </div>
            <!-- 评语 -->
            <div class="rounded-xl bg-white p-4 shadow-sm">
              <p class="mb-1.5 text-xs font-semibold text-gray-500">AI 评语</p>
              <p class="whitespace-pre-wrap text-sm leading-relaxed text-gray-700">{{ aiScoreResults[current.id].comment }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>