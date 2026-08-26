<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAiStore } from '@/stores/ai'
import { submitGenerateAsync, getGenerateTask } from '@/api/ai'
import { formatSubmitError } from '@/utils/aiError'
import { showToast } from '@/utils/toast'
import { getSessionItem, setSessionItem, removeSessionItem } from '@/utils/storage'
import type { AsyncGenerateTask, AiGeneratedQuestion } from '@/types/ai'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

interface ChapterLike {
  id: number
  name: string
  subject_id: number
  sub_subject_id: number
  material_id?: number
}

const props = defineProps<{
  visible: boolean
  chapter: ChapterLike | null
  subjectName?: string
  subSubjectName?: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'generated', payload: { questions: AiGeneratedQuestion[]; chapterId: number; count: number }): void
}>()

const router = useRouter()
const aiStore = useAiStore()

type QuestionType = 'single' | 'multi' | 'judge' | 'case_study' | 'essay'
type Difficulty = 'easy' | 'medium' | 'hard'

const questionType = ref<QuestionType>('single')
const difficulty = ref<Difficulty>('medium')
const count = ref(5)
const submitting = ref(false)
const error = ref('')
const currentTask = ref<AsyncGenerateTask | null>(null)
const confirmExit = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const TASK_STORAGE_KEY = 'chapter_ai_task_id'

interface TypeOption {
  value: QuestionType
  label: string
  desc: string
  color: 'indigo' | 'violet' | 'emerald' | 'amber' | 'rose'
  icon: string
}

const typeOptions: TypeOption[] = [
  {
    value: 'single', label: '单选题', desc: '4 选 1',
    color: 'indigo',
    icon: 'M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM3.75 12h.007v.008H3.75V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm-.375 5.25h.007v.008H3.75v-.008zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z',
  },
  {
    value: 'multi', label: '多选题', desc: '多选 2~4',
    color: 'violet',
    icon: 'M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
  },
  {
    value: 'judge', label: '判断题', desc: '正确 / 错误',
    color: 'emerald',
    icon: 'M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z',
  },
  {
    value: 'case_study', label: '案例分析', desc: '背景 + 子问题',
    color: 'amber',
    icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z',
  },
  {
    value: 'essay', label: '论文题', desc: '仅出题',
    color: 'rose',
    icon: 'M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.0 0 015.25 6H10',
  },
]

interface DifficultyOption {
  value: Difficulty
  label: string
  desc: string
  color: 'emerald' | 'amber' | 'red'
}

const difficultyOptions: DifficultyOption[] = [
  { value: 'easy', label: '简单', desc: '基础概念', color: 'emerald' },
  { value: 'medium', label: '中等', desc: '综合应用', color: 'amber' },
  { value: 'hard', label: '困难', desc: '深入分析', color: 'red' },
]

const countOptions = computed<number[]>(() => {
  if (questionType.value === 'case_study' || questionType.value === 'essay') return [1, 2, 3]
  return [3, 5, 10, 15, 20]
})

const colorClassMap: Record<string, { active: string; inactive: string; icon: string; dot: string }> = {
  indigo: {
    active: 'border-indigo-500 bg-indigo-50/60 ring-1 ring-indigo-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-indigo-300 hover:bg-indigo-50/30',
    icon: 'bg-indigo-50 text-indigo-600',
    dot: 'bg-indigo-500',
  },
  violet: {
    active: 'border-violet-500 bg-violet-50/60 ring-1 ring-violet-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-violet-300 hover:bg-violet-50/30',
    icon: 'bg-violet-50 text-violet-600',
    dot: 'bg-violet-500',
  },
  emerald: {
    active: 'border-emerald-500 bg-emerald-50/60 ring-1 ring-emerald-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-emerald-300 hover:bg-emerald-50/30',
    icon: 'bg-emerald-50 text-emerald-600',
    dot: 'bg-emerald-500',
  },
  amber: {
    active: 'border-amber-500 bg-amber-50/60 ring-1 ring-amber-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-amber-300 hover:bg-amber-50/30',
    icon: 'bg-amber-50 text-amber-600',
    dot: 'bg-amber-500',
  },
  rose: {
    active: 'border-rose-500 bg-rose-50/60 ring-1 ring-rose-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-rose-300 hover:bg-rose-50/30',
    icon: 'bg-rose-50 text-rose-600',
    dot: 'bg-rose-500',
  },
  red: {
    active: 'border-red-500 bg-red-50/60 ring-1 ring-red-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-red-300 hover:bg-red-50/30',
    icon: 'bg-red-50 text-red-600',
    dot: 'bg-red-500',
  },
}

function getTypeColor(c: string) { return colorClassMap[c] || colorClassMap.indigo }
function getDifficultyColor(c: string) { return colorClassMap[c] || colorClassMap.amber }

const taskInFlight = computed(() => {
  if (!currentTask.value) return false
  return currentTask.value.status === 'pending' || currentTask.value.status === 'running'
})

const hasMaterial = computed(() => !!props.chapter?.material_id && props.chapter.material_id > 0)

watch(questionType, (t) => {
  if (t === 'case_study' || t === 'essay') {
    if (count.value > 3) count.value = 1
  } else if (count.value === 1 || count.value === 2) {
    count.value = 5
  }
})

watch(() => props.visible, async (v) => {
  if (v) {
    error.value = ''
    aiStore.syncConfig()
    const savedId = getSessionItem(TASK_STORAGE_KEY)
    if (savedId) {
      try {
        const t = await getGenerateTask(savedId)
        if (t && (t.status === 'pending' || t.status === 'running')) {
          currentTask.value = t
          startPolling(savedId)
        } else {
          removeSessionItem(TASK_STORAGE_KEY)
        }
      } catch {
        removeSessionItem(TASK_STORAGE_KEY)
      }
    }
  } else {
    stopPolling()
  }
})

onUnmounted(() => {
  stopPolling()
})

function close() {
  if (taskInFlight.value) {
    confirmExit.value = true
    return
  }
  emit('update:visible', false)
}

function forceClose() {
  stopPolling()
  confirmExit.value = false
  emit('update:visible', false)
}

function goConfig() {
  showToast('请先在「AI 配置」中填写 API Key', 'info')
  router.push('/ai/config')
  emit('update:visible', false)
}

async function startGenerate() {
  if (!aiStore.hasConfig) {
    goConfig()
    return
  }
  if (!props.chapter) {
    error.value = '章节信息缺失'
    return
  }
  if (!props.chapter.subject_id) {
    error.value = '章节未关联科目，请刷新页面后重试'
    return
  }

  error.value = ''
  currentTask.value = null
  submitting.value = true

  try {
    const resp = await submitGenerateAsync({
      api_config: aiStore.getConfig(),
      subject_id: props.chapter.subject_id,
      chapter_id: props.chapter.id,
      types: [questionType.value],
      difficulty: difficulty.value,
      count: count.value,
    })
    setSessionItem(TASK_STORAGE_KEY, resp.task_id)
    currentTask.value = {
      id: resp.task_id,
      status: resp.status || 'pending',
      question_type: questionType.value,
      chapter_id: props.chapter.id,
      chapter_name: props.chapter.name,
      difficulty: difficulty.value,
      count: count.value,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }
    submitting.value = false
    startPolling(resp.task_id)
  } catch (e) {
    submitting.value = false
    error.value = formatSubmitError((e as Error).message)
  }
}

function startPolling(taskId: string) {
  stopPolling()
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
      removeSessionItem(TASK_STORAGE_KEY)
      handleSuccess(t)
    } else if (t.status === 'failed') {
      stopPolling()
      removeSessionItem(TASK_STORAGE_KEY)
      error.value = t.error || 'AI 出题失败，请稍后重试'
    }
  } catch {
    stopPolling()
    error.value = '任务查询失败，请稍后重试'
  }
}

function handleSuccess(t: AsyncGenerateTask) {
  if (!t.result || !t.result.questions || t.result.questions.length === 0) {
    error.value = '任务完成但未返回题目，请稍后在「AI 出题历史」中查看'
    return
  }
  const count = t.result.questions.length
  showToast(`已为「${props.chapter?.name}」新增 ${count} 道 AI 题目，可立即练习`, 'success')
  emit('generated', {
    questions: t.result.questions,
    chapterId: props.chapter?.id || 0,
    count,
  })
  emit('update:visible', false)
}

function retry() {
  error.value = ''
  startGenerate()
}

function typeLabel(t: QuestionType) {
  return typeOptions.find(o => o.value === t)?.label || t
}
</script>

<template>
  <BaseModal
    :model-value="visible"
    :title="`为「${chapter?.name || ''}」AI 出题`"
    width="max-w-xl"
    @update:model-value="(v) => v ? null : close()"
  >
    <div class="space-y-5">
      <!-- 上下文芯片 -->
      <div class="flex flex-wrap items-center gap-2 rounded-xl border border-indigo-100 bg-indigo-50/40 p-3">
        <svg class="h-4 w-4 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456z" />
        </svg>
        <span class="text-xs text-gray-600">将基于本章节知识点（考纲 + 已上传文档）出题</span>
        <span v-if="subjectName" class="rounded-full bg-white px-2 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-200">
          {{ subjectName }}
        </span>
        <span v-if="subSubjectName" class="rounded-full bg-white px-2 py-0.5 text-xs font-medium text-violet-700 ring-1 ring-inset ring-violet-200">
          {{ subSubjectName }}
        </span>
        <span class="rounded-full bg-white px-2 py-0.5 text-xs font-medium text-gray-700 ring-1 ring-inset ring-gray-200">
          {{ chapter?.name }}
        </span>
        <span
          v-if="!hasMaterial"
          class="ml-auto rounded-full bg-amber-50 px-2 py-0.5 text-xs text-amber-700 ring-1 ring-inset ring-amber-200"
          title="未上传知识点文档，出题质量取决于考纲与章节名匹配度"
        >
          未上传文档
        </span>
      </div>

      <!-- 生成中状态 -->
      <div
        v-if="taskInFlight"
        class="rounded-xl border border-indigo-100 bg-gradient-to-br from-indigo-50/60 to-violet-50/60 p-5"
      >
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-full bg-white shadow-sm">
            <svg class="h-5 w-5 animate-spin text-indigo-600" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-semibold text-gray-900">
              {{ currentTask?.status === 'pending' ? '排队中…' : 'AI 正在出题…' }}
            </p>
            <p class="mt-0.5 text-xs text-gray-500">
              本次将生成 {{ currentTask?.count }} 道 {{ typeLabel((currentTask?.question_type || questionType) as QuestionType) }}，通常 30s~3min，请保持页面打开或稍后到「AI 出题历史」查看
            </p>
          </div>
        </div>
      </div>

      <!-- 错误提示 -->
      <div
        v-else-if="error"
        class="rounded-xl border border-red-100 bg-red-50/60 p-4"
      >
        <div class="flex items-start gap-2">
          <svg class="mt-0.5 h-4 w-4 shrink-0 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
          </svg>
          <div class="min-w-0 flex-1">
            <p class="text-sm text-red-700">{{ error }}</p>
            <div class="mt-2 flex gap-2">
              <BaseButton type="secondary" size="sm" @click="error = ''">关闭</BaseButton>
              <BaseButton type="primary" size="sm" @click="retry">重试</BaseButton>
            </div>
          </div>
        </div>
      </div>

      <!-- 配置表单 -->
      <template v-else>
        <!-- 题型选择 -->
        <div>
          <div class="mb-2 flex items-center justify-between">
            <label class="text-sm font-semibold text-gray-900">题型</label>
            <span class="text-xs text-gray-400">单选</span>
          </div>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-5">
            <button
              v-for="opt in typeOptions"
              :key="opt.value"
              type="button"
              :disabled="taskInFlight"
              :class="[
                'group flex flex-col items-center gap-1 rounded-xl border-2 px-2.5 py-2.5 text-center transition-all duration-200 disabled:cursor-not-allowed disabled:opacity-50',
                questionType === opt.value
                  ? getTypeColor(opt.color).active
                  : getTypeColor(opt.color).inactive,
              ]"
              @click="questionType = opt.value"
            >
              <span :class="['flex h-7 w-7 items-center justify-center rounded-full', getTypeColor(opt.color).icon]">
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" :d="opt.icon" />
                </svg>
              </span>
              <span class="text-xs font-medium text-gray-900">{{ opt.label }}</span>
              <span class="text-[10px] text-gray-400">{{ opt.desc }}</span>
            </button>
          </div>
        </div>

        <!-- 难度选择 -->
        <div>
          <label class="mb-2 block text-sm font-semibold text-gray-900">难度</label>
          <div class="grid grid-cols-3 gap-2">
            <button
              v-for="opt in difficultyOptions"
              :key="opt.value"
              type="button"
              :disabled="taskInFlight"
              :class="[
                'flex items-center justify-center gap-1.5 rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200 disabled:cursor-not-allowed disabled:opacity-50',
                difficulty === opt.value
                  ? getDifficultyColor(opt.color).active
                  : getDifficultyColor(opt.color).inactive,
              ]"
              @click="difficulty = opt.value"
            >
              <span :class="['h-2 w-2 rounded-full', getDifficultyColor(opt.color).dot]" />
              {{ opt.label }}
              <span class="hidden text-[10px] text-gray-400 sm:inline">· {{ opt.desc }}</span>
            </button>
          </div>
        </div>

        <!-- 数量选择 -->
        <div>
          <label class="mb-2 block text-sm font-semibold text-gray-900">数量</label>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="c in countOptions"
              :key="c"
              type="button"
              :disabled="taskInFlight"
              :class="[
                'rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200 disabled:cursor-not-allowed disabled:opacity-50',
                count === c
                  ? 'border-indigo-500 bg-indigo-50 text-indigo-700 shadow-sm'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-indigo-300 hover:bg-indigo-50/30',
              ]"
              @click="count = c"
            >
              {{ c }} 题
            </button>
          </div>
          <p v-if="(questionType === 'case_study' || questionType === 'essay')" class="mt-1.5 text-xs text-amber-600">
            案例分析 / 论文题单次输出内容较长，单批最多 3 道
          </p>
        </div>
      </template>
    </div>

    <!-- 底部操作 -->
    <template v-if="!taskInFlight && !error">
      <div class="mt-6 flex items-center justify-between gap-3 border-t border-gray-100 pt-5">
        <div class="text-xs text-gray-400">
          预计耗时 30s~3min · 完成后会保留在题库中
        </div>
        <div class="flex gap-2">
          <BaseButton type="secondary" @click="close">取消</BaseButton>
          <BaseButton type="primary" :loading="submitting" @click="startGenerate">
            <span class="flex items-center gap-1.5">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
              </svg>
              开始生成 {{ count }} 题
            </span>
          </BaseButton>
        </div>
      </div>
    </template>
    <template v-else-if="taskInFlight">
      <div class="mt-6 flex justify-end gap-2 border-t border-gray-100 pt-5">
        <BaseButton type="secondary" @click="close">后台运行并关闭</BaseButton>
      </div>
    </template>

    <ConfirmDialog
      v-if="confirmExit"
      title="任务正在生成中"
      description="关闭弹窗后任务会在后台继续运行，你仍可在「AI 出题历史」中查看进度。确定要关闭吗？"
      confirm-text="关闭"
      @confirm="forceClose"
      @cancel="confirmExit = false"
    />
  </BaseModal>
</template>
