<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getRandomQuestions } from '@/api/practice'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import type { Question } from '@/types/question'

const router = useRouter()
const auth = useAuthStore()

const subjectId = computed(() => auth.selectedSubjectId)
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

onMounted(async () => {
  if (!auth.initialized) await auth.checkAuth()
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
      <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm sm:p-7">
        <div class="mb-6 flex items-start gap-3.5">
          <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-indigo-50">
            <svg class="h-5.5 w-5.5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.77 4.769L4 5.5m0 0H4m0 0V4m10.5 15.5h-.582m-15.356-2a8.001 8.001 0 0015.356 1.731L20 18.5m0 0V19m0 0h.5m0 0V14" />
            </svg>
          </div>
          <div>
            <h2 class="text-base font-semibold tracking-tight text-gray-900">随机抽题模式</h2>
            <p class="mt-0.5 text-sm text-gray-500">系统将在当前科目（{{ auth.user?.subject_name || '未选择科目' }}）的全部题库中随机抽取题目</p>
          </div>
        </div>

        <div v-if="error" class="mb-4 rounded-xl border border-red-100 bg-red-50 px-4 py-2.5 text-sm text-red-600">{{ error }}</div>

        <div class="space-y-5">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">难度</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="opt in difficultyOptions"
                :key="opt.value"
                class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
                :class="difficulty === opt.value
                  ? 'border-indigo-500 bg-indigo-50 text-indigo-700 shadow-sm'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-indigo-300 hover:text-indigo-600'"
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
                class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
                :class="count === n
                  ? 'border-indigo-500 bg-indigo-50 text-indigo-700 shadow-sm'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-indigo-300 hover:text-indigo-600'"
                @click="count = n"
              >
                {{ n }} 题
              </button>
            </div>
          </div>
        </div>

        <div class="mt-7 flex justify-end gap-2 border-t border-gray-50 pt-5">
          <BaseButton type="secondary" @click="router.push('/subjects')">浏览科目</BaseButton>
          <BaseButton @click="start">开始练习</BaseButton>
        </div>
      </div>
    </div>

    <div v-else>
      <BaseSkeleton v-if="loading" variant="question" />
      <PracticeRunner
        v-else
        :questions="questions"
        mode="random"
        @back="reset"
      />
    </div>
  </div>
</template>
