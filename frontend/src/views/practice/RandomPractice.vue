<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubjectStore } from '@/stores/subject'
import { getRandomQuestions } from '@/api/practice'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import type { Question } from '@/types/question'

const router = useRouter()
const auth = useAuthStore()
const subjectStore = useSubjectStore()

const subjectId = ref(0)
const difficulty = ref('')
const count = ref(10)
const questions = ref<Question[]>([])
const started = ref(false)
const loading = ref(false)
const error = ref('')

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
    questions.value = await getRandomQuestions(subjectId.value, count.value, difficulty.value || undefined)
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
    <BasePageHeader title="随机练习" subtitle="从题库中随机抽题，检测综合掌握情况" />

    <!-- 配置面板 -->
    <div v-if="!started" class="mx-auto max-w-2xl">
      <div class="rounded-lg bg-white p-6 shadow-sm">
        <div class="mb-5 flex items-start gap-3">
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-indigo-50">
            <svg class="h-5 w-5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.77 4.769L4 5.5m0 0H4m0 0V4m10.5 15.5h-.582m-15.356-2a8.001 8.001 0 0015.356 1.731L20 18.5m0 0V19m0 0h.5m0 0V14" />
            </svg>
          </div>
          <div>
            <h2 class="text-base font-semibold text-gray-900">随机抽题模式</h2>
            <p class="mt-0.5 text-sm text-gray-500">系统将在所选科目的全部题库中随机抽取题目</p>
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
                v-for="n in [5, 10, 15, 20]"
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

        <div class="mt-6 flex justify-end gap-2">
          <BaseButton type="secondary" @click="router.push('/subjects')">浏览科目</BaseButton>
          <BaseButton @click="start">开始练习</BaseButton>
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
        mode="random"
        @back="reset"
      />
    </div>
  </div>
</template>
