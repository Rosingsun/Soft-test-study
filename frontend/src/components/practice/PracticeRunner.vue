<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { submitPractice } from '@/api/practice'
import { addFavorite, removeFavorite, checkFavorited } from '@/api/bookmark'
import { addMark, removeMark, checkMarked } from '@/api/mark'
import { essayScore, checkEssayScore } from '@/api/ai'
import { useAiStore } from '@/stores/ai'
import { sanitizeHtml } from '@/utils/sanitize'
import { typeLabel, parseOptions, parseBlankOptions, isCorrectAnswer, aiScoreSpec, aiDimensionScore, isAiSource, sourceLabel } from '@/utils/question'
import type { Question, PracticeRecordResp } from '@/types/question'
import type { EssayScoreResp } from '@/types/ai'
import ExtractKnowledgeModal from '@/components/knowledge/ExtractKnowledgeModal.vue'

const router = useRouter()
const aiStore = useAiStore()

const props = withDefaults(defineProps<{
  questions: Question[]
  mode?: string
  title?: string
  startIndex?: number
  initialAnswers?: Record<number, string>
  initialSubmitted?: Record<number, boolean>
  submitHandler?: (questionId: number, answer: string, duration: number) => Promise<PracticeRecordResp>
}>(), {
  mode: 'chapter',
  title: '',
  startIndex: 0,
})

const emit = defineEmits<{ (e: 'back'): void; (e: 'finish'): void }>()

// 单选题答完后是否自动跳转到下一题。默认开启，localStorage 持久化。
// 仅作用于「选中后 1s 跳下一题」逻辑；自动提交与解析展示不受影响。
const AUTO_NEXT_KEY = 'practice_auto_next'
const autoNext = ref(localStorage.getItem(AUTO_NEXT_KEY) !== '0')
function toggleAutoNext() {
  autoNext.value = !autoNext.value
  localStorage.setItem(AUTO_NEXT_KEY, autoNext.value ? '1' : '0')
}

const currentIndex = ref(props.startIndex ?? 0)
const answers = ref<Record<number, string>>({ ...(props.initialAnswers || {}) })
const submitted = ref<Record<number, boolean>>({ ...(props.initialSubmitted || {}) })
// 每道题首次进入的时间戳，用于统计本题作答耗时
const questionStart = ref<Record<number, number>>({})
const marked = ref<Record<number, boolean>>({})
const favorited = ref<Record<number, boolean>>({})
const showCard = ref(false)
const loadingFavorites = ref(true)
const caseExpanded = ref(true)
// 案例分析题：把单题答案拆成若干个小节（“小的文章”）分别作答，提交时合并为一段
const caseParts = ref<Record<number, string[]>>({})
const CASE_PART_COUNT = 3

function ensureCaseParts(q: Question): string[] {
  if (q.type !== 'case_study') return []
  let parts = caseParts.value[q.id] || []
  const existing = answers.value[q.id] || ''
  if (!parts.length) {
    // 尝试从已存答案恢复（历史提交 / 重新进入）
    if (existing) {
      parts = existing.split(/\r?\n\s*\n/).filter(Boolean)
    } else {
      parts = Array.from({ length: CASE_PART_COUNT }, () => '')
    }
    caseParts.value[q.id] = parts
  }
  return parts
}

function updateCaseAnswers(questionId: number, parts: string[]) {
  caseParts.value[questionId] = parts
  const joined = parts.map(p => p.trim()).filter(Boolean)
  answers.value[questionId] = joined.join('\n\n')
}

// 分段作答的某一小节变化
function onCaseInput(q: Question, idx: number, val: string) {
  const parts = ensureCaseParts(q)
  parts[idx] = val
  updateCaseAnswers(q.id, parts)
}

const current = computed(() => props.questions[currentIndex.value])
const total = computed(() => props.questions.length)
const pageTitle = computed(() => {
  const map: Record<string, string> = {
    special: '专项练习',
    random: '随机练习',
    chapter: '章节练习',
    wrong: '错题练习',
    ai: 'AI 练习',
  }
  return map[props.mode] || '练习'
})

const progressText = computed(() => `第 ${currentIndex.value + 1} 题 / 共 ${total.value} 题`)
const progressPct = computed(() => total.value ? ((currentIndex.value + 1) / total.value) * 100 : 0)
const answeredCount = computed(() => Object.keys(answers.value).length)
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

function ensureTiming(qid: number) {
  if (!questionStart.value[qid]) questionStart.value[qid] = Date.now()
}

watch(() => [currentIndex.value, props.questions], () => {
  const q = props.questions[currentIndex.value]
  if (q) ensureTiming(q.id)
}, { immediate: true })

function goTo(index: number) {
  if (index >= 0 && index < total.value) currentIndex.value = index
}

function handleNext() {
  if (currentIndex.value === total.value - 1) {
    emit('finish')
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
    // 单选题：显示答案后自动跳转下一题（多选、输入框不自动跳转）
    if (q.type === 'single' && autoNext.value) {
      const fromIndex = currentIndex.value
      setTimeout(() => {
        // 用户已手动切换题目则不再自动跳转
        if (currentIndex.value !== fromIndex) return
        if (currentIndex.value === total.value - 1) {
          emit('finish')
          emit('back')
          return
        }
        goTo(currentIndex.value + 1)
      }, 1000)
    }
  }
}

// ===== 多空题（multi_blank）=====
// 答案以 JSON 数组字符串存储（如 '["A","C"]'），每个空为独立单选

function blankList(q: Question): { blank_index: number; options: { id: string; content: string }[] }[] {
  return parseBlankOptions(q.blank_options)
}

// 取某空已选中的选项 id
function blankAnswer(q: Question, blankIndex: number): string {
  const arr = parseBlankAnswerList(answers.value[q.id])
  return arr[blankIndex - 1] || ''
}

// 解析当前已选的答案数组
function parseBlankAnswerList(answer: string | undefined): string[] {
  if (!answer) return []
  const s = answer.trim()
  if (s.startsWith('[')) {
    try {
      const arr = JSON.parse(s)
      if (Array.isArray(arr)) return arr.map(String)
    } catch { /* 忽略 */ }
  }
  return s.split(',').map(x => x.trim()).filter(Boolean)
}

// 选择某空的选项（同空单选，重选覆盖）
function selectBlankAnswer(q: Question, blankIndex: number, optionId: string) {
  if (submitted.value[q.id]) return
  const blanks = blankList(q)
  const arr = parseBlankAnswerList(answers.value[q.id])
  // 保证数组长度与空数一致
  while (arr.length < blanks.length) arr.push('')
  arr[blankIndex - 1] = optionId
  answers.value[q.id] = JSON.stringify(arr)
}

// 某空的选项状态
function blankOptionState(q: Question, blankIndex: number, optionId: string): string {
  const isSub = !!submitted.value[q.id]
  const selected = blankAnswer(q, blankIndex) === optionId
  if (!isSub) return selected ? 'selected' : 'idle'
  // 提交后：从正确答案 JSON 解析该空正确答案
  const correctArr = parseBlankAnswerList(q.answer)
  const correct = correctArr[blankIndex - 1]
  if (optionId === correct) return 'correct'
  if (selected) return 'wrong'
  return 'dim'
}

// 多空题是否全部作答完成
function blankAllAnswered(q: Question): boolean {
  const blanks = blankList(q)
  if (!blanks.length) return false
  const arr = parseBlankAnswerList(answers.value[q.id])
  return blanks.every(b => arr[b.blank_index - 1])
}

// 多空题选项字母徽章样式
function optionBadgeClassForBlank(q: Question, blankIndex: number, optionId: string, state: string): string {
  const selected = blankAnswer(q, blankIndex) === optionId
  if (!submitted.value[q.id] && selected) return 'border-indigo-500 bg-indigo-500 text-white'
  if (state === 'correct') return 'border-emerald-500 bg-emerald-500 text-white'
  if (state === 'wrong') return 'border-red-500 bg-red-500 text-white'
  return 'border-gray-300 text-gray-500'
}

// 正确答案的可读文本（多空题转成每空答案）
function correctAnswerText(q: Question): string {
  if (q.type === 'multi_blank') return parseBlankAnswerList(q.answer).join('、')
  return q.answer
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
// 练习上报状态：用于解决"上报失败 → 永远卡在'记录提交中' → AI 评分按钮不可见"的问题
const submitting = ref<Record<number, boolean>>({})
const submitError = ref<Record<number, string>>({})

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
  if (!recordId) {
    // 练习记录未上报成功（网络异常 / 后端报错），不允许静默退出 AI 评分；
    // 显式提示用户先重新提交练习记录
    const reason = submitError.value[q.id] || '练习记录尚未上报成功'
    aiScoreError.value[q.id] = `AI 评分需要练习记录，请先重新提交练习。原因：${reason}`
    return
  }

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
  if (!q || !answers.value[q.id]) return
  // 已在请求中（防双击/防重入）
  if (submitting.value[q.id]) return
  // 已有练习记录（重做 / 二次进入），仅置已提交位即可
  if (practiceRecordIds.value[q.id]) {
    submitted.value[q.id] = true
    return
  }
  // 乐观提交：先同步置位已提交，立即渲染解析与对错
  submitted.value[q.id] = true
  submitting.value[q.id] = true
  submitError.value[q.id] = ''
  const duration = Math.max(1, Math.round((Date.now() - (questionStart.value[q.id] || Date.now())) / 1000))
  try {
    const submit = props.submitHandler
      ? props.submitHandler
      : (questionId: number, answer: string, _duration: number) => submitPractice({
        question_id: questionId,
        mode: props.mode,
        answer,
        duration: _duration,
      })
    const res = await submit(q.id, answers.value[q.id], duration)
    if (res.answer) answers.value[q.id] = res.answer
    // 保存练习记录ID，供后续 AI 评分使用
    practiceRecordIds.value[q.id] = res.id
  } catch (e) {
    // 上报失败：保留 submitted 状态（用户仍能看到解析），但暴露错误与重提入口，
    // 避免"记录提交中…"永久悬挂、AI 评分按钮永不可用
    submitError.value[q.id] = (e as Error)?.message || '练习记录上报失败，AI 评分暂不可用'
  } finally {
    submitting.value[q.id] = false
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
  if (q.type === 'multi_blank') {
    return parseBlankAnswerList(ans).join('、')
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

const showExtractModal = ref(false)

function stripHtml(html: string) {
  return (html || '').replace(/<[^>]*>/g, '')
}

function openExtractModal() {
  showExtractModal.value = true
}
</script>

<template>
  <div class="mx-auto max-w-6xl">
    <!-- 顶部：状态栏（粘性） -->
    <div class="sticky top-0 z-20 mb-4 space-y-2">
      <div class="rounded-xl border border-gray-100 bg-white/95 p-3 shadow-sm backdrop-blur">
        <div class="flex items-center justify-between">
          <div class="flex min-w-0 items-center gap-3">
            <button
              class="cursor-pointer rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
              title="返回"
              @click="emit('back')"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <span class="truncate text-sm font-semibold text-gray-800">{{ title || pageTitle }}</span>
          </div>
          <div class="flex items-center gap-3">
            <div class="hidden items-center gap-2 sm:flex">
              <div class="h-1.5 w-24 overflow-hidden rounded-full bg-gray-100">
                <div
                  class="h-full rounded-full bg-gradient-to-r from-indigo-500 to-violet-500 transition-all duration-300"
                  :style="{ width: progressPct + '%' }"
                />
              </div>
              <span class="text-xs text-gray-400">{{ answeredCount }}/{{ total }}</span>
            </div>
            <span class="text-xs text-gray-400 sm:hidden">{{ progressText }}</span>
            <button
              class="cursor-pointer rounded-lg border px-2.5 py-1 text-xs font-medium transition-colors"
              :class="showCard ? 'border-indigo-300 bg-indigo-50 text-indigo-600' : 'border-gray-200 bg-white text-gray-600'"
              @click="showCard = !showCard"
            >
              答题卡
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="flex flex-col gap-4 lg:flex-row">
      <!-- 题目区 -->
      <div class="min-w-0 flex-1">

    <!-- 题目操作栏（题目卡片外正上方） -->
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2 rounded-xl border border-gray-100 bg-white/95 p-3 shadow-sm backdrop-blur">
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
          type="button"
          role="switch"
          :aria-checked="autoNext"
          :title="autoNext ? '关闭自动跳转' : '开启自动跳转'"
          class="flex cursor-pointer items-center gap-1.5 rounded-lg border px-3 py-2 text-sm font-medium transition-all duration-200"
          :class="autoNext
            ? 'border-indigo-200 bg-indigo-50 text-indigo-700'
            : 'border-gray-200 bg-white text-gray-500 shadow-sm hover:border-gray-300 hover:bg-gray-50'"
          @click="toggleAutoNext"
        >
          <span
            class="relative inline-flex h-4 w-7 items-center rounded-full transition-colors duration-200"
            :class="autoNext ? 'bg-indigo-500' : 'bg-gray-300'"
          >
            <span
              class="absolute h-3 w-3 rounded-full bg-white shadow transition-transform duration-200"
              :class="autoNext ? 'translate-x-3.5' : 'translate-x-0.5'"
            />
          </span>
          自动跳转
        </button>
        <button
          v-if="current && !submitted[current.id] && current.type !== 'single' && current.type !== 'judge'"
          class="bg-brand-gradient cursor-pointer rounded-lg px-5 py-2 text-sm font-semibold text-white shadow-md shadow-indigo-600/25 transition-all duration-200 hover:shadow-lg hover:brightness-110 active:scale-[0.98] disabled:opacity-40 disabled:shadow-none"
          :disabled="current && current.type === 'multi_blank' ? !blankAllAnswered(current) : !answers[current.id]"
          @click="handleSubmit"
        >
          提交答案
        </button>
      </div>
    </div>

    <!-- 题目卡片 -->
    <div v-if="current" class="min-h-[560px] rounded-xl border border-gray-100 bg-white p-6 shadow-sm sm:p-7">
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
        <span
          v-if="current.source"
          :class="isAiSource(current.source)
            ? 'bg-violet-50 text-violet-600 ring-violet-600/20'
            : 'bg-emerald-50 text-emerald-600 ring-emerald-600/20'"
          class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset"
        >{{ sourceLabel(current.source) }}</span>
        <span v-if="current.year" class="text-xs text-gray-400">{{ current.year }} 年真题</span>
        <span class="text-xs text-gray-400">题号 #{{ current.id }}</span>
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

      <!-- 论文：与模拟考试论文模块一致的作答区 -->
      <div v-else-if="current.type === 'essay'">
        <textarea
          v-model="answers[current.id]"
          :disabled="!!submitted[current.id]"
          :rows="10"
          class="w-full rounded-xl border border-gray-300 bg-gray-50/50 px-4 py-3 text-sm leading-6 transition-colors focus:border-indigo-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500/20 disabled:bg-gray-100"
          placeholder="请在此撰写论文正文…"
        />
        <p class="mt-2 text-xs text-gray-400">论文题无标准答案，作答后仅记录提交内容，不自动判分。</p>
      </div>

      <!-- 案例分析：多个小的篇章分别作答 -->
      <div v-else-if="current.type === 'case_study'">
        <div class="space-y-4">
          <div
            v-for="(part, idx) in ensureCaseParts(current)"
            :key="idx"
            class="rounded-xl border border-gray-300 bg-white p-4 focus-within:border-violet-500 focus-within:ring-2 focus-within:ring-violet-500/20"
          >
            <div class="mb-1.5 flex items-center justify-between">
              <span class="inline-flex items-center gap-1.5 text-xs font-semibold text-violet-700">
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                回答 {{ idx + 1 }}
              </span>
              <span class="text-xs text-gray-400">{{ part.length }} 字</span>
            </div>
            <textarea
              :value="part"
              :disabled="!!submitted[current.id]"
              rows="4"
              class="w-full resize-y bg-transparent px-1 py-1.5 text-sm leading-6 text-gray-800 outline-none placeholder:text-gray-400 disabled:bg-gray-50"
              :placeholder="idx === 0 ? '请结合上方案例材料，分点作答…' : `继续作答第 ${idx + 1} 部分…`"
              @input="onCaseInput(current, idx, ($event.target as HTMLTextAreaElement).value)"
            />
          </div>
        </div>
        <p class="mt-2 text-xs text-gray-400">案例分析题为主观题，可分段（如要点一、要点二…）作答，作答后仅展示参考答案与解析，不自动判分。</p>
      </div>

      <!-- 多空题：每个空为独立单选 -->
      <div v-else-if="current.type === 'multi_blank'" class="space-y-5">
        <p v-if="!submitted[current.id]" class="mb-1 text-xs text-gray-400">本题为多空题，请在下方每个空位选择对应选项</p>
        <div
          v-for="blank in blankList(current)"
          :key="blank.blank_index"
          class="rounded-xl border border-gray-200 p-4"
          :class="submitted[current.id] && blankAnswer(current, blank.blank_index) !== parseBlankAnswerList(current.answer)[blank.blank_index - 1] ? 'border-red-300 bg-red-50/30' : ''"
        >
          <div class="mb-2.5 flex items-center gap-2">
            <span class="inline-flex h-6 w-6 items-center justify-center rounded-full bg-indigo-50 text-xs font-bold text-indigo-600">
              {{ blank.blank_index }}
            </span>
            <span class="text-sm font-medium text-gray-700">第 {{ blank.blank_index }} 空</span>
            <span v-if="submitted[current.id]" class="ml-auto text-xs" :class="blankAnswer(current, blank.blank_index) === parseBlankAnswerList(current.answer)[blank.blank_index - 1] ? 'text-emerald-600' : 'text-red-500'">
              {{ blankAnswer(current, blank.blank_index) === parseBlankAnswerList(current.answer)[blank.blank_index - 1] ? '正确' : '错误' }}
            </span>
          </div>
          <div class="grid gap-2 sm:grid-cols-2">
            <button
              v-for="opt in blank.options"
              :key="opt.id"
              class="flex w-full cursor-pointer items-center gap-3 rounded-xl border-2 px-3.5 py-2.5 text-left text-sm transition-all duration-200"
              :class="optionClass(blankOptionState(current, blank.blank_index, opt.id))"
              @click="selectBlankAnswer(current, blank.blank_index, opt.id)"
            >
              <span
                class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 text-xs font-semibold transition-colors"
                :class="optionBadgeClassForBlank(current, blank.blank_index, opt.id, blankOptionState(current, blank.blank_index, opt.id))"
              >{{ opt.id }}</span>
              <span class="leading-6">{{ opt.content }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 填空 / 简答 / 综合 -->
      <div v-else>
        <textarea
          v-model="answers[current.id]"
          :disabled="!!submitted[current.id]"
          rows="4"
          class="w-full rounded-xl border border-gray-300 bg-gray-50/50 px-4 py-3 text-sm leading-6 transition-colors focus:border-indigo-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500/20 disabled:bg-gray-100"
          :class="{
            'border-emerald-500 bg-emerald-50/50': submitted[current.id] && isCorrect(current),
            'border-red-400 bg-red-50/50': submitted[current.id] && answers[current.id] && !isCorrect(current),
          }"
          placeholder="请输入答案"
        />
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
          <button
            class="ml-auto inline-flex cursor-pointer items-center gap-1 rounded-lg border border-indigo-200 bg-white px-2.5 py-1 text-xs font-medium text-indigo-700 shadow-sm transition-all duration-200 hover:border-indigo-300 hover:bg-indigo-50 active:scale-[0.97]"
            title="AI 提取核心知识点"
            @click="openExtractModal"
          >
            <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
            </svg>
            获取知识点
          </button>
        </div>

        <div class="space-y-2 text-sm">
          <div class="flex gap-2">
            <span class="shrink-0 text-gray-500">你的答案</span>
            <span class="font-medium text-gray-800">{{ answeredText(current) }}</span>
          </div>
          <div v-if="current.type !== 'essay' && !isCorrect(current)" class="flex gap-2">
            <span class="shrink-0 text-gray-500">正确答案</span>
            <span class="font-medium text-emerald-700">{{ correctAnswerText(current) }}</span>
          </div>
          <div v-if="current.type !== 'essay' && current.analysis" class="rounded-lg bg-white/70 p-3 leading-6 text-gray-600">
            <span class="font-medium text-gray-800">解析：</span>{{ current.analysis }}
          </div>
        </div>

        <!-- AI 评分按钮与结果（论文 / 案例分析） -->
        <div v-if="isSubjective(current)" class="mt-4 border-t border-indigo-100 pt-4">
          <!-- 提交失败：明确错误 + 重新提交入口（避免 AI 评分按钮永远不出现） -->
          <div
            v-if="submitError[current.id] && !practiceRecordIds[current.id]"
            class="mb-3 flex flex-wrap items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800"
          >
            <svg class="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
            </svg>
            <span class="flex-1 min-w-0">练习记录未上报，AI 评分暂不可用：{{ submitError[current.id] }}</span>
            <button
              class="cursor-pointer rounded-md bg-amber-600 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-amber-700 disabled:opacity-50"
              :disabled="submitting[current.id]"
              @click="handleSubmit"
            >
              {{ submitting[current.id] ? '重试中…' : '重新提交' }}
            </button>
          </div>
          <button
            v-if="!aiScoring[current.id] && !aiScoreResults[current.id] && !practiceRecordIds[current.id] && !submitError[current.id]"
            class="cursor-pointer rounded-lg px-4 py-2 text-sm font-medium text-gray-400 ring-1 ring-inset ring-gray-200"
            disabled
          >
            记录提交中…
          </button>
          <button
            v-else-if="!aiScoring[current.id] && !aiScoreResults[current.id] && practiceRecordIds[current.id]"
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

      <!-- 答题卡（桌面常驻 / 移动端抽屉） -->
      <div
        class="h-fit shrink-0 rounded-xl border border-gray-100 bg-white p-4 shadow-sm lg:sticky lg:top-24 lg:w-60"
        :class="showCard ? 'block' : 'hidden lg:block'"
      >
        <div class="mb-3 flex items-center justify-between">
          <h3 class="text-sm font-semibold text-gray-800">答题卡</h3>
          <button class="cursor-pointer text-xs text-gray-400 hover:text-gray-600 lg:hidden" @click="showCard = false">收起</button>
        </div>
        <div class="mb-3 flex flex-wrap gap-2 text-[10px] text-gray-400">
          <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded-full bg-emerald-500" />正确</span>
          <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded-full bg-red-500" />错误</span>
          <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded-full bg-indigo-500" />已答</span>
          <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded-full bg-amber-400" />标记</span>
          <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded-full bg-gray-200" />未答</span>
        </div>
        <div class="grid grid-cols-8 gap-1.5 lg:grid-cols-5">
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
        <div class="mt-3 rounded-xl bg-gray-50 p-3 text-center">
          <p class="text-xs text-gray-500">已答 <span class="font-bold text-indigo-600">{{ answeredCount }}</span> / {{ total }}</p>
          <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-gray-200">
            <div class="h-full rounded-full bg-gradient-to-r from-indigo-500 to-violet-500 transition-all duration-300" :style="{ width: progressPct + '%' }" />
          </div>
          <p class="mt-1.5 text-[10px] text-gray-400">{{ Math.round(progressPct) }}% 完成</p>
        </div>
      </div>
    </div>

    <ExtractKnowledgeModal
      v-if="current"
      v-model="showExtractModal"
      :question-id="current.id"
      :question-content="stripHtml(current.content)"
      :question-type="current.type"
      :question-answer="current.answer"
      :question-analysis="current.analysis"
      :subject-id="current.subject_id"
    />
  </div>
</template>