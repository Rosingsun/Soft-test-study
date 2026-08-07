<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getReviewToday, submitReviewAnswer, getReviewOverview } from '@/api/review'
import type { ReviewTodayResp, ReviewOverviewResp } from '@/types/review'
import type { Question } from '@/types/question'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import { showToast } from '@/utils/toast'

const router = useRouter()
const loading = ref(true)
const error = ref('')
const today = ref<ReviewTodayResp | null>(null)
const overview = ref<ReviewOverviewResp | null>(null)
const reviewed = ref(0)

const questions = computed<Question[]>(() => (today.value?.questions || []))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [t, o] = await Promise.all([getReviewToday(), getReviewOverview()])
    today.value = t
    overview.value = o
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function submitHandler(questionId: number, answer: string, duration: number) {
  const res = await submitReviewAnswer({ question_id: questionId, answer, duration })
  reviewed.value++
  if (res.mastered) {
    showToast('已掌握，移出错题本', 'success')
  }
  return res
}

async function handleFinish() {
  overview.value = await getReviewOverview()
  showToast('今日复习完成', 'success')
}

function goBack() {
  router.push('/wrong-questions')
}

function fmtCount(n: number) {
  return `${n} 道`
}
</script>

<template>
  <div>
    <BasePageHeader title="今日复习" subtitle="基于遗忘曲线（SM-2）智能安排错题复习节奏" />

    <BaseSkeleton v-if="loading" variant="question" />

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
      <p class="text-sm text-red-500">{{ error }}</p>
      <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重新加载</BaseButton>
    </div>

    <template v-else>
      <!-- 复习概览 -->
      <div v-if="overview" class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-5">
        <div class="rounded-xl border border-red-100 bg-red-50/60 p-4 shadow-sm">
          <p class="text-2xl font-bold tracking-tight text-red-500">{{ overview.due_today }}</p>
          <p class="mt-1 text-xs text-gray-500">今日待复习</p>
        </div>
        <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ overview.due_tomorrow }}</p>
          <p class="mt-1 text-xs text-gray-400">明日待复习</p>
        </div>
        <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
          <p class="text-2xl font-bold tracking-tight text-indigo-600">{{ overview.reviewing }}</p>
          <p class="mt-1 text-xs text-gray-400">复习中</p>
        </div>
        <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
          <p class="text-2xl font-bold tracking-tight text-emerald-600">{{ overview.mastered }}</p>
          <p class="mt-1 text-xs text-gray-400">已掌握</p>
        </div>
        <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ overview.total_reviews }}</p>
          <p class="mt-1 text-xs text-gray-400">累计复习</p>
        </div>
      </div>

      <div v-if="questions.length > 0">
        <div class="mb-4 flex items-center gap-2 rounded-xl border border-amber-100 bg-amber-50/70 px-4 py-3 text-sm text-amber-700">
          <svg class="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5" />
          </svg>
          共 {{ fmtCount(questions.length) }} 道待复习题目，答对会按遗忘曲线延长下次复习间隔
        </div>
        <PracticeRunner
          :questions="questions"
          mode="review"
          title="今日复习"
          :submit-handler="submitHandler"
          @finish="handleFinish"
          @back="goBack"
        />
      </div>

      <div v-else class="rounded-xl border border-gray-100 bg-white shadow-sm">
        <BaseEmpty title="今日没有待复习题目" :description="overview && overview.reviewing > 0 ? '全部完成，继续保持' : '答错的题目会自动进入复习队列'">
          <template #action>
            <BaseButton type="secondary" size="sm" class="mt-4" @click="router.push('/practice/random')">
              去刷题
            </BaseButton>
          </template>
        </BaseEmpty>
      </div>
    </template>
  </div>
</template>
