<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAiStore } from '@/stores/ai'
import { generateQuestions } from '@/api/ai'
import { getChapters, getSubSubjects } from '@/api/subject'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import type { AiGeneratedQuestion } from '@/types/ai'
import type { Question } from '@/types/question'

const router = useRouter()
const auth = useAuthStore()
const aiStore = useAiStore()

const step = ref<'config' | 'generating' | 'practice' | 'result'>('config')
const generating = ref(false)
const error = ref('')

const subjectId = computed(() => auth.selectedSubjectId)
const chapterId = ref(0)
const types = ref<string[]>(['single', 'multi'])
const difficulty = ref('medium')
const count = ref(5)

const chapters = ref<{ id: number; name: string }[]>([])
const aiQuestions = ref<AiGeneratedQuestion[]>([])
const answers = ref<Record<number, string>>({})
const submittedIds = ref<Set<number>>(new Set())

const typeOptions = [
  { value: 'single', label: '单选题' },
  { value: 'multi', label: '多选题' },
]
const difficultyOptions = [
  { value: 'easy', label: '简单' },
  { value: 'medium', label: '中等' },
  { value: 'hard', label: '困难' },
]
const countOptions = [3, 5, 10, 15, 20]

const practiceQuestions = computed<Question[]>(() => {
  return aiQuestions.value.map(q => ({
    id: q.id,
    subject_id: subjectId.value,
    sub_subject_id: 0,
    chapter_id: chapterId.value,
    type: q.type as Question['type'],
    difficulty: q.difficulty as Question['difficulty'],
    content: q.content,
    options: q.options,
    answer: q.answer,
    analysis: q.analysis,
    year: 0,
    source: q.source || 'ai',
  }))
})

const correctCount = computed(() => {
  return aiQuestions.value.filter(q => {
    const ans = answers.value[q.id] || ''
    if (q.type === 'multi') {
      const userSet = ans.split(',').map(s => s.trim()).sort()
      const correctSet = (q.answer || '').split(',').map(s => s.trim()).sort()
      return userSet.join(',') === correctSet.join(',')
    }
    return ans.trim().toUpperCase() === (q.answer || '').trim().toUpperCase()
  }).length
})

const totalCount = computed(() => aiQuestions.value.length)
const accuracy = computed(() => totalCount.value ? Math.round(correctCount.value / totalCount.value * 100) : 0)

const typeLabel = (t: string) => t === 'single' ? '单选题' : '多选题'

onMounted(async () => {
  if (!auth.initialized) await auth.checkAuth()
  if (subjectId.value) await loadChapters()
})

async function loadChapters() {
  try {
    const subSubjects = await getSubSubjects(subjectId.value)
    if (subSubjects.length > 0) {
      const chList = await getChapters(subSubjects[0].id)
      chapters.value = [{ id: 0, name: '不限定章节' }, ...chList]
    } else {
      chapters.value = [{ id: 0, name: '不限定章节' }]
    }
  } catch {
    chapters.value = [{ id: 0, name: '不限定章节' }]
  }
}

function toggleType(type: string) {
  const idx = types.value.indexOf(type)
  if (idx >= 0) types.value.splice(idx, 1)
  else types.value.push(type)
}

async function startGenerate() {
  if (!aiStore.hasConfig) {
    router.push('/ai/config')
    return
  }
  if (!subjectId.value) {
    error.value = '请选择科目'
    return
  }
  if (types.value.length === 0) {
    error.value = '请至少选择一种题型'
    return
  }

  step.value = 'generating'
  generating.value = true
  error.value = ''

  try {
    const res = await generateQuestions({
      api_config: aiStore.getConfig(),
      subject_id: subjectId.value,
      chapter_id: chapterId.value,
      types: types.value,
      difficulty: difficulty.value,
      count: count.value,
    })
    aiQuestions.value = res.questions
    answers.value = {}
    submittedIds.value = new Set()
    step.value = 'practice'
  } catch (e) {
    error.value = (e as Error).message
    step.value = 'config'
  } finally {
    generating.value = false
  }
}

function onPracticeBack() {
  step.value = 'result'
}

function finishPractice() {
  step.value = 'result'
}

function retry() {
  step.value = 'config'
  aiQuestions.value = []
  answers.value = {}
  submittedIds.value = new Set()
}

function isCorrect(q: AiGeneratedQuestion): boolean {
  const ans = answers.value[q.id] || ''
  if (q.type === 'multi') {
    const userSet = ans.split(',').map(s => s.trim()).sort()
    const correctSet = (q.answer || '').split(',').map(s => s.trim()).sort()
    return userSet.join(',') === correctSet.join(',')
  }
  return ans.trim().toUpperCase() === (q.answer || '').trim().toUpperCase()
}
</script>

<template>
  <div>
    <!-- 头部 -->
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">AI 智能出题</h1>
      <p class="mt-1 text-sm text-gray-500">由 AI 根据你的学习需求随机生成选择题，答案实时判分</p>
    </div>

    <!-- 配置面板 -->
    <div v-if="step === 'config'" class="mx-auto max-w-2xl">
      <div v-if="!aiStore.hasConfig" class="mb-4 rounded-lg border border-yellow-200 bg-yellow-50 p-4">
        <div class="flex items-center gap-3">
          <svg class="h-5 w-5 text-yellow-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
          </svg>
          <div>
            <p class="text-sm font-medium text-yellow-800">尚未配置 AI API</p>
            <p class="text-xs text-yellow-600">请先前往 AI 配置页面设置你的 API Key</p>
          </div>
          <button
            class="ml-auto cursor-pointer rounded-lg bg-yellow-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-yellow-600"
            @click="router.push('/ai/config')"
          >
            去配置 →
          </button>
        </div>
      </div>

      <div class="rounded-lg bg-white p-6 shadow-sm">
        <div v-if="error" class="mb-4 rounded-lg bg-red-50 px-4 py-2 text-sm text-red-600">{{ error }}</div>

        <div class="space-y-4">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">知识点范围</label>
            <select
              v-model="chapterId"
              class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            >
              <option v-for="ch in chapters" :key="ch.id" :value="ch.id">{{ ch.name }}</option>
            </select>
          </div>

          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">题目类型</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="opt in typeOptions"
                :key="opt.value"
                class="cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors"
                :class="types.includes(opt.value)
                  ? 'border-indigo-500 bg-indigo-50 text-indigo-600'
                  : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
                @click="toggleType(opt.value)"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>

          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">难度</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="opt in difficultyOptions"
                :key="opt.value"
                class="cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors"
                :class="difficulty === opt.value
                  ? 'border-indigo-500 bg-indigo-50 text-indigo-600'
                  : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
                @click="difficulty = opt.value"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>

          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">题目数量</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="n in countOptions"
                :key="n"
                class="cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors"
                :class="count === n
                  ? 'border-indigo-500 bg-indigo-50 text-indigo-600'
                  : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
                @click="count = n"
              >
                {{ n }} 题
              </button>
            </div>
          </div>
        </div>

        <div class="mt-6 flex justify-between">
          <button
            class="cursor-pointer rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm text-gray-600 hover:bg-gray-50"
            @click="router.push('/ai/config')"
          >
            配置 AI
          </button>
          <button
            class="cursor-pointer rounded-lg bg-indigo-600 px-5 py-2 text-sm text-white hover:bg-indigo-700 disabled:opacity-50"
            :disabled="!aiStore.hasConfig"
            @click="startGenerate"
          >
            AI 生成题目
          </button>
        </div>
      </div>
    </div>

    <!-- 生成中 -->
    <div v-else-if="step === 'generating'" class="mx-auto max-w-2xl">
      <div class="rounded-lg bg-white p-12 shadow-sm text-center">
        <div class="mx-auto mb-4 h-12 w-12 animate-spin rounded-full border-4 border-indigo-200 border-t-indigo-600" />
        <p class="text-sm text-gray-600">AI 正在生成题目，请稍候...</p>
        <p class="mt-1 text-xs text-gray-400">这可能需要 10-30 秒</p>
      </div>
    </div>

    <!-- 答题中 -->
    <div v-else-if="step === 'practice'">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-semibold text-gray-800">AI 出题练习</h2>
        <button
          class="cursor-pointer rounded-lg bg-emerald-600 px-4 py-2 text-sm text-white hover:bg-emerald-700"
          @click="finishPractice"
        >
          完成答题
        </button>
      </div>
      <PracticeRunner
        :questions="practiceQuestions"
        mode="ai"
        :submit-handler="async (qid, ans) => {
          answers[qid] = ans
          submittedIds.add(qid)
          return { id: 0, question_id: qid, mode: 'ai', answer: ans, is_correct: isCorrect(aiQuestions.find(q => q.id === qid)!)? 1 : 0, duration: 0 }
        }"
        @back="onPracticeBack"
      />
    </div>

    <!-- 结果页 -->
    <div v-else-if="step === 'result'" class="mx-auto max-w-3xl">
      <div class="rounded-lg bg-white p-6 shadow-sm">
        <h2 class="mb-4 text-lg font-semibold text-gray-900">答题结果</h2>
        <div class="mb-6 grid grid-cols-3 gap-4">
          <div class="rounded-lg bg-indigo-50 p-4 text-center">
            <p class="text-2xl font-bold text-indigo-600">{{ totalCount }}</p>
            <p class="text-xs text-gray-500">总题数</p>
          </div>
          <div class="rounded-lg bg-emerald-50 p-4 text-center">
            <p class="text-2xl font-bold text-emerald-600">{{ correctCount }}</p>
            <p class="text-xs text-gray-500">正确</p>
          </div>
          <div class="rounded-lg p-4 text-center" :class="accuracy >= 60 ? 'bg-emerald-50' : 'bg-red-50'">
            <p class="text-2xl font-bold" :class="accuracy >= 60 ? 'text-emerald-600' : 'text-red-600'">{{ accuracy }}%</p>
            <p class="text-xs text-gray-500">正确率</p>
          </div>
        </div>

        <div class="space-y-3">
          <div
            v-for="(q, idx) in aiQuestions"
            :key="q.id"
            class="rounded-lg border p-4"
            :class="isCorrect(q) ? 'border-emerald-200 bg-emerald-50/30' : 'border-red-200 bg-red-50/30'"
          >
            <div class="flex items-start gap-2">
              <span class="mt-0.5 shrink-0 text-sm font-bold" :class="isCorrect(q) ? 'text-emerald-600' : 'text-red-600'">
                {{ idx + 1 }}.
              </span>
              <div class="flex-1 min-w-0">
                <p class="text-sm text-gray-800" v-html="q.content" />
                <div class="mt-2 flex flex-wrap gap-1.5">
                  <span class="rounded bg-white px-1.5 py-0.5 text-xs text-gray-500 border">{{ typeLabel(q.type) }}</span>
                  <span v-if="q.knowledge_point" class="rounded bg-white px-1.5 py-0.5 text-xs text-gray-500 border">{{ q.knowledge_point }}</span>
                </div>
                <p class="mt-2 text-xs" :class="isCorrect(q) ? 'text-emerald-600' : 'text-red-600'">
                  {{ isCorrect(q) ? '✓ 正确' : '✗ 错误' }}
                  <span v-if="q.answer">&nbsp;正确答案：{{ q.answer }}</span>
                </p>
                <p v-if="q.analysis" class="mt-1.5 text-xs text-gray-500 leading-relaxed">{{ q.analysis }}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-2">
          <button
            class="cursor-pointer rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm text-gray-600 hover:bg-gray-50"
            @click="retry"
          >
            重新生成
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
