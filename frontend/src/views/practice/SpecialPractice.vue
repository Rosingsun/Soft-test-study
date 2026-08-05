<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubjectStore } from '@/stores/subject'
import { getSpecialQuestions, getEssayQuestions, generateQuestions } from '@/api/practice'
import { useAiStore } from '@/stores/ai'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import { typeLabel } from '@/utils/question'
import type { Question } from '@/types/question'

const router = useRouter()
const auth = useAuthStore()
const subjectStore = useSubjectStore()
const aiStore = useAiStore()

const subjectId = ref(0)
const type = ref('single')
const difficulty = ref('')
const count = ref(10)
const questions = ref<Question[]>([])
const started = ref(false)
const loading = ref(false)
const error = ref('')
const essayTopics = ref<Question[]>([])
const selectedEssayId = ref(0)
const essayMode = ref<'fixed' | 'random' | 'ai'>('fixed')

const typeOptions = ['single', 'multi', 'judge', 'fill', 'short', 'comprehensive', 'case_study', 'essay']
const difficultyOptions = [
  { value: '', label: '全部难度' },
  { value: 'easy', label: '简单' },
  { value: 'medium', label: '中等' },
  { value: 'hard', label: '困难' },
]

const subjectsForLevel = computed(() => subjectStore.subjects.filter(s => s.level_id === auth.selectedLevelId))
const isEssay = computed(() => type.value === 'essay')
const showCount = computed(() => !isEssay.value)

onMounted(async () => {
  if (!auth.initialized) await auth.checkAuth()
  if (subjectStore.levels.length === 0) await subjectStore.fetchLevels()
  if (auth.selectedLevelId) await subjectStore.ensureSubjects(auth.selectedLevelId)
  subjectId.value = auth.selectedSubjectId || subjectsForLevel.value[0]?.id || 0
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
      // 后端返回 { questions: [...] }
      questions.value = (resp.questions || []).map((q: any) => ({
        id: q.id,
        subject_id: subjectId.value,
        sub_subject_id: 0,
        chapter_id: 0,
        type: q.type,
        difficulty: q.difficulty,
        content: q.content,
        options: q.options || '[]',
        answer: q.answer || '',
        analysis: q.analysis || '',
        year: q.year || 0,
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
</script>

<template>
  <div>
    <BasePageHeader title="专项练习" subtitle="针对特定题型集中训练，逐个突破薄弱环节" />

    <!-- 配置面板 -->
    <div v-if="!started" class="mx-auto max-w-2xl">
      <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm sm:p-7">
        <div class="mb-6 flex items-start gap-3.5">
          <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-50">
            <svg class="h-5.5 w-5.5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          <div>
            <h2 class="text-base font-semibold tracking-tight text-gray-900">题型专项训练</h2>
            <p class="mt-0.5 text-sm text-gray-500">选择科目与题型，按需强化训练</p>
          </div>
        </div>

        <div v-if="error" class="mb-4 rounded-xl border border-red-100 bg-red-50 px-4 py-2.5 text-sm text-red-600">{{ error }}</div>

        <div class="space-y-5">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">选择科目</label>
            <select
              v-model="subjectId"
              class="w-full cursor-pointer rounded-lg border border-gray-300 bg-white px-3.5 py-2.5 text-sm shadow-sm transition-colors hover:border-gray-400 focus:border-emerald-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
            >
              <option v-for="s in subjectsForLevel" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">题型</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="t in typeOptions"
                :key="t"
                class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
                :class="type === t
                  ? 'border-emerald-500 bg-emerald-50 text-emerald-700 shadow-sm'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-300 hover:text-emerald-600'"
                @click="type = t"
              >
                {{ typeLabel(t) }}
              </button>
            </div>
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">难度</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="opt in difficultyOptions"
                :key="opt.value"
                class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
                :class="difficulty === opt.value
                  ? 'border-emerald-500 bg-emerald-50 text-emerald-700 shadow-sm'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-300 hover:text-emerald-600'"
                @click="difficulty = opt.value"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>
          <div v-if="showCount">
            <label class="mb-1.5 block text-sm font-medium text-gray-700">题目数量</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="n in [5, 10, 15, 20]"
                :key="n"
                class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
                :class="count === n
                  ? 'border-emerald-500 bg-emerald-50 text-emerald-700 shadow-sm'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-300 hover:text-emerald-600'"
                @click="count = n"
              >
                {{ n }} 题
              </button>
            </div>
          </div>
          <div v-if="isEssay">
            <label class="mb-1.5 block text-sm font-medium text-gray-700">论文题目</label>
            <div class="flex flex-col gap-3">
              <div class="flex flex-wrap gap-2">
                <button
                  class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
                  :class="essayMode === 'fixed'
                    ? 'border-emerald-500 bg-emerald-50 text-emerald-700 shadow-sm'
                    : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-300 hover:text-emerald-600'"
                  @click="essayMode = 'fixed'"
                >
                  固定题目
                </button>
                <button
                  class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
                  :class="essayMode === 'random'
                    ? 'border-emerald-500 bg-emerald-50 text-emerald-700 shadow-sm'
                    : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-300 hover:text-emerald-600'"
                  @click="essayMode = 'random'; resetEssayModeIfNeeded()"
                >
                  随机
                </button>
                <button
                  class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
                  :class="essayMode === 'ai'
                    ? 'border-emerald-500 bg-emerald-50 text-emerald-700 shadow-sm'
                    : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-300 hover:text-emerald-600'"
                  @click="essayMode = 'ai'; resetEssayModeIfNeeded()"
                >
                  AI生成
                </button>
              </div>
              <select
                v-if="essayMode === 'fixed'"
                v-model="selectedEssayId"
                class="w-full cursor-pointer rounded-lg border border-gray-300 bg-white px-3.5 py-2.5 text-sm shadow-sm transition-colors hover:border-gray-400 focus:border-emerald-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
              >
                <option value="0">请选择论文题目</option>
                <option v-for="topic in essayTopics" :key="topic.id" :value="topic.id">
                  {{ topic.year }} {{ topic.content }}
                </option>
              </select>
              <p v-if="essayMode !== 'fixed'" class="text-sm text-gray-500">选择随机或AI生成后，将自动抽取/生成一题，当前论文题目选择将被清空。</p>
            </div>
          </div>
        </div>

        <div class="mt-7 flex justify-end gap-2 border-t border-gray-50 pt-5">
          <BaseButton type="secondary" @click="router.push('/practice/random')">随机练习</BaseButton>
          <BaseButton type="success" @click="start">开始训练</BaseButton>
        </div>
      </div>
    </div>

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
</template>
