<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import { startExam, submitAnswer, submitExam } from '@/api/exam'
import { useExamStore } from '@/stores/exam'
import { sanitizeHtml } from '@/utils/sanitize'
import { parseOptions, typeLabel } from '@/utils/question'
import { showToast } from '@/utils/toast'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import type { StartExamResp, ExamQuesResp } from '@/types/exam'

const route = useRoute()
const router = useRouter()
const examStore = useExamStore()
const examData = ref<StartExamResp | null>(null)
const isAiMode = computed(() => route.query.source === 'ai')
const currentIndex = ref(0)
const loading = ref(true)
const error = ref('')
const showCard = ref(false)
const submitting = ref(false)
const submitted = ref(false)
const caseExpanded = ref(true)
const confirmDialog = ref<InstanceType<typeof ConfirmDialog> | null>(null)
const remainingSeconds = ref(0)
// 已确认完成的题目（单选/判断选中即完成；多选/主观题手动完成）
const completed = ref<Record<number, boolean>>({})
let timer: number | null = null
let deadline = 0
let lastSaveErrorAt = 0

const current = computed((): ExamQuesResp => examData.value?.questions[currentIndex.value] as ExamQuesResp)
const total = computed(() => examData.value?.questions.length || 0)
const answers = ref<Record<number, string>>({})

const answeredCount = computed(() => Object.keys(answers.value).filter(k => (answers.value as Record<string, string>)[k]).length)
const progressPct = computed(() => total.value ? Math.round((answeredCount.value / total.value) * 100) : 0)
const timeColor = computed(() => {
  if (remainingSeconds.value <= 0) return 'text-gray-700'
  if (remainingSeconds.value < 300) return 'text-red-500 animate-pulse'
  if (remainingSeconds.value < 600) return 'text-yellow-600'
  return 'text-gray-700'
})

onMounted(async () => {
  window.addEventListener('beforeunload', handleBeforeUnload)
  try {
    if (isAiMode.value) {
      const data = examStore.getAndClearAiExamData()
      if (!data) {
        error.value = 'AI 考试数据丢失，请重新组卷'
        return
      }
      examData.value = data
    } else {
      const id = Number(route.params.id)
      examData.value = await startExam(id)
    }
    if (examData.value.answers) answers.value = { ...examData.value.answers }
    const elapsed = examData.value.elapsed || 0
    remainingSeconds.value = examData.value.duration * 60 - elapsed
    if (examData.value.status === 'finished') {
      submitted.value = true
      showToast('考试时间已到，系统已自动交卷')
      router.replace(`/exam/${examData.value.record_id}/result`)
      return
    }
    startTimer()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  stopTimer()
  window.removeEventListener('beforeunload', handleBeforeUnload)
})

onBeforeRouteLeave(() => {
  if (submitted.value || submitting.value) return true
  return window.confirm('考试尚未交卷，离开后将无法继续本次考试，确定离开吗？')
})

function handleBeforeUnload(e: BeforeUnloadEvent) {
  if (submitted.value || submitting.value) return
  e.preventDefault()
  e.returnValue = ''
}

function startTimer() {
  deadline = Date.now() + remainingSeconds.value * 1000
  tick()
  timer = window.setInterval(tick, 1000)
}

function tick() {
  remainingSeconds.value = Math.max(0, Math.round((deadline - Date.now()) / 1000))
  if (remainingSeconds.value <= 0) {
    stopTimer()
    handleSubmit()
  }
}

function stopTimer() {
  if (timer !== null) {
    clearInterval(timer)
    timer = null
  }
}

function formatTime(sec: number): string {
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  const mm = String(m).padStart(2, '0')
  const ss = String(s).padStart(2, '0')
  return h > 0 ? `${String(h).padStart(2, '0')}:${mm}:${ss}` : `${mm}:${ss}`
}

function goTo(index: number) {
  if (index >= 0 && index < total.value) currentIndex.value = index
}

function selectAnswer(questionId: number, value: string) {
  const q = examData.value?.questions.find(x => x.id === questionId)
  if (!q) return
  if (q.type === 'multi') {
    const cur = answers.value[questionId] ? answers.value[questionId].split(',') : []
    const idx = cur.indexOf(value)
    if (idx >= 0) cur.splice(idx, 1)
    else cur.push(value)
    cur.sort()
    answers.value[questionId] = cur.join(',')
  } else {
    answers.value[questionId] = value
  }
  saveAnswer(questionId, answers.value[questionId])
  // 单选 / 判断题：选中后自动标记完成并跳到下一题
  if ((q.type === 'single' || q.type === 'judge') && !completed.value[questionId]) {
    completed.value[questionId] = true
    goToNextUnanswered()
  }
}

// 标记当前题完成（多选 / 主观题手动触发）
function completeQuestion() {
  if (!current.value) return
  completed.value[current.value.id] = true
  if (answers.value[current.value.id]) saveAnswer(current.value.id, answers.value[current.value.id])
  goToNextUnanswered()
}

// 跳转到下一个未答或未完成的题目
function goToNextUnanswered() {
  const questions = examData.value?.questions || []
  for (let i = currentIndex.value + 1; i < questions.length; i++) {
    const q = questions[i]
    if (!answers.value[q.id]) {
      currentIndex.value = i
      return
    }
  }
  showToast('已是最后一题，可点击交卷', 'info')
}

function saveAnswer(questionId: number, answer: string) {
  if (!examData.value) return
  submitAnswer(examData.value.record_id, { question_id: questionId, answer }).catch(() => {
    const now = Date.now()
    if (now - lastSaveErrorAt > 3000) {
      lastSaveErrorAt = now
      showToast('答案保存失败，请检查网络后重试')
    }
  })
}

function isSelected(questionId: number, value: string) {
  return (answers.value[questionId] || '').split(',').includes(value)
}

function statusClass(questionId: number) {
  if (completed.value[questionId]) return 'bg-emerald-500 text-white'
  if (answers.value[questionId]) return 'bg-indigo-600 text-white'
  return 'bg-gray-200 text-gray-600'
}

function askSubmit() {
  confirmDialog.value?.open()
}

async function handleSubmit() {
  if (!examData.value || submitting.value) return
  submitting.value = true
  stopTimer()
  try {
    await submitExam(examData.value.record_id)
    submitted.value = true
    router.replace(`/exam/${examData.value.record_id}/result`)
  } catch (e) {
    submitting.value = false
    showToast((e as Error).message || '交卷失败，请重试')
  }
}
</script>

<template>
  <div class="mx-auto max-w-6xl">
    <BaseSkeleton v-if="loading" variant="question" />

    <div v-else-if="error" class="rounded-xl border border-red-100 bg-red-50 p-6">
      <p class="text-sm text-red-600">{{ error }}</p>
      <button class="mt-4 cursor-pointer text-sm font-medium text-indigo-600 hover:underline" @click="router.push('/exam')">返回试卷列表</button>
    </div>

    <div v-else-if="!examData" class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
      <p class="text-sm text-gray-500">试卷加载失败</p>
    </div>

    <template v-else>
      <!-- 顶部：状态栏 + 导航（粘性） -->
      <div class="sticky top-0 z-20 mb-4 space-y-2">
        <div class="rounded-xl border border-gray-100 bg-white/95 p-3 shadow-sm backdrop-blur">
          <div class="flex items-center justify-between">
            <div class="flex min-w-0 items-center gap-3">
              <button
                class="cursor-pointer rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
                title="返回列表"
                @click="router.push('/exam')"
              >
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              <span class="truncate text-sm font-semibold text-gray-800">{{ examData.template_name }}</span>
            </div>
            <div class="flex items-center gap-3">
              <div class="hidden items-center gap-2 sm:flex">
                <div class="h-1.5 w-24 overflow-hidden rounded-full bg-gray-100">
                  <div class="h-full rounded-full bg-gradient-to-r from-indigo-500 to-violet-500 transition-all duration-300" :style="{ width: progressPct + '%' }" />
                </div>
                <span class="text-xs text-gray-400">{{ answeredCount }}/{{ total }}</span>
              </div>
              <span
                class="inline-flex items-center gap-1.5 rounded-full bg-gray-50 px-3 py-1 font-mono text-sm font-bold ring-1 ring-inset ring-gray-200"
                :class="timeColor"
              >
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {{ formatTime(remainingSeconds) }}
              </span>
              <button
                class="cursor-pointer rounded-lg border px-2.5 py-1 text-xs font-medium transition-colors lg:hidden"
                :class="showCard ? 'border-indigo-300 bg-indigo-50 text-indigo-600' : 'border-gray-200 bg-white text-gray-600'"
                @click="showCard = !showCard"
              >
                答题卡
              </button>
              <button
                class="cursor-pointer rounded-lg bg-red-500 px-3.5 py-1.5 text-xs font-semibold text-white shadow-sm transition-all duration-200 hover:bg-red-600 hover:shadow-md active:scale-[0.98] disabled:opacity-50"
                :disabled="submitting"
                @click="askSubmit"
              >
                {{ submitting ? '提交中…' : '交卷' }}
              </button>
            </div>
          </div>
        </div>

        <!-- 上一题/下一题已上移到题目卡片上方 -->
      </div>

      <div class="flex flex-col gap-4 lg:flex-row">
        <!-- 题目区 -->
        <div class="min-w-0 flex-1">
          <!-- 题目操作栏：上一题/下一题/标记/收藏/完成本题/提交答案（题目卡片外正上方） -->
          <div class="mb-3 flex flex-wrap items-center justify-between gap-2 rounded-xl border border-gray-100 bg-white/95 p-3 shadow-sm backdrop-blur">
            <div class="flex gap-2">
              <button
                class="flex cursor-pointer items-center gap-1 rounded-lg border border-gray-200 bg-white px-3.5 py-2 text-sm font-medium text-gray-600 transition-all duration-200 hover:border-gray-300 hover:bg-gray-50 disabled:opacity-30"
                :disabled="currentIndex === 0"
                @click="goTo(currentIndex - 1)"
              >
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
                </svg>
                上一题
              </button>
              <button
                class="flex cursor-pointer items-center gap-1 rounded-lg border border-gray-200 bg-white px-3.5 py-2 text-sm font-medium text-gray-600 transition-all duration-200 hover:border-gray-300 hover:bg-gray-50 disabled:opacity-30"
                :disabled="currentIndex === total - 1"
                @click="goTo(currentIndex + 1)"
              >
                下一题
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                </svg>
              </button>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <span
                v-if="completed[current.id]"
                class="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-600 ring-1 ring-inset ring-emerald-600/20"
              >
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                本题已完成
              </span>
              <button
                v-if="!completed[current.id] && current.type !== 'single' && current.type !== 'judge'"
                class="bg-brand-gradient cursor-pointer rounded-lg px-5 py-2 text-sm font-semibold text-white shadow-md shadow-indigo-600/25 transition-all duration-200 hover:shadow-lg hover:brightness-110 active:scale-[0.98] disabled:opacity-40 disabled:shadow-none"
                :disabled="!answers[current.id]"
                @click="completeQuestion"
              >
                完成本题
              </button>
            </div>
          </div>

          <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm sm:p-7">
            <div class="mb-5 flex flex-wrap items-center gap-2">
              <span class="inline-flex items-center rounded-full bg-indigo-50 px-2.5 py-0.5 text-xs font-medium text-indigo-600 ring-1 ring-inset ring-indigo-600/20">
                {{ currentIndex + 1 }} / {{ total }} · {{ typeLabel(current.type) }}
              </span>
              <span class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10">
                {{ current.score }} 分
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

            <div class="mb-6 text-[15px] leading-7 text-gray-800" v-html="sanitizeHtml(current.content)" />

            <div v-if="current.type === 'single' || current.type === 'multi'" class="space-y-2.5">
              <p v-if="current.type === 'multi'" class="mb-1 text-xs text-gray-400">本题为多选题，可选择多个选项</p>
              <button
                v-for="opt in parseOptions(current.options)"
                :key="opt.label"
                class="flex w-full cursor-pointer items-center gap-3.5 rounded-xl border-2 px-4 py-3.5 text-left text-sm transition-all duration-200"
                :class="isSelected(current.id, opt.label)
                  ? 'border-indigo-500 bg-indigo-50/70 shadow-sm'
                  : 'border-gray-200 hover:border-indigo-300 hover:bg-indigo-50/40 hover:shadow-sm'"
                @click="selectAnswer(current.id, opt.label)"
              >
                <span
                  class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border-2 text-xs font-semibold transition-colors"
                  :class="isSelected(current.id, opt.label) ? 'border-indigo-500 bg-indigo-500 text-white' : 'border-gray-300 text-gray-500'"
                >{{ opt.label }}</span>
                <span class="leading-6">{{ opt.text }}</span>
              </button>
            </div>

            <div v-else-if="current.type === 'judge'" class="grid grid-cols-2 gap-3">
              <button
                v-for="val in ['正确', '错误']"
                :key="val"
                class="flex cursor-pointer items-center justify-center gap-2 rounded-xl border-2 px-6 py-4 text-sm font-semibold transition-all duration-200"
                :class="isSelected(current.id, val)
                  ? 'border-indigo-500 bg-indigo-50/70 text-indigo-700 shadow-sm'
                  : 'border-gray-200 hover:border-indigo-300 hover:bg-indigo-50/40 hover:shadow-sm'"
                @click="selectAnswer(current.id, val)"
              >
                <svg v-if="val === '正确'" class="h-5 w-5" :class="isSelected(current.id, val) ? 'text-indigo-500' : 'text-gray-400'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <svg v-else class="h-5 w-5" :class="isSelected(current.id, val) ? 'text-indigo-500' : 'text-gray-400'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {{ val }}
              </button>
            </div>

            <div v-else>
              <textarea
                :value="answers[current.id] || ''"
                :rows="current.type === 'essay' || current.type === 'case_study' ? 10 : 4"
                class="w-full rounded-xl border border-gray-300 bg-gray-50/50 px-4 py-3 text-sm leading-6 transition-colors focus:border-indigo-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                :placeholder="current.type === 'essay' ? '请在此撰写论文正文…' : current.type === 'case_study' ? '请结合上方案例材料作答…' : '请输入答案'"
                @input="selectAnswer(current.id, ($event.target as HTMLTextAreaElement).value)"
              />
              <p v-if="current.type === 'essay' || current.type === 'case_study'" class="mt-2 text-xs text-gray-400">
                {{ current.type === 'essay' ? '论文题无标准答案，作答后仅记录提交内容，不自动判分。' : '案例分析题为主观题，作答后仅记录提交内容，不自动判分。' }}
              </p>
            </div>
          </div>
        </div>

        <!-- 答题卡（桌面常驻 / 移动端抽屉） -->
        <div
          class="h-fit shrink-0 rounded-xl border border-gray-100 bg-white p-4 shadow-sm lg:sticky lg:top-32 lg:w-64"
          :class="showCard ? 'block' : 'hidden lg:block'"
        >
          <div class="mb-3 flex items-center justify-between">
            <h3 class="text-sm font-semibold text-gray-800">答题卡</h3>
            <button class="cursor-pointer text-xs text-gray-400 hover:text-gray-600 lg:hidden" @click="showCard = false">收起</button>
          </div>
          <div class="mb-3 flex flex-wrap gap-2 text-[10px] text-gray-400">
            <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded-full bg-emerald-500" />已完成</span>
            <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded-full bg-indigo-500" />已答</span>
            <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded-full bg-gray-200" />未答</span>
          </div>
          <div class="grid grid-cols-8 gap-1.5 lg:grid-cols-5">
            <button
              v-for="(q, idx) in examData.questions"
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
    </template>

    <ConfirmDialog
      ref="confirmDialog"
      title="确认交卷？"
      description="交卷后无法修改答案，请确认已答完所有题目。"
      confirm-text="确认交卷"
      danger
      @confirm="handleSubmit"
    />
  </div>
</template>
