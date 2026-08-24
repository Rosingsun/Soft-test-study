<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getSpecialQuestions, getEssayQuestions, generateQuestions } from '@/api/practice'
import { useAiStore } from '@/stores/ai'
import { useSubjectStore } from '@/stores/subject'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import type { Question } from '@/types/question'

const router = useRouter()
const auth = useAuthStore()
const aiStore = useAiStore()
const subjectStore = useSubjectStore()

const subjectId = computed(() => auth.selectedSubjectId)
const currentSubjectName = computed(() => {
  const id = subjectId.value
  if (!id) return auth.user?.subject_name || '未选择'
  return subjectStore.subjects.find(s => s.id === id)?.name || auth.user?.subject_name || `科目#${id}`
})

const type = ref('')
const difficulty = ref('')
const count = ref(10)
const questions = ref<Question[]>([])
const started = ref(false)
const loading = ref(false)
const error = ref('')
const essayTopics = ref<Question[]>([])
const selectedEssayId = ref(0)
const essayMode = ref<'fixed' | 'random' | 'ai'>('fixed')

const typeOptions = [
  { value: '', label: '全部题型', desc: '混合练习', icon: 'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z' },
  { value: 'single', label: '单选题', desc: '4 选 1', icon: 'M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM3.75 12h.007v.008H3.75V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm-.375 5.25h.007v.008H3.75v-.008zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z' },
  { value: 'multi', label: '多选题', desc: '多选 2~4', icon: 'M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
  { value: 'judge', label: '判断题', desc: '正确 / 错误', icon: 'M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z' },
  { value: 'fill', label: '填空题', desc: '填入答案', icon: 'M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L14.146 6.32a1.875 1.875 0 012.652 2.652z' },
  { value: 'short', label: '简答题', desc: '简要作答', icon: 'M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z' },
  { value: 'comprehensive', label: '综合题', desc: '多知识点', icon: 'M6 6.878V6a2.25 2.25 0 012.25-2.25h7.5A2.25 2.25 0 0118 6v.878m-12 0c.235-.083.487-.128.75-.128h10.5c.263 0 .515.045.75.128m-12 0A2.25 2.25 0 004.5 9v.878m13.5-3A2.25 2.25 0 0119.5 9v.878m0 0a2.246 2.246 0 00-.75-.128H5.25a2.246 2.246 0 00-.75.128m15 0A2.25 2.25 0 0121 12v6a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 18v-6a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 12z' },
  { value: 'case_study', label: '案例分析', desc: '背景 + 子问题', icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z' },
  { value: 'essay', label: '论文题', desc: '长文作答', icon: 'M4.26 10.147a60.438 60.438 0 0 0-.491 6.347A48.62 48.62 0 0 1 12 20.904a48.62 48.62 0 0 1 8.232-4.41 60.46 60.46 0 0 0-.491-6.347m-15.482 0a50.636 50.636 0 0 0-2.658-.813A59.906 59.906 0 0 1 12 3.493a59.903 59.903 0 0 1 10.399 5.84c-.896.248-1.783.52-2.658.814m-15.482 0A50.717 50.717 0 0 1 12 13.489a50.702 50.702 0 0 1 7.74-3.342M6.75 15a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm0 0v-3.675A55.378 55.378 0 0 1 12 8.443m-7.007 11.55A5.981 5.981 0 0 0 6.75 15.75v-1.5' },
  { value: 'multi_blank', label: '多空题', desc: '多空填写', icon: 'M8.242 5.992a.75.75 0 01-.75.75H4.5a.75.75 0 010-1.5h2.992a.75.75 0 01.75.75zM8.242 11.996a.75.75 0 01-.75.75H4.5a.75.75 0 010-1.5h2.992a.75.75 0 01.75.75zM8.242 17.752a.75.75 0 01-.75.75H4.5a.75.75 0 010-1.5h2.992a.75.75 0 01.75.75zM19.5 6h-7.5v1.5h7.5V6zM19.5 12h-7.5v1.5h7.5V12zM19.5 18h-7.5v1.5h7.5V18z' },
]

const difficultyOptions = [
  { value: '', label: '全部', desc: '不限', color: 'gray' },
  { value: 'easy', label: '简单', desc: '基础概念', color: 'emerald' },
  { value: 'medium', label: '中等', desc: '综合应用', color: 'amber' },
  { value: 'hard', label: '困难', desc: '深入分析', color: 'red' },
]

const essayModeOptions = [
  { value: 'fixed', label: '固定题目', desc: '从题库选择历年真题', icon: 'M5.25 8.25h15m-16.5 7.5h15m-1.8-13.5l-3.9 19.5m-2.1-19.5l-3.9 19.5' },
  { value: 'random', label: '随机抽取', desc: '从题库随机一题', icon: 'M19.5 12c0-1.232-.046-2.453-.138-3.662a4.006 4.006 0 00-3.7-3.7 48.678 48.678 0 00-7.324 0 4.006 4.006 0 00-3.7 3.7c-.017.22-.032.441-.046.662M19.5 12l3-3m-3 3l-3-3m-12 3c0 1.232.046 2.453.138 3.662a4.006 4.006 0 003.7 3.7 48.656 48.656 0 007.324 0 4.006 4.006 0 003.7-3.7c.017-.22.032-.441.046-.662M4.5 12l3 3m-3-3l-3 3' },
  { value: 'ai', label: 'AI 生成', desc: '由 AI 现场出题', icon: 'M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z' },
]

const isEssay = computed(() => type.value === 'essay')
const showCount = computed(() => !isEssay.value)

const colorClassMap: Record<string, { active: string; inactive: string; icon: string }> = {
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
  red: {
    active: 'border-red-500 bg-red-50/60 ring-1 ring-red-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-red-300 hover:bg-red-50/30',
    icon: 'bg-red-50 text-red-600',
  },
  gray: {
    active: 'border-gray-500 bg-gray-100 ring-1 ring-gray-500/30 shadow-sm',
    inactive: 'border-gray-200 bg-white hover:border-gray-300 hover:bg-gray-50',
    icon: 'bg-gray-100 text-gray-600',
  },
}

function getColorStyle(c: string) {
  return colorClassMap[c] || colorClassMap.gray
}

onMounted(async () => {
  if (!auth.initialized) await auth.checkAuth()
  if (subjectId.value && !subjectStore.subjects.length) {
    try {
      if (auth.selectedLevelId) {
        await subjectStore.fetchSubjectsByLevel(auth.selectedLevelId)
      }
    } catch {
      // 静默忽略
    }
  }
  if (isEssay.value && subjectId.value) {
    await loadEssayTopics()
  }
})

watch(subjectId, async (val) => {
  if (isEssay.value && val) {
    await loadEssayTopics()
  }
})

watch(type, async (val) => {
  if (val === 'essay') {
    count.value = 1
    essayMode.value = 'fixed'
    selectedEssayId.value = 0
    if (subjectId.value) await loadEssayTopics()
  }
})

async function loadEssayTopics() {
  try {
    essayTopics.value = await getEssayQuestions(subjectId.value, 5)
  } catch (e) {
    essayTopics.value = []
  }
}

function resetEssayModeIfNeeded() {
  if (essayMode.value !== 'fixed') {
    selectedEssayId.value = 0
  }
}

async function start() {
  if (!subjectId.value) {
    error.value = '请先选择科目'
    return
  }

  if (isEssay.value) {
    if (essayMode.value === 'fixed' && selectedEssayId.value === 0) {
      error.value = '请选择论文题目或切换为随机/AI生成'
      return
    }
    count.value = 1
  }

  started.value = true
  loading.value = true
  error.value = ''

  try {
    if (isEssay.value && essayMode.value === 'fixed') {
      const topic = essayTopics.value.find(item => item.id === selectedEssayId.value)
      if (!topic) {
        throw new Error('所选论文题目不存在，请重新选择')
      }
      questions.value = [topic]
    } else if (isEssay.value && essayMode.value === 'random') {
      questions.value = await getSpecialQuestions(subjectId.value, 'essay', difficulty.value || undefined, 1)
    } else if (isEssay.value && essayMode.value === 'ai') {
      if (!aiStore.hasConfig) {
        throw new Error('尚未配置 AI API，请先在 AI 配置 中保存 API 信息')
      }
      const resp = await generateQuestions({
        api_config: aiStore.getConfig(),
        subject_id: subjectId.value,
        chapter_id: 0,
        types: ['essay'],
        difficulty: difficulty.value || undefined,
        count: 1,
      })
      questions.value = (resp.questions || []).map((q: any) => ({
        id: q.id,
        subject_id: subjectId.value,
        sub_subject_id: 0,
        chapter_id: 0,
        parent_id: 0,
        type: isEssay.value ? 'essay' : (q.type || 'essay'),
        difficulty: q.difficulty || 'medium',
        content: q.content,
        options: '[]',
        answer: q.answer || '',
        analysis: q.analysis || '',
        year: q.year || 0,
        source: q.source || '',
      }))
    } else {
      questions.value = await getSpecialQuestions(subjectId.value, type.value, difficulty.value || undefined, count.value)
    }

    if (questions.value.length === 0) {
      started.value = false
      error.value = '当前条件下暂无题目，请调整筛选条件'
    }
  } catch (e) {
    started.value = false
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function reset() {
  started.value = false
  questions.value = []
}

function typeLabelByValue(v: string) {
  return typeOptions.find(t => t.value === v)?.label || '全部题型'
}

function difficultyLabelByValue(v: string) {
  return difficultyOptions.find(d => d.value === v)?.label || '全部难度'
}
</script>

<template>
  <div>
    <BasePageHeader title="专项练习" subtitle="针对特定题型集中训练，逐个突破薄弱环节">
      <template #actions>
        <span class="inline-flex items-center gap-1.5 rounded-full bg-indigo-50 px-3 py-1 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-600/20">
          <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0l4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0l-5.571 3-5.571-3" />
          </svg>
          题源：真题 + AI 题目
        </span>
        <BaseButton type="secondary" @click="router.push('/practice/random')">
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 12c0-1.232-.046-2.453-.138-3.662a4.006 4.006 0 00-3.7-3.7 48.678 48.678 0 00-7.324 0 4.006 4.006 0 00-3.7 3.7c-.017.22-.032.441-.046.662M19.5 12l3-3m-3 3l-3-3m-12 3c0 1.232.046 2.453.138 3.662a4.006 4.006 0 003.7 3.7 48.656 48.656 0 007.324 0 4.006 4.006 0 003.7-3.7c.017-.22.032-.441-.046-.662M4.5 12l3 3m-3-3l-3 3" />
          </svg>
          切换随机练习
        </BaseButton>
      </template>
    </BasePageHeader>

    <div class="flex flex-col gap-6 lg:flex-row lg:items-start">
      <!-- 左侧：主配置 -->
      <div class="min-w-0 flex-1">
        <!-- 科目提示 -->
        <div
          v-if="subjectId"
          class="mb-5 flex items-center gap-3 rounded-xl border border-emerald-100 bg-gradient-to-r from-emerald-50/80 to-teal-50/60 px-4 py-3"
        >
          <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-white text-emerald-600 shadow-sm ring-1 ring-emerald-100">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-xs text-emerald-600/80">当前科目</p>
            <p class="text-sm font-semibold text-emerald-900">{{ currentSubjectName }}</p>
          </div>
          <span class="hidden text-xs text-emerald-500 sm:inline">训练题目将从此科目抽取</span>
        </div>
        <div v-else class="mb-5 flex items-center gap-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3">
          <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-amber-100 text-amber-600">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
            </svg>
          </div>
          <p class="text-sm text-amber-800">请先在顶部导航选择要学习的科目。</p>
        </div>

        <!-- 配置面板 -->
        <div v-if="!started">
          <BaseCard>
            <div v-if="error" class="mb-5 flex items-start gap-2 rounded-lg border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700">
              <svg class="mt-0.5 h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
              </svg>
              <span>{{ error }}</span>
            </div>

            <!-- 题型 -->
            <div class="mb-6">
              <div class="mb-3 flex items-center gap-2">
                <div class="h-1 w-1 rounded-full bg-emerald-500"></div>
                <h3 class="text-sm font-semibold text-gray-800">题型</h3>
                <span class="text-xs text-gray-400">选择要集中训练的题型</span>
              </div>
              <div class="grid grid-cols-2 gap-2.5 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                <button
                  v-for="t in typeOptions"
                  :key="t.value"
                  class="group cursor-pointer rounded-xl border-2 p-3 text-left transition-all duration-200"
                  :class="type === t.value
                    ? 'border-emerald-500 bg-emerald-50/60 ring-1 ring-emerald-500/30 shadow-sm'
                    : 'border-gray-200 bg-white hover:border-emerald-300 hover:bg-emerald-50/30'"
                  @click="type = t.value"
                >
                  <div class="mb-1.5 flex items-center gap-2">
                    <div
                      class="flex h-7 w-7 items-center justify-center rounded-lg transition-colors"
                      :class="type === t.value ? 'bg-emerald-100 text-emerald-600' : 'bg-gray-50 text-gray-500 group-hover:bg-emerald-50 group-hover:text-emerald-600'"
                    >
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" :d="t.icon" />
                      </svg>
                    </div>
                    <span class="text-sm font-semibold text-gray-800">{{ t.label }}</span>
                  </div>
                  <p class="text-[11px] text-gray-500">{{ t.desc }}</p>
                </button>
              </div>
            </div>

            <!-- 难度 + 数量（双列） -->
            <div class="mb-6 grid gap-6 sm:grid-cols-2">
              <!-- 难度 -->
              <div>
                <div class="mb-3 flex items-center gap-2">
                  <div class="h-1 w-1 rounded-full bg-emerald-500"></div>
                  <h3 class="text-sm font-semibold text-gray-800">难度</h3>
                </div>
                <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
                  <button
                    v-for="opt in difficultyOptions"
                    :key="opt.value"
                    class="cursor-pointer rounded-xl border-2 p-2.5 text-center transition-all duration-200"
                    :class="difficulty === opt.value ? getColorStyle(opt.color).active : getColorStyle(opt.color).inactive"
                    @click="difficulty = opt.value"
                  >
                    <p class="text-sm font-semibold text-gray-800">{{ opt.label }}</p>
                    <p class="mt-0.5 text-[11px] text-gray-500">{{ opt.desc }}</p>
                  </button>
                </div>
              </div>

              <!-- 数量 -->
              <div v-if="showCount">
                <div class="mb-3 flex items-center gap-2">
                  <div class="h-1 w-1 rounded-full bg-emerald-500"></div>
                  <h3 class="text-sm font-semibold text-gray-800">题目数量</h3>
                </div>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="n in [5, 10, 15, 20]"
                    :key="n"
                    class="cursor-pointer rounded-lg border px-3.5 py-2 text-sm font-medium transition-all duration-200"
                    :class="count === n
                      ? 'border-emerald-500 bg-emerald-50 text-emerald-700 shadow-sm'
                      : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-300 hover:bg-emerald-50/30'"
                    @click="count = n"
                  >
                    {{ n }} 题
                  </button>
                </div>
                <p v-if="type === 'case_study'" class="mt-2.5 flex items-start gap-1.5 text-xs text-gray-400">
                  <svg class="mt-0.5 h-3.5 w-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
                  </svg>
                  案例分析按大题计数，每道大题包含若干小题，小题不计入总数量
                </p>
              </div>
            </div>

            <!-- 论文题专属 -->
            <div v-if="isEssay" class="mb-2">
              <div class="mb-3 flex items-center gap-2">
                <div class="h-1 w-1 rounded-full bg-emerald-500"></div>
                <h3 class="text-sm font-semibold text-gray-800">论文题目来源</h3>
                <span class="text-xs text-gray-400">单次训练一题</span>
              </div>
              <div class="grid gap-2.5 sm:grid-cols-3">
                <button
                  v-for="m in essayModeOptions"
                  :key="m.value"
                  class="group cursor-pointer rounded-xl border-2 p-3 text-left transition-all duration-200"
                  :class="essayMode === m.value
                    ? 'border-emerald-500 bg-emerald-50/60 ring-1 ring-emerald-500/30 shadow-sm'
                    : 'border-gray-200 bg-white hover:border-emerald-300 hover:bg-emerald-50/30'"
                  @click="essayMode = m.value as 'fixed' | 'random' | 'ai'; resetEssayModeIfNeeded()"
                >
                  <div class="mb-1.5 flex items-center gap-2">
                    <div
                      class="flex h-7 w-7 items-center justify-center rounded-lg transition-colors"
                      :class="essayMode === m.value ? 'bg-emerald-100 text-emerald-600' : 'bg-gray-50 text-gray-500 group-hover:bg-emerald-50 group-hover:text-emerald-600'"
                    >
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" :d="m.icon" />
                      </svg>
                    </div>
                    <span class="text-sm font-semibold text-gray-800">{{ m.label }}</span>
                  </div>
                  <p class="text-[11px] text-gray-500">{{ m.desc }}</p>
                </button>
              </div>

              <!-- 固定题目时的下拉 -->
              <div v-if="essayMode === 'fixed'" class="mt-4">
                <label class="mb-1.5 block text-xs font-medium text-gray-600">选择论文题目</label>
                <div class="relative">
                  <svg class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                  </svg>
                  <select
                    v-model="selectedEssayId"
                    class="w-full appearance-none rounded-lg border border-gray-300 bg-white py-2.5 pl-10 pr-10 text-sm text-gray-800 transition-colors focus:border-emerald-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
                  >
                    <option value="0">请选择论文题目</option>
                    <option v-for="topic in essayTopics" :key="topic.id" :value="topic.id">
                      {{ topic.year }} {{ topic.content }}
                    </option>
                  </select>
                  <svg class="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                  </svg>
                </div>
              </div>
              <div v-else class="mt-3 flex items-start gap-2 rounded-lg border border-emerald-100 bg-emerald-50/40 px-3 py-2.5 text-xs text-emerald-700">
                <svg class="mt-0.5 h-3.5 w-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
                </svg>
                <span v-if="essayMode === 'random'">选择「随机抽取」后，将从题库随机选一题。</span>
                <span v-else>选择「AI 生成」后，将调用 AI 现场出题，请先在 AI 配置中设置 API Key。</span>
              </div>
            </div>

            <!-- 操作区 -->
            <div class="mt-6 flex flex-col-reverse gap-3 border-t border-gray-100 pt-5 sm:flex-row sm:items-center sm:justify-between">
              <p class="text-xs text-gray-400">
                <span class="font-medium text-gray-600">{{ typeLabelByValue(type) }}</span>
                <template v-if="difficulty">
                  ·
                  <span>{{ difficultyLabelByValue(difficulty) }}</span>
                </template>
                <template v-if="showCount">
                  ·
                  <span>{{ count }} 题</span>
                </template>
                <template v-if="isEssay">
                  ·
                  <span class="text-emerald-600">{{ essayModeOptions.find(m => m.value === essayMode)?.label }}</span>
                </template>
              </p>
              <div class="flex gap-2 sm:justify-end">
                <BaseButton type="secondary" @click="router.push('/practice/random')">随机练习</BaseButton>
                <BaseButton type="success" @click="start">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 010 1.972l-11.54 6.347a1.125 1.125 0 01-1.667-.986V5.653z" />
                  </svg>
                  开始训练
                </BaseButton>
              </div>
            </div>
          </BaseCard>
        </div>

        <!-- 训练中 -->
        <div v-else>
          <BaseSkeleton v-if="loading" variant="question" />
          <PracticeRunner
            v-else
            :questions="questions"
            mode="special"
            @back="reset"
          />
        </div>
      </div>

      <!-- 右侧：训练贴士 -->
      <aside v-if="!started" class="shrink-0 lg:w-72">
        <BaseCard>
          <div class="mb-4 flex items-center gap-2">
            <div class="flex h-7 w-7 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 18v-5.25m0 0a6.01 6.01 0 001.5-.189m-1.5.189a6.01 6.01 0 01-1.5-.189m3.75 7.478a12.06 12.06 0 01-4.5 0m3.75 2.383a14.406 14.406 0 01-3 0M14.25 18v-.192c0-.983.658-1.823 1.508-2.316a7.5 7.5 0 10-7.517 0c.85.493 1.509 1.333 1.509 2.316V18" />
              </svg>
            </div>
            <h2 class="text-sm font-semibold text-gray-800">训练贴士</h2>
          </div>

          <div class="space-y-3.5">
            <div class="flex gap-3">
              <div class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
                <span class="text-xs font-bold">1</span>
              </div>
              <div>
                <p class="text-sm font-medium text-gray-800">先选弱项题型</p>
                <p class="mt-0.5 text-xs leading-relaxed text-gray-500">通过错题统计找到薄弱环节，针对性训练更高效</p>
              </div>
            </div>
            <div class="flex gap-3">
              <div class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
                <span class="text-xs font-bold">2</span>
              </div>
              <div>
                <p class="text-sm font-medium text-gray-800">由易到难</p>
                <p class="mt-0.5 text-xs leading-relaxed text-gray-500">建议从「简单」开始建立信心，再逐步挑战更高难度</p>
              </div>
            </div>
            <div class="flex gap-3">
              <div class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
                <span class="text-xs font-bold">3</span>
              </div>
              <div>
                <p class="text-sm font-medium text-gray-800">关注解析</p>
                <p class="mt-0.5 text-xs leading-relaxed text-gray-500">错题不只看正确答案，更要理解解析里的知识点</p>
              </div>
            </div>
            <div class="flex gap-3">
              <div class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
                <span class="text-xs font-bold">4</span>
              </div>
              <div>
                <p class="text-sm font-medium text-gray-800">论文 AI 评分</p>
                <p class="mt-0.5 text-xs leading-relaxed text-gray-500">论文题作答后可一键获取 AI 专家点评和评分</p>
              </div>
            </div>
          </div>
        </BaseCard>

        <!-- 快捷入口 -->
        <BaseCard class="mt-4">
          <div class="mb-3 flex items-center gap-2">
            <div class="flex h-7 w-7 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
              </svg>
            </div>
            <h2 class="text-sm font-semibold text-gray-800">其他练习模式</h2>
          </div>
          <div class="space-y-1.5">
            <button
              class="flex w-full cursor-pointer items-center justify-between rounded-lg px-2.5 py-2 text-left text-sm text-gray-700 transition-colors hover:bg-indigo-50/50 hover:text-indigo-700"
              @click="router.push('/practice/random')"
            >
              <span class="flex items-center gap-2">
                <svg class="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 12c0-1.232-.046-2.453-.138-3.662a4.006 4.006 0 00-3.7-3.7 48.678 48.678 0 00-7.324 0 4.006 4.006 0 00-3.7 3.7c-.017.22-.032.441-.046.662M19.5 12l3-3m-3 3l-3-3m-12 3c0 1.232.046 2.453.138 3.662a4.006 4.006 0 003.7 3.7 48.656 48.656 0 007.324 0 4.006 4.006 0 003.7-3.7c.017-.22.032-.441.046-.662M4.5 12l3 3m-3-3l-3 3" />
                </svg>
                随机练习
              </span>
              <svg class="h-3.5 w-3.5 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
              </svg>
            </button>
            <button
              class="flex w-full cursor-pointer items-center justify-between rounded-lg px-2.5 py-2 text-left text-sm text-gray-700 transition-colors hover:bg-indigo-50/50 hover:text-indigo-700"
              @click="router.push('/ai/practice')"
            >
              <span class="flex items-center gap-2">
                <svg class="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
                </svg>
                AI 智能出题
              </span>
              <svg class="h-3.5 w-3.5 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
              </svg>
            </button>
            <button
              class="flex w-full cursor-pointer items-center justify-between rounded-lg px-2.5 py-2 text-left text-sm text-gray-700 transition-colors hover:bg-indigo-50/50 hover:text-indigo-700"
              @click="router.push('/practice/wrong')"
            >
              <span class="flex items-center gap-2">
                <svg class="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
                </svg>
                错题练习
              </span>
              <svg class="h-3.5 w-3.5 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
              </svg>
            </button>
          </div>
        </BaseCard>
      </aside>
    </div>
  </div>
</template>
