<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAiStore } from '@/stores/ai'
import { useSubjectStore } from '@/stores/subject'
import { submitGenerateAsync, getGenerateTask, listGenerateHistory, listInflightTasks, getGenerateBatch } from '@/api/ai'
import { submitPractice } from '@/api/practice'
import { getChapters, getSubSubjects } from '@/api/subject'
import { showToast } from '@/utils/toast'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import ExtractKnowledgeModal from '@/components/knowledge/ExtractKnowledgeModal.vue'
import type { AiGeneratedQuestion, AsyncGenerateTask, AiBatchHistoryItem } from '@/types/ai'
import type { Question } from '@/types/question'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const aiStore = useAiStore()
const subjectStore = useSubjectStore()

const step = ref<'config' | 'pending' | 'practice' | 'result'>('config')
const error = ref('')

const subjectId = computed(() => auth.selectedSubjectId)
const currentSubjectName = computed(() => {
  const id = subjectId.value
  if (!id) return '未选择'
  return subjectStore.subjects.find(s => s.id === id)?.name || `科目#${id}`
})
const chapterId = ref(0)
const questionType = ref<string>('single')
const difficulty = ref('medium')
const count = ref(5)

const chapters = ref<{ id: number; name: string }[]>([])
const aiQuestions = ref<AiGeneratedQuestion[]>([])
const answers = ref<Record<number, string>>({})
const answersCS = ref<Record<number, Record<number, string>>>({})
const submittedIds = ref<Set<number>>(new Set())

// 知识点提取弹窗:对应当前正在提取的 AI 生成题。
// 弹窗复用 components/knowledge/ExtractKnowledgeModal.vue,仅在 knowledge_point 为空时打开。
const showExtractModal = ref(false)
const extractTarget = ref<AiGeneratedQuestion | null>(null)

const currentTask = ref<AsyncGenerateTask | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

const historyList = ref<AiBatchHistoryItem[]>([])
const historyLoading = ref(false)
const historyError = ref('')

const typeOptions = [
  {
    value: 'single', label: '单选题', desc: '4 选 1',
    icon: 'M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM3.75 12h.007v.008H3.75V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm-.375 5.25h.007v.008H3.75v-.008zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z',
    color: 'indigo',
  },
  {
    value: 'multi', label: '多选题', desc: '多选 2~4',
    icon: 'M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    color: 'violet',
  },
  {
    value: 'judge', label: '判断题', desc: '正确 / 错误',
    icon: 'M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z',
    color: 'emerald',
  },
  {
    value: 'case_study', label: '案例分析', desc: '背景 + 子问题',
    icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z',
    color: 'amber',
  },
  {
    value: 'essay', label: '论文题', desc: '仅出题',
    icon: 'M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10',
    color: 'rose',
  },
]

const difficultyOptions = [
  { value: 'easy', label: '简单', desc: '基础概念', color: 'emerald' },
  { value: 'medium', label: '中等', desc: '综合应用', color: 'amber' },
  { value: 'hard', label: '困难', desc: '深入分析', color: 'red' },
]

const countOptions = computed(() => {
  if (questionType.value === 'case_study' || questionType.value === 'essay') {
    return [1, 2, 3]
  }
  return [3, 5, 10, 15, 20]
})

const usePracticeRunner = computed(() => questionType.value !== 'case_study')
const useCaseStudyRunner = computed(() => questionType.value === 'case_study')

const practiceQuestions = computed<Question[]>(() => {
  return aiQuestions.value.map(q => ({
    id: q.id,
    subject_id: subjectId.value,
    sub_subject_id: 0,
    chapter_id: chapterId.value,
    parent_id: 0,
    type: q.type as Question['type'],
    difficulty: q.difficulty as Question['difficulty'],
    content: q.content,
    case_material: q.case_material,
    options: q.options,
    answer: q.answer,
    analysis: q.analysis,
    year: 0,
    source: q.source || 'ai',
    knowledge_point: q.knowledge_point,
  }))
})

const objectiveQuestions = computed(() =>
  aiQuestions.value.filter(q => q.type !== 'essay' && q.type !== 'case_study'),
)
const correctCount = computed(() => objectiveQuestions.value.filter(q => isCorrect(q)).length)
const totalCount = computed(() => objectiveQuestions.value.length)
const accuracy = computed(() => totalCount.value ? Math.round(correctCount.value / totalCount.value * 100) : 0)

const colorClassMap: Record<string, { active: string; inactive: string; icon: string }> = {
  indigo: {
    active: 'border-indigo-500 bg-indigo-50/60 ring-1 ring-indigo-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-indigo-300 hover:bg-indigo-50/30',
    icon: 'bg-indigo-50 text-indigo-600',
  },
  violet: {
    active: 'border-violet-500 bg-violet-50/60 ring-1 ring-violet-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-violet-300 hover:bg-violet-50/30',
    icon: 'bg-violet-50 text-violet-600',
  },
  emerald: {
    active: 'border-emerald-500 bg-emerald-50/60 ring-1 ring-emerald-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-emerald-300 hover:bg-emerald-50/30',
    icon: 'bg-emerald-50 text-emerald-600',
  },
  amber: {
    active: 'border-amber-500 bg-amber-50/60 ring-1 ring-amber-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-amber-300 hover:bg-amber-50/30',
    icon: 'bg-amber-50 text-amber-600',
  },
  rose: {
    active: 'border-rose-500 bg-rose-50/60 ring-1 ring-rose-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-rose-300 hover:bg-rose-50/30',
    icon: 'bg-rose-50 text-rose-600',
  },
  red: {
    active: 'border-red-500 bg-red-50/60 ring-1 ring-red-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-red-300 hover:bg-red-50/30',
    icon: 'bg-red-50 text-red-600',
  },
}

function getTypeColor(t: string) {
  return colorClassMap[t] || colorClassMap.indigo
}

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

function difficultyLabel(d: string) {
  switch (d) {
    case 'easy': return '简单'
    case 'medium': return '中等'
    case 'hard': return '困难'
    default: return d || '中等'
  }
}

onMounted(async () => {
  if (!auth.initialized) await auth.checkAuth()
  if (subjectId.value) {
    try {
      const sub = subjectStore.subjects.find(s => s.id === subjectId.value)
      if (!sub && auth.selectedLevelId) {
        await subjectStore.fetchSubjectsByLevel(auth.selectedLevelId)
      }
    } catch {
      // 静默忽略
    }
    await loadChapters()
  }
  const taskId = route.query.task_id as string | undefined
  if (taskId) {
    await loadTaskById(taskId)
  }
  loadHistory()
})

onUnmounted(() => {
  stopPolling()
})

// 监听 task_id 查询参数变化：用户在 AI 练习页时点通知进入、带 task_id 的链接被复用时，
// 需要重新拉取任务状态并切换到答题/结果页。
watch(
  () => route.query.task_id,
  (newId) => {
    if (typeof newId === 'string' && newId && newId !== currentTask.value?.id) {
      loadTaskById(newId)
    }
  },
)

async function loadHistory() {
  historyLoading.value = true
  historyError.value = ''
  try {
    // 两个端点并行拉：/ai/history（已完成）+ /ai/tasks/inflight（生成中）。
    // 独立端点保证其中一个失败时另一个仍能提供数据，避免「刷新一下就没了」。
    const [historyResp, inflightResp] = await Promise.allSettled([
      listGenerateHistory(),
      listInflightTasks(),
    ])

    const completed: AiBatchHistoryItem[] =
      historyResp.status === 'fulfilled' ? (historyResp.value.list || []) : []
    const inflight: AiBatchHistoryItem[] =
      inflightResp.status === 'fulfilled' ? (inflightResp.value.list || []) : []

    if (historyResp.status === 'rejected' && inflightResp.status === 'rejected') {
      throw historyResp.reason || inflightResp.reason
    }

    // 合并：去重 by batch_id（inflight 优先，因为状态最新）；
    // 同时把 completed 里已变成 success 的 inflight 任务过滤掉（理论上不会重叠）。
    const inflightByID = new Map<string, AiBatchHistoryItem>()
    for (const it of inflight) {
      if (it.status === 'pending' || it.status === 'running') {
        inflightByID.set(it.batch_id, it)
      }
    }
    const completedFiltered = completed.filter(c => !inflightByID.has(c.batch_id))
    const merged = [...inflight, ...completedFiltered]
    // 按 created_at 倒序（infllight 通常是最新，排在前；字符串比较即可，格式统一）
    merged.sort((a, b) => (b.created_at > a.created_at ? 1 : b.created_at < a.created_at ? -1 : 0))
    historyList.value = merged
  } catch (e) {
    historyError.value = (e as Error).message || '加载历史记录失败'
  } finally {
    historyLoading.value = false
  }
}

async function openHistory(batch: AiBatchHistoryItem) {
  if (historyLoading.value) return
  if (batch.status === 'pending' || batch.status === 'running') {
    showToast('题目未完全生成，请稍后再试', 'info')
    return
  }
  try {
    const resp = await getGenerateBatch(batch.batch_id)
    if (!resp || !resp.questions || !resp.questions.length) {
      error.value = '该批题目不存在，可能已被清理'
      return
    }
    questionType.value = batch.type as 'single' | 'multi' | 'judge' | 'case_study' | 'essay'
    applyTaskResult({
      id: batch.batch_id,
      status: 'success',
      question_type: batch.type,
      chapter_id: batch.chapter_id,
      chapter_name: batch.chapter_name,
      difficulty: batch.difficulty,
      count: batch.count,
      created_at: batch.created_at,
      updated_at: batch.created_at,
      finished_at: batch.created_at,
      result: resp,
    })
  } catch (e) {
    error.value = (e as Error).message || '加载该批题目失败'
  }
}

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
  if (t === 'case_study' || t === 'essay') {
    if (count.value > 3) count.value = 1
  } else if (count.value === 1 || count.value === 2) {
    count.value = 5
  }
}

// 当前选中的章节名（用于在历史列表中渲染占位项）
function currentChapterName(): string {
  const ch = chapters.value.find(c => c.id === chapterId.value)
  return ch?.name || '不限定'
}

// patchHistoryItem 按 batch_id 在 historyList 中查找并就地更新；找到则改 status/failed
// 找不到则根据传入字段在列表前部插入占位（用于刚提交任务，loadHistory 还没回来时）。
function patchHistoryItem(item: Partial<AiBatchHistoryItem> & { batch_id: string }) {
  const idx = historyList.value.findIndex(h => h.batch_id === item.batch_id)
  if (idx >= 0) {
    historyList.value[idx] = { ...historyList.value[idx], ...item }
  } else {
    const placeholder: AiBatchHistoryItem = {
      subject_id: subjectId.value || 0,
      subject_name: currentSubjectName.value,
      chapter_id: chapterId.value,
      chapter_name: currentChapterName(),
      type: questionType.value,
      type_label: typeLabel(questionType.value),
      difficulty: difficulty.value,
      count: 0,
      created_at: new Date().toISOString().replace('T', ' ').slice(0, 16),
      answered_count: 0,
      correct_count: 0,
      knowledge_points: [],
      ...item,
    }
    historyList.value = [placeholder, ...historyList.value]
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
    const submitResp = await submitGenerateAsync({
      api_config: aiStore.getConfig(),
      subject_id: subjectId.value,
      chapter_id: chapterId.value,
      types: [questionType.value],
      difficulty: difficulty.value,
      count: count.value,
    })
    // 立即在历史列表前部插入「排队中」占位，避免等待第一次 loadHistory 才出现
    patchHistoryItem({ batch_id: submitResp.task_id, status: 'pending' })
    step.value = 'pending'
    startPolling(submitResp.task_id)
  } catch (e) {
    error.value = formatSubmitError((e as Error).message)
  }
}

function formatSubmitError(msg: string): string {
  if (!msg) return '提交失败，请稍后重试'
  if (msg.includes('429') || msg.includes('rate')) return '请求过于频繁，请稍候 1 分钟再试'
  if (msg.includes('timeout') || msg.includes('Timeout')) return 'AI 响应超时，请稍后重试'
  if (msg.includes('API Key') || msg.includes('api_key')) return 'API Key 无效，请前往「AI 配置」检查'
  if (msg.includes('401') || msg.includes('403')) return 'API 鉴权失败，请检查 API Key 是否正确'
  return msg
}

function startPolling(taskId: string) {
  stopPolling()
  fetchTaskOnce(taskId)
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
    // 实时同步状态到历史列表（不依赖 loadHistory 全量刷新）
    if (t.status === 'pending' || t.status === 'running') {
      patchHistoryItem({ batch_id: t.id, status: t.status })
    }
    if (t.status === 'success') {
      stopPolling()
      applyTaskResult(t)
    } else if (t.status === 'failed') {
      stopPolling()
      // 失败任务按计划不留在历史中（后端 ListGenerateHistory 不返回 failed），
      // 本地立刻移除占位；服务端会通过消息中心通知。
      historyList.value = historyList.value.filter(h => h.batch_id !== t.id)
      error.value = formatTaskError(t.error || '未知错误')
      step.value = 'config'
    }
  } catch (e) {
    stopPolling()
    error.value = '任务不存在或已过期，请重新生成'
    step.value = 'config'
  }
}

function formatTaskError(raw: string): string {
  if (!raw) return 'AI 出题失败，请稍后重试'
  if (raw.includes('多次重试') || raw.includes('解析失败')) {
    return 'AI 多次输出格式异常，已自动重试失败。建议：减少题量后重试，或换一个更稳定的模型'
  }
  if (raw.includes('timeout') || raw.includes('超时')) {
    return 'AI 响应超时，请稍后重试或减少题量'
  }
  if (raw.includes('API') && (raw.includes('Key') || raw.includes('key'))) {
    return 'API Key 鉴权失败，请前往「AI 配置」检查'
  }
  return 'AI 出题失败：' + raw + '。建议：减少题量、更换模型或稍后重试'
}

async function loadTaskById(taskId: string) {
  try {
    const t = await getGenerateTask(taskId)
    currentTask.value = t
    if (t.status === 'pending' || t.status === 'running') {
      patchHistoryItem({ batch_id: t.id, status: t.status })
    }
    if (t.status === 'success') {
      applyTaskResult(t)
    } else if (t.status === 'failed') {
      historyList.value = historyList.value.filter(h => h.batch_id !== t.id)
      error.value = formatTaskError(t.error || '未知错误')
    } else {
      step.value = 'pending'
      startPolling(taskId)
    }
    router.replace({ path: route.path, query: { task_id: taskId } })
  } catch (e) {
    console.warn('[AIPractice] load task failed', e)
  }
}

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
  // 任务成功：先把历史列表中的占位标为 success（保证徽章立即消失），
  // 然后再异步拉一次完整列表补全 subject_name / count 等真实字段
  patchHistoryItem({ batch_id: t.id, status: 'success' })
  loadHistory()
  if (usePracticeRunner.value || useCaseStudyRunner.value) {
    step.value = 'practice'
  } else {
    step.value = 'result'
  }
}

function onPracticeBack() {
  step.value = 'result'
  loadHistory()
}

async function finishPractice() {
  if (useCaseStudyRunner.value) {
    const tasks = aiQuestions.value
      .filter(q => q.type === 'case_study')
      .map(q => {
        const parts = parseCaseStudySubs(q.content)
        const joined = parts
          .map(sub => getCSAnswer(q.id, sub.index).trim())
          .filter(Boolean)
          .join('\n\n')
        if (!joined) return Promise.resolve()
        return submitPractice({
          question_id: q.id,
          mode: 'ai',
          answer: joined,
          duration: 0,
        }).catch(() => undefined)
      })
    await Promise.allSettled(tasks)
  }
  step.value = 'result'
  loadHistory()
}

function retry() {
  stopPolling()
  currentTask.value = null
  aiQuestions.value = []
  answers.value = {}
  answersCS.value = {}
  submittedIds.value = new Set()
  step.value = 'config'
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

function formatText(text: string): string {
  return (text || '').replace(/\n/g, '<br>')
}

// 去掉 HTML 标签,供 ExtractKnowledgeModal question_content 使用
function stripHtml(html: string): string {
  return (html || '').replace(/<[^>]*>/g, '')
}

// 打开 AI 提取知识点弹窗。弹窗内容来源于 aiQuestions(q 是 AiGeneratedQuestion),
// 复用 components/knowledge/ExtractKnowledgeModal.vue 已有的提取+加入知识点库能力。
// AI 未配置时给出 toast 提示并跳转配置页,避免在弹窗内再次失败。
function openExtractModal(q: AiGeneratedQuestion) {
  if (!aiStore.hasConfig) {
    showToast('尚未配置 AI API，请先在「AI 配置」中设置', 'error')
    router.push('/ai/config')
    return
  }
  extractTarget.value = q
  showExtractModal.value = true
}

// 监听 ExtractKnowledgeModal 的 added 事件:
// 用户把 AI 提取的知识点加入自己的知识点库成功后,把考点名称回填到当前题目的 knowledge_point 字段。
// 这样原卡片上的"AI获取考点知识"按钮会立即消失、徽章会显示,避免重复提取。
function onKnowledgeAdded(items: { name: string; status: string }[]) {
  if (!extractTarget.value || !items.length) return
  const targetId = extractTarget.value.id
  const addedNames = items.map(it => it.name).filter(Boolean)
  if (!addedNames.length) return

  const idx = aiQuestions.value.findIndex(q => q.id === targetId)
  if (idx < 0) return

  const existing = aiQuestions.value[idx].knowledge_point
  const merged = existing
    ? [existing, ...addedNames].join('、')
    : addedNames.join('、')
  aiQuestions.value[idx] = {
    ...aiQuestions.value[idx],
    knowledge_point: merged,
  }
  // 同步更新 extractTarget,保持弹窗关联题目的引用与列表一致
  extractTarget.value = aiQuestions.value[idx]
}

function parseCaseStudySubs(content: string): { index: number; text: string }[] {
  if (!content) return []
  const cleaned = content
    .split('\n')
    .map(l => l.trim())
    .filter(l => l && !/^案例背景见\s*case_material/.test(l))
    .join('\n')

  const re = /[（(](\d+)[）)]|^\s*(\d+)\s*[\.、]/gm
  const matches: { index: number; pos: number }[] = []
  let m: RegExpExecArray | null
  while ((m = re.exec(cleaned)) !== null) {
    const idx = parseInt(m[1] || m[2], 10)
    matches.push({ index: idx, pos: m.index })
  }
  if (matches.length === 0) {
    return [{ index: 1, text: cleaned.trim() }]
  }
  const subs: { index: number; text: string }[] = []
  for (let i = 0; i < matches.length; i++) {
    const start = matches[i].pos
    const end = i + 1 < matches.length ? matches[i + 1].pos : cleaned.length
    let slice = cleaned.slice(start, end).trim()
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

function getCSReferenceAnswer(q: AiGeneratedQuestion, subIndex: number): string {
  const ref = q.answer || ''
  const parts = ref.split(/\n|(?=[（(]\d+[）)])/g)
  for (const p of parts) {
    const m = p.match(/^[（(](\d+)[）)]/)
    if (m && parseInt(m[1], 10) === subIndex) {
      return p.replace(/^[（(]\d+[）)]\s*/, '').trim()
    }
  }
  const subs = parseCaseStudySubs(q.content || '')
  const idx = subs.findIndex(s => s.index === subIndex)
  if (idx >= 0) {
    const fallback = parts.filter(p => p.trim()).map(p => p.replace(/^[（(]\d+[）)]\s*/, '').trim())
    return fallback[idx] || ''
  }
  return ''
}

function difficultyBadgeType(d: string): 'success' | 'warning' | 'danger' | 'default' {
  if (d === 'easy') return 'success'
  if (d === 'medium') return 'warning'
  if (d === 'hard') return 'danger'
  return 'default'
}
</script>

<template>
  <div>
    <BasePageHeader title="AI 智能出题" subtitle="由 AI 根据你的学习需求生成题目，客观题实时判分、主观题智能点评">
      <template #actions>
        <BaseButton type="secondary" @click="router.push('/ai/config')">
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z" />
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          AI 配置
        </BaseButton>
      </template>
    </BasePageHeader>

    <div class="flex flex-col gap-6 lg:flex-row lg:items-start">
      <!-- 左侧：历史生成记录 -->
      <aside class="shrink-0 lg:w-60 xl:w-64">
        <BaseCard>
          <div class="mb-4 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <div class="flex h-7 w-7 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h2 class="text-sm font-semibold text-gray-800">历史记录</h2>
              <span v-if="historyList.length" class="rounded-full bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500">{{ historyList.length }}</span>
            </div>
            <button
              class="cursor-pointer rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 disabled:opacity-50"
              title="刷新历史"
              :disabled="historyLoading"
              @click="loadHistory"
            >
              <svg class="h-3.5 w-3.5" :class="historyLoading ? 'animate-spin' : ''" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
              </svg>
            </button>
          </div>

          <div class="max-h-[calc(100vh-200px)] overflow-y-auto pr-1">
            <div v-if="historyLoading">
              <BaseSkeleton variant="list" :count="3" />
            </div>
            <BaseEmpty
              v-else-if="!historyList.length"
              title="暂无生成记录"
              description="使用「AI 生成题目」后，这里会保留每次生成的批次"
            />
            <div v-else-if="historyError" class="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600">{{ historyError }}</div>
            <div v-else class="space-y-2">
              <button
                v-for="item in historyList"
                :key="item.batch_id"
                class="group w-full cursor-pointer rounded-xl border border-gray-100 bg-white p-3 text-left transition-all duration-200 hover:border-indigo-200 hover:bg-indigo-50/40 hover:shadow-sm"
                :class="(item.status === 'pending' || item.status === 'running') ? 'cursor-not-allowed border-indigo-100 bg-indigo-50/30' : ''"
                :title="(item.status === 'pending' || item.status === 'running') ? '题目未完全生成，请稍后再试' : ''"
                @click="openHistory(item)"
              >
                <div class="mb-1.5 flex items-center justify-between gap-2">
                  <span class="flex items-center gap-1.5">
                    <span class="rounded-full bg-indigo-50 px-2 py-0.5 text-[11px] font-semibold text-indigo-600 ring-1 ring-inset ring-indigo-200">
                      {{ item.type_label }}
                    </span>
                    <span
                      v-if="item.status === 'pending'"
                      class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-medium text-gray-600 ring-1 ring-inset ring-gray-200"
                    >
                      <span class="h-1.5 w-1.5 rounded-full bg-gray-400"></span>
                      排队中
                    </span>
                    <span
                      v-else-if="item.status === 'running'"
                      class="inline-flex items-center gap-1 rounded-full bg-indigo-50 px-2 py-0.5 text-[11px] font-medium text-indigo-600 ring-1 ring-inset ring-indigo-200"
                    >
                      <svg class="h-2.5 w-2.5 animate-spin" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                      </svg>
                      生成中
                    </span>
                  </span>
                  <span class="text-[11px] font-medium text-gray-500">{{ item.count || '—' }} 题</span>
                </div>
                <p class="truncate text-xs font-medium text-gray-800 group-hover:text-indigo-700">{{ item.subject_name || '未知科目' }}</p>
                <p class="mt-0.5 truncate text-[11px] text-gray-500">
                  {{ item.chapter_name || '全部章节' }}
                </p>
                <div v-if="(item.status === 'pending' || item.status === 'running')" class="mt-1.5 flex items-center gap-1.5 text-[11px]">
                  <BaseBadge :type="difficultyBadgeType(item.difficulty)">{{ difficultyLabel(item.difficulty) }}</BaseBadge>
                  <span class="text-gray-400">生成完成后可查看</span>
                </div>
                <template v-else>
                  <div class="mt-1.5 flex items-center gap-1.5 text-[11px]">
                    <BaseBadge :type="difficultyBadgeType(item.difficulty)">{{ difficultyLabel(item.difficulty) }}</BaseBadge>
                    <span class="font-medium" :class="item.answered_count >= item.count ? 'text-emerald-600' : 'text-indigo-600'">
                      {{ item.answered_count }}/{{ item.count }}
                    </span>
                  </div>
                  <p class="mt-1 text-[11px] text-gray-400">{{ item.created_at }}</p>
                </template>
              </button>
            </div>
          </div>
        </BaseCard>
      </aside>

      <!-- 右侧：主流程 -->
      <div class="min-w-0 flex-1">
        <!-- ============== 配置面板 ============== -->
        <div v-if="step === 'config'">
          <!-- 科目提示 -->
          <div
            v-if="subjectId"
            class="mb-5 flex items-center gap-3 rounded-xl border border-indigo-100 bg-gradient-to-r from-indigo-50/80 to-violet-50/60 px-4 py-3"
          >
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-white text-indigo-600 shadow-sm ring-1 ring-indigo-100">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4.26 10.147a60.438 60.438 0 0 0-.491 6.347A48.62 48.62 0 0 1 12 20.904a48.62 48.62 0 0 1 8.232-4.41 60.46 60.46 0 0 0-.491-6.347m-15.482 0a50.636 50.636 0 0 0-2.658-.813A59.906 59.906 0 0 1 12 3.493a59.903 59.903 0 0 1 10.399 5.84c-.896.248-1.783.52-2.658.814m-15.482 0A50.717 50.717 0 0 1 12 13.489a50.702 50.702 0 0 1 7.74-3.342M6.75 15a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm0 0v-3.675A55.378 55.378 0 0 1 12 8.443m-7.007 11.55A5.981 5.981 0 0 0 6.75 15.75v-1.5" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-xs text-indigo-600/80">当前科目</p>
              <p class="text-sm font-semibold text-indigo-900">{{ currentSubjectName }}</p>
            </div>
            <span class="hidden text-xs text-indigo-500 sm:inline">AI 将按此科目考纲命制</span>
          </div>
          <div v-else class="mb-5 flex items-center gap-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3">
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-amber-100 text-amber-600">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
              </svg>
            </div>
            <p class="text-sm text-amber-800">请先在顶部导航选择要学习的科目，否则 AI 无法命制对应题目。</p>
          </div>

          <!-- API 未配置提示 -->
          <div v-if="!aiStore.hasConfig" class="mb-5 rounded-xl border border-yellow-200 bg-yellow-50 p-4">
            <div class="flex items-center gap-3">
              <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-yellow-100 text-yellow-600">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z" />
                </svg>
              </div>
              <div class="flex-1">
                <p class="text-sm font-semibold text-yellow-900">尚未配置 AI API</p>
                <p class="mt-0.5 text-xs text-yellow-700">请先前往 AI 配置页面设置你的 API Key，才能生成题目</p>
              </div>
              <BaseButton type="primary" size="sm" @click="router.push('/ai/config')">去配置</BaseButton>
            </div>
          </div>

          <!-- 主配置卡片 -->
          <BaseCard>
            <div v-if="error" class="mb-5 flex items-start gap-2 rounded-lg border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700">
              <svg class="mt-0.5 h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
              </svg>
              <span>{{ error }}</span>
            </div>

            <!-- 知识点范围 -->
            <div class="mb-6">
              <div class="mb-3 flex items-center gap-2">
                <div class="h-1 w-1 rounded-full bg-indigo-500"></div>
                <h3 class="text-sm font-semibold text-gray-800">知识点范围</h3>
                <span class="text-xs text-gray-400">题目将围绕此范围命制</span>
              </div>
              <div class="relative">
                <svg class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z" />
                </svg>
                <select
                  v-model="chapterId"
                  class="w-full appearance-none rounded-lg border border-gray-300 bg-white py-2.5 pl-10 pr-10 text-sm text-gray-800 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                >
                  <option v-for="ch in chapters" :key="ch.id" :value="ch.id">{{ ch.name }}</option>
                </select>
                <svg class="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                </svg>
              </div>
            </div>

            <!-- 题目类型 -->
            <div class="mb-6">
              <div class="mb-3 flex items-center gap-2">
                <div class="h-1 w-1 rounded-full bg-indigo-500"></div>
                <h3 class="text-sm font-semibold text-gray-800">题目类型</h3>
                <span class="text-xs text-gray-400">单次仅生成一种题型</span>
              </div>
              <div class="grid grid-cols-2 gap-2.5 sm:grid-cols-3 lg:grid-cols-5">
                <button
                  v-for="opt in typeOptions"
                  :key="opt.value"
                  class="group cursor-pointer rounded-xl border-2 p-3 text-left transition-all duration-200"
                  :class="questionType === opt.value ? getTypeColor(opt.color).active : getTypeColor(opt.color).inactive"
                  @click="selectType(opt.value)"
                >
                  <div class="mb-1.5 flex items-center gap-2">
                    <div
                      class="flex h-7 w-7 items-center justify-center rounded-lg transition-colors"
                      :class="getTypeColor(opt.color).icon"
                    >
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" :d="opt.icon" />
                      </svg>
                    </div>
                    <span class="text-sm font-semibold text-gray-800">{{ opt.label }}</span>
                  </div>
                  <p class="text-[11px] text-gray-500">{{ opt.desc }}</p>
                </button>
              </div>
            </div>

            <!-- 难度 + 数量（双列） -->
            <div class="mb-6 grid gap-6 sm:grid-cols-2">
              <!-- 难度 -->
              <div>
                <div class="mb-3 flex items-center gap-2">
                  <div class="h-1 w-1 rounded-full bg-indigo-500"></div>
                  <h3 class="text-sm font-semibold text-gray-800">难度</h3>
                </div>
                <div class="grid grid-cols-3 gap-2">
                  <button
                    v-for="opt in difficultyOptions"
                    :key="opt.value"
                    class="cursor-pointer rounded-xl border-2 p-2.5 text-center transition-all duration-200"
                    :class="difficulty === opt.value ? getTypeColor(opt.color).active : getTypeColor(opt.color).inactive"
                    @click="difficulty = opt.value"
                  >
                    <p class="text-sm font-semibold text-gray-800">{{ opt.label }}</p>
                    <p class="mt-0.5 text-[11px] text-gray-500">{{ opt.desc }}</p>
                  </button>
                </div>
              </div>

              <!-- 数量 -->
              <div>
                <div class="mb-3 flex items-center gap-2">
                  <div class="h-1 w-1 rounded-full bg-indigo-500"></div>
                  <h3 class="text-sm font-semibold text-gray-800">题目数量</h3>
                </div>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="n in countOptions"
                    :key="n"
                    class="cursor-pointer rounded-lg border px-3.5 py-2 text-sm font-medium transition-all duration-200"
                    :class="count === n
                      ? 'border-indigo-500 bg-indigo-50 text-indigo-700 shadow-sm'
                      : 'border-gray-200 bg-white text-gray-600 hover:border-indigo-300 hover:bg-indigo-50/30'"
                    @click="count = n"
                  >
                    {{ n }} 题
                  </button>
                </div>
              </div>
            </div>

            <!-- 操作区 -->
            <div class="flex flex-col-reverse gap-3 border-t border-gray-100 pt-5 sm:flex-row sm:items-center sm:justify-between">
              <p class="text-xs text-gray-400">
                <span class="font-medium text-gray-600">{{ typeLabel(questionType) }}</span>
                ·
                <span>{{ difficultyLabel(difficulty) }}</span>
                ·
                <span>{{ count }} 题</span>
              </p>
              <div class="flex gap-2 sm:justify-end">
                <BaseButton type="secondary" @click="router.push('/ai/config')">配置 AI</BaseButton>
                <BaseButton type="primary" :disabled="!aiStore.hasConfig" @click="startGenerate">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
                  </svg>
                  AI 生成题目
                </BaseButton>
              </div>
            </div>
          </BaseCard>
        </div>

        <!-- ============== 任务已提交 ============== -->
        <div v-else-if="step === 'pending'">
          <BaseCard>
            <div class="mb-5 flex items-center gap-4">
              <div class="relative flex h-12 w-12 shrink-0 items-center justify-center">
                <div class="absolute inset-0 animate-ping rounded-full bg-indigo-200 opacity-50"></div>
                <div class="relative flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br from-indigo-100 to-violet-100">
                  <svg class="h-6 w-6 animate-pulse text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
                  </svg>
                </div>
              </div>
              <div class="flex-1">
                <h3 class="text-base font-semibold text-gray-900">AI 正在生成题目</h3>
                <p class="mt-0.5 text-xs text-gray-500">已提交后台任务，你可以浏览其他页面，完成后会通过消息中心通知你</p>
              </div>
            </div>

            <div class="space-y-2.5 rounded-xl border border-gray-100 bg-gray-50/60 p-4 text-sm">
              <div class="flex items-center justify-between">
                <span class="text-gray-500">任务 ID</span>
                <code class="rounded bg-white px-2 py-0.5 font-mono text-xs text-gray-600 ring-1 ring-gray-200">{{ currentTask?.id?.slice(0, 16) }}…</code>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">状态</span>
                <span class="flex items-center gap-1.5 text-indigo-600">
                  <span class="relative flex h-2 w-2">
                    <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-indigo-400 opacity-75"></span>
                    <span class="relative inline-flex h-2 w-2 rounded-full bg-indigo-500"></span>
                  </span>
                  <span class="text-xs font-medium">{{ currentTask?.status === 'running' ? '生成中' : '排队中' }}</span>
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">题型</span>
                <span class="font-medium text-gray-800">{{ typeLabel(questionType) }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">数量</span>
                <span class="font-medium text-gray-800">{{ count }} 道</span>
              </div>
            </div>

            <div class="mt-5 flex items-center justify-between">
              <p class="text-xs text-gray-400">每 3 秒自动刷新状态</p>
              <BaseButton type="secondary" size="sm" @click="retry">取消并返回</BaseButton>
            </div>
          </BaseCard>
        </div>

        <!-- ============== 答题中：客观题 ============== -->
        <div v-else-if="step === 'practice' && usePracticeRunner">
          <div class="mb-4 flex items-center justify-between">
            <div>
              <h2 class="text-lg font-semibold text-gray-800">AI 出题练习</h2>
              <p class="mt-0.5 text-xs text-gray-500">作答后点击下方按钮查看结果</p>
            </div>
            <BaseButton type="success" @click="finishPractice">完成答题</BaseButton>
          </div>
          <PracticeRunner
            :questions="practiceQuestions"
            mode="ai"
            :hide-extract-when-has-knowledge="true"
            :submit-handler="async (qid, ans, duration) => {
              answers[qid] = ans
              submittedIds.add(qid)
              const res = await submitPractice({ question_id: qid, mode: 'ai', answer: ans, duration })
              return res
            }"
            @back="onPracticeBack"
          />
        </div>

        <!-- ============== 答题中：案例分析 ============== -->
        <div v-else-if="step === 'practice' && useCaseStudyRunner">
          <div class="sticky top-4 z-10 mb-4 flex items-center justify-between rounded-xl bg-white/95 p-3 shadow-sm ring-1 ring-gray-100 backdrop-blur">
            <div>
              <h2 class="text-lg font-semibold text-gray-800">案例分析作答</h2>
              <p class="mt-0.5 text-xs text-gray-500">阅读案例背景，回答下列子问题</p>
            </div>
            <BaseButton type="success" @click="finishPractice">完成作答</BaseButton>
          </div>

          <div class="space-y-5">
            <div
              v-for="(q, idx) in aiQuestions"
              :key="q.id"
              class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm sm:p-7"
            >
              <div class="mb-4 flex flex-wrap items-center gap-2">
                <span class="rounded-full bg-indigo-50 px-2.5 py-0.5 text-xs font-semibold text-indigo-600 ring-1 ring-inset ring-indigo-200">第 {{ idx + 1 }} 题</span>
                <span class="rounded-full bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-700 ring-1 ring-inset ring-amber-200">案例分析</span>
                <span v-if="q.knowledge_point" class="rounded bg-gray-50 px-2 py-0.5 text-xs text-gray-500">{{ q.knowledge_point }}</span>
                <button
                  v-else
                  type="button"
                  class="inline-flex cursor-pointer items-center gap-1 rounded-lg border border-indigo-200 bg-indigo-50/60 px-2.5 py-1 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-200 transition-all duration-200 hover:border-indigo-300 hover:bg-indigo-50 active:scale-[0.97]"
                  title="AI 提取核心考点"
                  @click="openExtractModal(q)"
                >
                  <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
                  </svg>
                  AI获取考点知识
                </button>
              </div>

              <div v-if="q.case_material" class="mb-5 rounded-lg border border-indigo-100 bg-indigo-50/40 p-4">
                <p class="mb-1.5 text-xs font-semibold text-indigo-700">【案例背景】</p>
                <p class="whitespace-pre-line text-sm leading-relaxed text-gray-800">{{ q.case_material }}</p>
              </div>

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
                    class="mt-2 w-full resize-y rounded-lg border border-gray-200 bg-white p-3 text-sm leading-relaxed text-gray-800 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                    :placeholder="`请输入第 ${sub.index} 题的答案...`"
                  />
                </div>
              </div>
            </div>
          </div>

          <div class="sticky bottom-4 mt-6 flex items-center justify-between rounded-xl bg-white/95 p-3 shadow-lg ring-1 ring-gray-100 backdrop-blur">
            <span class="text-xs text-gray-500">作答完成后点击右侧按钮查看参考答案与解析</span>
            <BaseButton type="success" @click="finishPractice">完成作答</BaseButton>
          </div>
        </div>

        <!-- ============== 结果页 ============== -->
        <div v-else-if="step === 'result'">
          <!-- 顶部统计 -->
          <div v-if="totalCount > 0" class="mb-5 grid grid-cols-3 gap-3 sm:gap-4">
            <div class="rounded-xl border border-indigo-100 bg-gradient-to-br from-indigo-50 to-white p-4 text-center sm:p-5">
              <p class="text-2xl font-bold tracking-tight text-indigo-600 sm:text-3xl">{{ totalCount }}</p>
              <p class="mt-1 text-xs text-gray-500">总题数</p>
            </div>
            <div class="rounded-xl border border-emerald-100 bg-gradient-to-br from-emerald-50 to-white p-4 text-center sm:p-5">
              <p class="text-2xl font-bold tracking-tight text-emerald-600 sm:text-3xl">{{ correctCount }}</p>
              <p class="mt-1 text-xs text-gray-500">正确</p>
            </div>
            <div
              class="rounded-xl border p-4 text-center sm:p-5"
              :class="accuracy >= 60 ? 'border-emerald-100 bg-gradient-to-br from-emerald-50 to-white' : 'border-red-100 bg-gradient-to-br from-red-50 to-white'"
            >
              <p class="text-2xl font-bold tracking-tight sm:text-3xl" :class="accuracy >= 60 ? 'text-emerald-600' : 'text-red-600'">{{ accuracy }}%</p>
              <p class="mt-1 text-xs text-gray-500">正确率</p>
            </div>
          </div>

          <BaseCard>
            <div class="mb-5 flex items-center justify-between">
              <h3 class="text-base font-semibold text-gray-900">
                {{ usePracticeRunner ? '答题结果' : useCaseStudyRunner ? '案例作答结果' : '生成结果' }}
              </h3>
              <span class="text-xs text-gray-400">共 {{ aiQuestions.length }} 题</span>
            </div>

            <div class="space-y-3">
              <div
                v-for="(q, idx) in aiQuestions"
                :key="q.id"
                class="rounded-lg border p-4 transition-colors"
                :class="usePracticeRunner && !isCorrect(q) ? 'border-red-200 bg-red-50/30' : 'border-gray-200 bg-white'"
              >
                <div class="flex items-start gap-2">
                  <span class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-gray-100 text-xs font-bold text-gray-600">
                    {{ idx + 1 }}
                  </span>
                  <div class="flex-1 min-w-0">
                    <div v-if="q.case_material" class="mb-3 rounded-lg border border-indigo-100 bg-indigo-50/40 p-3">
                      <p class="mb-1 text-xs font-medium text-indigo-700">【案例背景】</p>
                      <p class="whitespace-pre-line text-sm leading-relaxed text-gray-800">{{ q.case_material }}</p>
                    </div>

                    <p class="text-sm text-gray-800" v-html="q.content" />

                    <div class="mt-2 flex flex-wrap items-center gap-1.5">
                      <span class="rounded bg-white px-1.5 py-0.5 text-xs text-gray-500 ring-1 ring-gray-200">{{ typeLabel(q.type) }}</span>
                      <span v-if="q.knowledge_point" class="rounded bg-white px-1.5 py-0.5 text-xs text-gray-500 ring-1 ring-gray-200">{{ q.knowledge_point }}</span>
                    </div>

                    <template v-if="usePracticeRunner && q.type !== 'essay'">
                      <p class="mt-2 text-xs" :class="isCorrect(q) ? 'text-emerald-600' : 'text-red-600'">
                        {{ isCorrect(q) ? '✓ 正确' : '✗ 错误' }}
                        <span v-if="q.answer">&nbsp;正确答案：{{ q.answer }}</span>
                      </p>
                    </template>

                    <template v-else-if="q.type === 'case_study'">
                      <div class="mt-3 space-y-2.5">
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

                    <template v-else-if="q.type === 'essay'">
                      <p class="mt-2 text-xs text-gray-500">论文题无标准答案，不判分；已记录你的作答，可前往「AI 评分」获取专家点评</p>
                    </template>

                    <div v-if="q.analysis" class="mt-2 flex items-start gap-3">
                      <p class="flex-1 whitespace-pre-line text-xs leading-relaxed text-gray-600">{{ q.analysis }}</p>
                      <button
                        v-if="!q.knowledge_point"
                        type="button"
                        class="inline-flex shrink-0 cursor-pointer items-center gap-1 rounded-lg border border-indigo-200 bg-indigo-50/60 px-2.5 py-1 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-200 transition-all duration-200 hover:border-indigo-300 hover:bg-indigo-50 active:scale-[0.97]"
                        title="AI 提取核心考点"
                        @click="openExtractModal(q)"
                      >
                        <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
                        </svg>
                        AI获取考点知识
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="mt-6 flex justify-end gap-2 border-t border-gray-100 pt-5">
              <BaseButton type="secondary" @click="retry">重新生成</BaseButton>
            </div>
          </BaseCard>
        </div>
      </div>
    </div>

    <!-- AI 提取知识点弹窗:案例分析区/结果页"AI获取考点知识"按钮触发,复用 components/knowledge/ExtractKnowledgeModal.vue -->
    <ExtractKnowledgeModal
      v-if="extractTarget"
      v-model="showExtractModal"
      :question-id="extractTarget.id"
      :question-content="stripHtml(extractTarget.content)"
      :question-type="extractTarget.type"
      :question-answer="extractTarget.answer"
      :question-analysis="extractTarget.analysis"
      :subject-id="subjectId"
      @added="onKnowledgeAdded"
    />
  </div>
</template>
