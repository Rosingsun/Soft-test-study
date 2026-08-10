<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAiStore } from '@/stores/ai'
import { submitGenerateAsync, getGenerateTask } from '@/api/ai'
import { getChapters, getSubSubjects } from '@/api/subject'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import type { AiGeneratedQuestion, AsyncGenerateTask } from '@/types/ai'
import type { Question } from '@/types/question'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const aiStore = useAiStore()

const step = ref<'config' | 'pending' | 'practice' | 'result'>('config')
const error = ref('')

const subjectId = computed(() => auth.selectedSubjectId)
const chapterId = ref(0)
// 题型改为单选互斥（每次只能出 1 种题型，避免不同模板混用）
const questionType = ref<string>('single')
const difficulty = ref('medium')
const count = ref(5)

const chapters = ref<{ id: number; name: string }[]>([])
const aiQuestions = ref<AiGeneratedQuestion[]>([])
const answers = ref<Record<number, string>>({})
// 案例题专用：按子问题拆分作答内容，answersCS[q.id] = { subIndex: text }
const answersCS = ref<Record<number, Record<number, string>>>({})
const submittedIds = ref<Set<number>>(new Set())

// 异步任务状态
const currentTask = ref<AsyncGenerateTask | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

const typeOptions = [
  { value: 'single', label: '单选题', desc: '4 选 1' },
  { value: 'multi', label: '多选题', desc: '多选 2~4' },
  { value: 'judge', label: '判断题', desc: '正确 / 错误' },
  { value: 'case_study', label: '案例分析', desc: '背景 + 子问题' },
  { value: 'essay', label: '论文题', desc: '仅出题' },
]
const difficultyOptions = [
  { value: 'easy', label: '简单' },
  { value: 'medium', label: '中等' },
  { value: 'hard', label: '困难' },
]
// 各题型的可选数量：案例/论文 token 占用大，限制小一些
const countOptions = computed(() => {
  if (questionType.value === 'case_study' || questionType.value === 'essay') {
    return [1, 2, 3]
  }
  return [3, 5, 10, 15, 20]
})

// 案例/论文不需要 PracticeRunner 组件；只在其他题型时使用
// essay 直接进入结果页（线下写作后再去 AI 评分），case_study 进入专门的 caseStudy 步骤作答
const usePracticeRunner = computed(() =>
  questionType.value !== 'essay' && questionType.value !== 'case_study',
)
const useCaseStudyRunner = computed(() => questionType.value === 'case_study')

const practiceQuestions = computed<Question[]>(() => {
  return aiQuestions.value.map(q => ({
    id: q.id,
    subject_id: subjectId.value,
    sub_subject_id: 0,
    chapter_id: chapterId.value,
    type: q.type as Question['type'],
    difficulty: q.difficulty as Question['difficulty'],
    content: q.content,
    case_material: q.case_material,
    options: q.options,
    answer: q.answer,
    analysis: q.analysis,
    year: 0,
    source: q.source || 'ai',
  }))
})

const correctCount = computed(() => {
  return aiQuestions.value.filter(q => isCorrect(q)).length
})

const totalCount = computed(() => aiQuestions.value.length)
const accuracy = computed(() => totalCount.value ? Math.round(correctCount.value / totalCount.value * 100) : 0)

function typeLabel(t: string) {
  switch (t) {
    case 'single': return '单选题'
    case 'multi': return '多选题'
    case 'judge': return '判断题'
    case 'case_study': return '案例分析'
    case 'essay': return '论文题'
    default: return t
  }
}

onMounted(async () => {
  if (!auth.initialized) await auth.checkAuth()
  if (subjectId.value) await loadChapters()
  // 如果 URL 带 task_id（如通知跳转过来），自动恢复任务状态
  const taskId = route.query.task_id as string | undefined
  if (taskId) {
    await loadTaskById(taskId)
  }
})

onUnmounted(() => {
  stopPolling()
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

function selectType(t: string) {
  questionType.value = t
  // 案例/论文题型较大，重置为较少的默认数量
  if (t === 'case_study' || t === 'essay') {
    if (count.value > 3) count.value = 1
  } else if (count.value === 1 || count.value === 2) {
    count.value = 5
  }
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
  if (!questionType.value) {
    error.value = '请选择题型'
    return
  }

  error.value = ''
  currentTask.value = null

  try {
    // 异步提交，立即返回 task_id
    const submitResp = await submitGenerateAsync({
      api_config: aiStore.getConfig(),
      subject_id: subjectId.value,
      chapter_id: chapterId.value,
      types: [questionType.value],
      difficulty: difficulty.value,
      count: count.value,
    })
    // 跳到 pending 步骤，开始轮询
    step.value = 'pending'
    startPolling(submitResp.task_id)
  } catch (e) {
    error.value = (e as Error).message
  }
}

// 启动任务状态轮询
function startPolling(taskId: string) {
  stopPolling()
  fetchTaskOnce(taskId) // 立即拉一次
  pollTimer = setInterval(() => fetchTaskOnce(taskId), 3000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function fetchTaskOnce(taskId: string) {
  try {
    const t = await getGenerateTask(taskId)
    currentTask.value = t
    if (t.status === 'success') {
      stopPolling()
      applyTaskResult(t)
    } else if (t.status === 'failed') {
      stopPolling()
      error.value = 'AI 出题失败：' + (t.error || '未知错误')
      step.value = 'config'
    }
    // pending / running 继续轮询
  } catch (e) {
    // 任务被清理（重启后丢失）→ 停止轮询
    stopPolling()
    error.value = '任务不存在或已过期，请重新生成'
    step.value = 'config'
  }
}

// 从 URL 携带的 task_id 恢复
async function loadTaskById(taskId: string) {
  try {
    const t = await getGenerateTask(taskId)
    currentTask.value = t
    if (t.status === 'success') {
      applyTaskResult(t)
    } else if (t.status === 'failed') {
      error.value = 'AI 出题失败：' + (t.error || '未知错误')
    } else {
      // pending / running：进入 pending 步骤继续轮询
      step.value = 'pending'
      startPolling(taskId)
    }
    // 同步 URL，保留 task_id 以便刷新恢复
    router.replace({ path: route.path, query: { task_id: taskId } })
  } catch (e) {
    // 任务不存在（重启后丢失），忽略，正常显示配置面板
    console.warn('[AIPractice] load task failed', e)
  }
}

// 把成功结果应用到本地状态，跳到对应步骤
function applyTaskResult(t: AsyncGenerateTask) {
  if (!t.result || !t.result.questions) {
    error.value = '任务结果为空'
    step.value = 'config'
    return
  }
  aiQuestions.value = t.result.questions
  answers.value = {}
  answersCS.value = {}
  submittedIds.value = new Set()
  if (usePracticeRunner.value) {
    step.value = 'practice'
  } else if (useCaseStudyRunner.value) {
    step.value = 'practice'
  } else {
    // 论文：直接到结果页（线下写作）
    step.value = 'result'
  }
}

function onPracticeBack() {
  step.value = 'result'
}

function finishPractice() {
  step.value = 'result'
}

function retry() {
  stopPolling()
  currentTask.value = null
  aiQuestions.value = []
  answers.value = {}
  answersCS.value = {}
  submittedIds.value = new Set()
  step.value = 'config'
  // 清掉 URL 上的 task_id
  router.replace({ path: route.path, query: {} })
}

function isCorrect(q: AiGeneratedQuestion): boolean {
  const ans = answers.value[q.id] || ''
  if (q.type === 'multi') {
    const userSet = ans.split(',').map(s => s.trim()).sort()
    const correctSet = (q.answer || '').split(',').map(s => s.trim()).sort()
    return userSet.join(',') === correctSet.join(',')
  }
  if (q.type === 'judge') {
    return ans.trim() === (q.answer || '').trim()
  }
  return ans.trim().toUpperCase() === (q.answer || '').trim().toUpperCase()
}

// 解析 HTML 换行符
function formatText(text: string): string {
  return (text || '').replace(/\n/g, '<br>')
}

// 解析案例题子问题：按 (1) (2) (3) 或 1. 2. 3. 切分
// 返回 [{ index: 1, text: '子问题内容' }, ...]
function parseCaseStudySubs(content: string): { index: number; text: string }[] {
  if (!content) return []
  // 去除「案例背景见 case_material 字段。」等提示行
  const cleaned = content
    .split('\n')
    .map(l => l.trim())
    .filter(l => l && !/^案例背景见\s*case_material/.test(l))
    .join('\n')

  // 优先匹配 "(1)" / "（1）" / "1." / "1、" 这类前缀
  const re = /[（(](\d+)[）)]|^\s*(\d+)\s*[\.、]/gm
  const matches: { index: number; pos: number }[] = []
  let m: RegExpExecArray | null
  while ((m = re.exec(cleaned)) !== null) {
    const idx = parseInt(m[1] || m[2], 10)
    matches.push({ index: idx, pos: m.index })
  }
  if (matches.length === 0) {
    // 没匹配到子问题：当一道大题对待
    return [{ index: 1, text: cleaned.trim() }]
  }
  // 按子问题切分
  const subs: { index: number; text: string }[] = []
  for (let i = 0; i < matches.length; i++) {
    const start = matches[i].pos
    const end = i + 1 < matches.length ? matches[i + 1].pos : cleaned.length
    let slice = cleaned.slice(start, end).trim()
    // 去掉开头的 "(1)" 这类编号，方便用户阅读
    slice = slice.replace(/^[（(]\d+[）)]\s*/, '')
    subs.push({ index: matches[i].index, text: slice })
  }
  return subs
}

function getCSAnswer(qId: number, subIndex: number): string {
  return answersCS.value[qId]?.[subIndex] || ''
}

function setCSAnswer(qId: number, subIndex: number, val: string) {
  if (!answersCS.value[qId]) answersCS.value[qId] = {}
  answersCS.value[qId][subIndex] = val
}

// 案例题参考答案按 (1)/(2)/(3) 拆开，与子问题索引对应
function getCSReferenceAnswer(q: AiGeneratedQuestion, subIndex: number): string {
  const ref = q.answer || ''
  // 用 \n 切分
  const parts = ref.split(/\n|(?=[（(]\d+[）)])/g)
  for (const p of parts) {
    const m = p.match(/^[（(](\d+)[）)]/)
    if (m && parseInt(m[1], 10) === subIndex) {
      return p.replace(/^[（(]\d+[）)]\s*/, '').trim()
    }
  }
  // 没匹配到（AI 没标编号），则按顺序对位
  const subs = parseCaseStudySubs(q.content || '')
  const idx = subs.findIndex(s => s.index === subIndex)
  if (idx >= 0) {
    const fallback = parts.filter(p => p.trim()).map(p => p.replace(/^[（(]\d+[）)]\s*/, '').trim())
    return fallback[idx] || ''
  }
  return ''
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
                :class="questionType === opt.value
                  ? 'border-indigo-500 bg-indigo-50 text-indigo-600'
                  : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
                @click="selectType(opt.value)"
              >
                <span class="font-medium">{{ opt.label }}</span>
                <span class="ml-1 text-xs text-gray-400">{{ opt.desc }}</span>
              </button>
            </div>
            <p class="mt-1 text-xs text-gray-400">单次仅生成一种题型，避免不同模板混用</p>
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

    <!-- 任务已提交：后台异步生成中 -->
    <div v-else-if="step === 'pending'" class="mx-auto max-w-2xl">
      <div class="rounded-lg bg-white p-8 shadow-sm">
        <div class="mb-4 flex items-center gap-3">
          <div class="h-10 w-10 animate-spin rounded-full border-4 border-indigo-200 border-t-indigo-600" />
          <div>
            <p class="text-base font-semibold text-gray-900">AI 正在生成题目</p>
            <p class="text-xs text-gray-500">已提交后台任务，你可以浏览其他页面，完成后会通过消息中心通知你</p>
          </div>
        </div>
        <div class="space-y-2 rounded-lg border border-gray-100 bg-gray-50/60 p-4 text-sm text-gray-600">
          <div class="flex items-center justify-between">
            <span>任务 ID</span>
            <code class="rounded bg-white px-2 py-0.5 font-mono text-xs text-gray-500">{{ currentTask?.id?.slice(0, 16) }}…</code>
          </div>
          <div class="flex items-center justify-between">
            <span>状态</span>
            <span class="flex items-center gap-1.5 text-indigo-600">
              <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-indigo-500"></span>
              {{ currentTask?.status === 'running' ? '生成中' : '排队中' }}
            </span>
          </div>
          <div class="flex items-center justify-between">
            <span>题型</span>
            <span class="text-gray-700">{{ typeLabel(questionType) }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span>数量</span>
            <span class="text-gray-700">{{ count }} 道</span>
          </div>
        </div>
        <div class="mt-4 flex items-center justify-between text-xs text-gray-500">
          <span>每 3 秒自动刷新状态</span>
          <button
            class="cursor-pointer rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-50"
            @click="retry"
          >取消并返回</button>
        </div>
      </div>
    </div>

    <!-- 答题中：客观题（单选/多选/判断）用 PracticeRunner -->
    <div v-else-if="step === 'practice' && usePracticeRunner">
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

    <!-- 答题中：案例分析（每个子问题一个 textarea） -->
    <div v-else-if="step === 'practice' && useCaseStudyRunner" class="mx-auto max-w-4xl">
      <div class="sticky top-4 z-10 mb-4 flex items-center justify-between rounded-xl bg-white/95 p-3 shadow-sm ring-1 ring-gray-100 backdrop-blur">
        <div>
          <h2 class="text-lg font-semibold text-gray-800">案例分析作答</h2>
          <p class="text-xs text-gray-500">阅读案例背景，回答下列子问题；提交后可查看参考答案与解析</p>
        </div>
        <button
          class="cursor-pointer rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-emerald-700"
          @click="finishPractice"
        >
          完成作答
        </button>
      </div>

      <div class="space-y-6">
        <div
          v-for="(q, idx) in aiQuestions"
          :key="q.id"
          class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm"
        >
          <!-- 题号 + 知识点 -->
          <div class="mb-3 flex flex-wrap items-center gap-2">
            <span class="rounded-full bg-indigo-50 px-2.5 py-0.5 text-xs font-semibold text-indigo-600 ring-1 ring-inset ring-indigo-200">第 {{ idx + 1 }} 题</span>
            <span class="rounded-full bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-700 ring-1 ring-inset ring-amber-200">案例分析</span>
            <span v-if="q.knowledge_point" class="rounded bg-gray-50 px-2 py-0.5 text-xs text-gray-500">{{ q.knowledge_point }}</span>
          </div>

          <!-- 案例背景 -->
          <div v-if="q.case_material" class="mb-4 rounded-lg border border-indigo-100 bg-indigo-50/40 p-4">
            <p class="mb-1.5 text-xs font-semibold text-indigo-700">【案例背景】</p>
            <p class="whitespace-pre-line text-sm leading-relaxed text-gray-800">{{ q.case_material }}</p>
          </div>

          <!-- 子问题列表 + 答题区 -->
          <div class="space-y-4">
            <div
              v-for="sub in parseCaseStudySubs(q.content)"
              :key="`${q.id}-${sub.index}`"
              class="rounded-lg border border-gray-100 bg-gray-50/40 p-4"
            >
              <div class="mb-2 flex items-start gap-2">
                <span class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-indigo-600 text-xs font-semibold text-white">{{ sub.index }}</span>
                <p class="flex-1 text-sm leading-relaxed text-gray-800" v-html="formatText(sub.text)" />
              </div>
              <textarea
                :value="getCSAnswer(q.id, sub.index)"
                @input="setCSAnswer(q.id, sub.index, ($event.target as HTMLTextAreaElement).value)"
                :rows="sub.text.length > 80 ? 6 : 4"
                class="mt-2 w-full resize-y rounded-lg border border-gray-200 bg-white p-3 text-sm leading-relaxed text-gray-800 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                :placeholder="`请输入第 ${sub.index} 题的答案...`"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 底部提交条 -->
      <div class="sticky bottom-4 mt-6 flex items-center justify-between rounded-xl bg-white/95 p-3 shadow-lg ring-1 ring-gray-100 backdrop-blur">
        <span class="text-xs text-gray-500">作答完成后点击右侧按钮查看参考答案与解析</span>
        <button
          class="cursor-pointer rounded-lg bg-emerald-600 px-5 py-2 text-sm font-medium text-white transition-colors hover:bg-emerald-700"
          @click="finishPractice"
        >
          完成作答
        </button>
      </div>
    </div>

    <!-- 结果页 -->
    <div v-else-if="step === 'result'" class="mx-auto max-w-3xl">
      <div class="rounded-lg bg-white p-6 shadow-sm">
        <h2 class="mb-4 text-lg font-semibold text-gray-900">
          {{ usePracticeRunner ? '答题结果' : useCaseStudyRunner ? '案例作答结果' : '生成结果' }}
        </h2>

        <!-- 客观题统计卡片：仅在练习模式下显示 -->
        <div v-if="usePracticeRunner" class="mb-6 grid grid-cols-3 gap-4">
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
            :class="usePracticeRunner && !isCorrect(q) ? 'border-red-200 bg-red-50/30' : 'border-gray-200 bg-white'"
          >
            <div class="flex items-start gap-2">
              <span class="mt-0.5 shrink-0 text-sm font-bold text-gray-500">
                {{ idx + 1 }}.
              </span>
              <div class="flex-1 min-w-0">
                <!-- 案例分析：先展示背景材料 -->
                <div v-if="q.case_material" class="mb-3 rounded-lg border border-indigo-100 bg-indigo-50/40 p-3">
                  <p class="mb-1 text-xs font-medium text-indigo-700">【案例背景】</p>
                  <p class="whitespace-pre-line text-sm leading-relaxed text-gray-800">{{ q.case_material }}</p>
                </div>

                <p class="text-sm text-gray-800" v-html="q.content" />

                <div class="mt-2 flex flex-wrap gap-1.5">
                  <span class="rounded bg-white px-1.5 py-0.5 text-xs text-gray-500 border">{{ typeLabel(q.type) }}</span>
                  <span v-if="q.knowledge_point" class="rounded bg-white px-1.5 py-0.5 text-xs text-gray-500 border">{{ q.knowledge_point }}</span>
                </div>

                <!-- 客观题：显示对错 + 答案 -->
                <template v-if="usePracticeRunner">
                  <p class="mt-2 text-xs" :class="isCorrect(q) ? 'text-emerald-600' : 'text-red-600'">
                    {{ isCorrect(q) ? '✓ 正确' : '✗ 错误' }}
                    <span v-if="q.answer">&nbsp;正确答案：{{ q.answer }}</span>
                  </p>
                </template>

                <!-- 案例分析：展示每题用户答案 vs 参考答案 -->
                <template v-else-if="q.type === 'case_study'">
                  <div class="mt-3 space-y-3">
                    <div
                      v-for="sub in parseCaseStudySubs(q.content)"
                      :key="`r-${q.id}-${sub.index}`"
                      class="rounded-lg border border-gray-100 bg-gray-50/40 p-3"
                    >
                      <div class="mb-1.5 flex items-center gap-2">
                        <span class="flex h-5 w-5 items-center justify-center rounded-full bg-indigo-600 text-[10px] font-semibold text-white">{{ sub.index }}</span>
                        <span class="text-xs font-medium text-gray-700">第 {{ sub.index }} 题</span>
                      </div>
                      <div class="grid gap-2 sm:grid-cols-2">
                        <div>
                          <p class="mb-1 text-[11px] font-semibold text-gray-500">你的答案</p>
                          <p class="whitespace-pre-line rounded border border-gray-200 bg-white p-2 text-xs leading-relaxed text-gray-800" :class="!getCSAnswer(q.id, sub.index) ? 'italic text-gray-400' : ''">
                            {{ getCSAnswer(q.id, sub.index) || '（未作答）' }}
                          </p>
                        </div>
                        <div>
                          <p class="mb-1 text-[11px] font-semibold text-amber-700">参考答案</p>
                          <p class="whitespace-pre-line rounded border border-amber-200 bg-amber-50/40 p-2 text-xs leading-relaxed text-gray-800">
                            {{ getCSReferenceAnswer(q, sub.index) || '（无）' }}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                </template>

                <!-- 论文：仅展示 -->
                <template v-else-if="q.type === 'essay'">
                  <p class="mt-2 text-xs text-gray-500">请独立写作后，前往「AI 评分」获取专家点评</p>
                </template>

                <p v-if="q.analysis" class="mt-2 whitespace-pre-line text-xs leading-relaxed text-gray-600">{{ q.analysis }}</p>
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
