<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubjectStore } from '@/stores/subject'
import { getSpecialQuestions } from '@/api/practice'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import { typeLabel } from '@/utils/question'
import type { Question } from '@/types/question'

const router = useRouter()
const auth = useAuthStore()
const subjectStore = useSubjectStore()

const subjectId = ref(0)
const type = ref('single')
const difficulty = ref('')
const count = ref(10)
const questions = ref<Question[]>([])
const started = ref(false)
const loading = ref(false)
const error = ref('')

const typeOptions = ['single', 'multi', 'judge', 'fill', 'short', 'comprehensive']
const difficultyOptions = [
  { value: '', label: '全部难度' },
  { value: 'easy', label: '简单' },
  { value: 'medium', label: '中等' },
  { value: 'hard', label: '困难' },
]

const subjectsForLevel = computed(() => subjectStore.subjects.filter(s => s.level_id === auth.selectedLevelId))

onMounted(async () => {
  if (!auth.initialized) await auth.checkAuth()
  if (subjectStore.levels.length === 0) await subjectStore.fetchLevels()
  if (auth.selectedLevelId) await subjectStore.ensureSubjects(auth.selectedLevelId)
  subjectId.value = auth.selectedSubjectId || subjectsForLevel.value[0]?.id || 0
})

async function start() {
  if (!subjectId.value) {
    error.value = '请先选择科目'
    return
  }
  started.value = true
  loading.value = true
  error.value = ''
  try {
    questions.value = await getSpecialQuestions(subjectId.value, type.value, difficulty.value || undefined, count.value)
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
      <div class="rounded-lg bg-white p-6 shadow-sm">
        <div class="mb-5 flex items-start gap-3">
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-emerald-50">
            <svg class="h-5 w-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          <div>
            <h2 class="text-base font-semibold text-gray-900">题型专项训练</h2>
            <p class="mt-0.5 text-sm text-gray-500">选择科目与题型，按需强化训练</p>
          </div>
        </div>

        <div v-if="error" class="mb-4 rounded-lg bg-red-50 px-4 py-2 text-sm text-red-600">{{ error }}</div>

        <div class="space-y-4">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">选择科目</label>
            <select
              v-model="subjectId"
              class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
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
                class="cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors"
                :class="type === t
                  ? 'border-emerald-500 bg-emerald-50 text-emerald-600'
                  : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
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
                class="cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors"
                :class="difficulty === opt.value
                  ? 'border-emerald-500 bg-emerald-50 text-emerald-600'
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
                v-for="n in [5, 10, 15, 20]"
                :key="n"
                class="cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors"
                :class="count === n
                  ? 'border-emerald-500 bg-emerald-50 text-emerald-600'
                  : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
                @click="count = n"
              >
                {{ n }} 题
              </button>
            </div>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-2">
          <BaseButton type="secondary" @click="router.push('/practice/random')">随机练习</BaseButton>
          <BaseButton type="success" @click="start">开始训练</BaseButton>
        </div>
      </div>
    </div>

    <div v-else>
      <div v-if="loading">
        <BaseLoading />
      </div>
      <PracticeRunner
        v-else
        :questions="questions"
        mode="special"
        @back="reset"
      />
    </div>
  </div>
</template>
