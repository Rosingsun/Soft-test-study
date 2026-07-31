<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { startExam, submitAnswer, submitExam } from '@/api/exam'
import { sanitizeHtml } from '@/utils/sanitize'
import { parseOptions, typeLabel } from '@/utils/question'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import type { StartExamResp, ExamQuesResp } from '@/types/exam'

const route = useRoute()
const router = useRouter()
const examData = ref<StartExamResp | null>(null)
const currentIndex = ref(0)
const loading = ref(true)
const error = ref('')
const showCard = ref(false)
const submitting = ref(false)
const confirmDialog = ref<InstanceType<typeof ConfirmDialog> | null>(null)
const remainingSeconds = ref(0)
let timer: number | null = null

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
  try {
    const id = Number(route.params.id)
    examData.value = await startExam(id)
    if (examData.value.answers) answers.value = { ...examData.value.answers }
    const elapsed = examData.value.elapsed || 0
    remainingSeconds.value = examData.value.duration * 60 - elapsed
    if (examData.value.status === 'finished') {
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

onUnmounted(stopTimer)

function startTimer() {
  timer = window.setInterval(() => {
    remainingSeconds.value--
    if (remainingSeconds.value <= 0) {
      stopTimer()
      handleSubmit()
    }
  }, 1000)
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
}

function saveAnswer(questionId: number, answer: string) {
  if (!examData.value) return
  submitAnswer(examData.value.record_id, { question_id: questionId, answer }).catch(() => {})
}

function isSelected(questionId: number, value: string) {
  return (answers.value[questionId] || '').split(',').includes(value)
}

function statusClass(questionId: number) {
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
    router.replace(`/exam/${examData.value.record_id}/result`)
  } catch {
    submitting.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-6xl">
    <div v-if="loading">
      <BaseLoading />
    </div>

    <div v-else-if="error" class="rounded-lg bg-white p-6 shadow-sm">
      <p class="text-red-500">{{ error }}</p>
      <button class="mt-4 text-sm text-indigo-600 hover:underline" @click="router.push('/exam')">返回试卷列表</button>
    </div>

    <div v-else-if="!examData" class="rounded-lg bg-white p-6 shadow-sm">
      <p class="text-gray-500">试卷加载失败</p>
    </div>

    <template v-else>
      <!-- 顶部状态栏 -->
      <div class="mb-4 rounded-lg border border-gray-200 bg-white p-3 shadow-sm">
        <div class="flex items-center justify-between">
          <div class="flex min-w-0 items-center gap-3">
            <button
              class="cursor-pointer rounded-lg p-1.5 text-gray-500 hover:bg-gray-100"
              title="返回列表"
              @click="router.push('/exam')"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <span class="truncate text-sm font-medium text-gray-800">{{ examData.template_name }}</span>
          </div>
          <div class="flex items-center gap-3">
            <div class="hidden items-center gap-1.5 sm:flex">
              <div class="h-1.5 w-24 overflow-hidden rounded-full bg-gray-100">
                <div class="h-full rounded-full bg-indigo-500 transition-all duration-300" :style="{ width: progressPct + '%' }" />
              </div>
              <span class="text-xs text-gray-400">{{ answeredCount }}/{{ total }}</span>
            </div>
            <span class="font-mono text-sm font-bold" :class="timeColor">{{ formatTime(remainingSeconds) }}</span>
            <button
              class="cursor-pointer rounded-lg border px-2.5 py-1 text-xs lg:hidden"
              :class="showCard ? 'border-indigo-300 bg-indigo-50 text-indigo-600' : 'border-gray-300 bg-white text-gray-600'"
              @click="showCard = !showCard"
            >
              答题卡
            </button>
            <button
              class="cursor-pointer rounded-lg bg-red-500 px-3 py-1.5 text-xs font-medium text-white hover:bg-red-600 disabled:opacity-50"
              :disabled="submitting"
              @click="askSubmit"
            >
              {{ submitting ? '提交中…' : '交卷' }}
            </button>
          </div>
        </div>
      </div>

      <div class="flex flex-col gap-4 lg:flex-row">
        <!-- 题目区 -->
        <div class="min-w-0 flex-1">
          <div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
            <div class="mb-4 flex flex-wrap items-center gap-2">
              <span class="rounded-lg bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-600">
                {{ currentIndex + 1 }} / {{ total }} · {{ typeLabel(current.type) }}
              </span>
              <span class="text-xs text-gray-400">{{ current.score }} 分</span>
            </div>

            <div class="mb-6 text-base leading-relaxed text-gray-800" v-html="sanitizeHtml(current.content)" />

            <div v-if="current.type === 'single' || current.type === 'multi'" class="space-y-2">
              <button
                v-for="opt in parseOptions(current.options)"
                :key="opt.label"
                class="flex w-full cursor-pointer items-center gap-3 rounded-lg border px-4 py-3 text-left text-sm transition-colors"
                :class="isSelected(current.id, opt.label) ? 'border-indigo-500 bg-indigo-50' : 'border-gray-200 hover:border-indigo-300'"
                @click="selectAnswer(current.id, opt.label)"
              >
                <span
                  class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-xs font-medium"
                  :class="isSelected(current.id, opt.label) ? 'border-indigo-500 bg-indigo-500 text-white' : 'border-gray-300'"
                >{{ opt.label }}</span>
                <span>{{ opt.text }}</span>
              </button>
            </div>

            <div v-else-if="current.type === 'judge'" class="flex gap-4">
              <button
                v-for="val in ['正确', '错误']"
                :key="val"
                class="flex-1 cursor-pointer rounded-lg border px-6 py-3 text-center text-sm font-medium transition-colors"
                :class="isSelected(current.id, val) ? 'border-indigo-500 bg-indigo-50 text-indigo-700' : 'border-gray-200 hover:border-indigo-300'"
                @click="selectAnswer(current.id, val)"
              >
                {{ val }}
              </button>
            </div>

            <div v-else>
              <textarea
                :value="answers[current.id] || ''"
                rows="3"
                class="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                placeholder="请输入答案"
                @input="selectAnswer(current.id, ($event.target as HTMLTextAreaElement).value)"
              />
            </div>
          </div>

          <div class="mt-4 flex items-center justify-between rounded-lg border border-gray-200 bg-white p-3 shadow-sm">
            <button
              class="cursor-pointer rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50 disabled:opacity-30"
              :disabled="currentIndex === 0"
              @click="goTo(currentIndex - 1)"
            >
              ← 上一题
            </button>
            <span class="text-xs text-gray-400">答案已自动保存</span>
            <button
              class="cursor-pointer rounded-lg bg-indigo-600 px-3 py-1.5 text-sm text-white hover:bg-indigo-700 disabled:opacity-30"
              :disabled="currentIndex === total - 1"
              @click="goTo(currentIndex + 1)"
            >
              下一题 →
            </button>
          </div>
        </div>

        <!-- 答题卡（桌面常驻 / 移动端抽屉） -->
        <div
          class="shrink-0 rounded-lg border border-gray-200 bg-white p-4 shadow-sm lg:w-64"
          :class="showCard ? 'block' : 'hidden lg:block'"
        >
          <div class="mb-3 flex items-center justify-between">
            <h3 class="text-sm font-semibold text-gray-800">答题卡</h3>
            <button class="cursor-pointer text-xs text-gray-400 lg:hidden" @click="showCard = false">收起</button>
          </div>
          <div class="mb-3 flex flex-wrap gap-1.5 text-[10px] text-gray-400">
            <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded bg-indigo-500" />已答</span>
            <span class="flex items-center gap-1"><span class="inline-block h-2.5 w-2.5 rounded bg-gray-200" />未答</span>
          </div>
          <div class="grid grid-cols-8 gap-1.5 lg:grid-cols-5">
            <button
              v-for="(q, idx) in examData.questions"
              :key="q.id"
              class="flex h-8 w-8 cursor-pointer items-center justify-center rounded text-xs font-medium transition-colors"
              :class="[statusClass(q.id), { 'ring-2 ring-indigo-400 ring-offset-1': idx === currentIndex }]"
              @click="goTo(idx)"
            >
              {{ idx + 1 }}
            </button>
          </div>
          <div class="mt-3 rounded-lg bg-gray-50 p-3 text-center">
            <p class="text-xs text-gray-500">已答 <span class="font-bold text-indigo-600">{{ answeredCount }}</span> / {{ total }}</p>
            <p class="mt-0.5 text-[10px] text-gray-400">{{ Math.round(progressPct) }}% 完成</p>
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
