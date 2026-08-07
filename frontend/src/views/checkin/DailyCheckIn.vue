<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getCheckInToday, submitCheckInAnswer, completeCheckIn, getCheckInStats, resetCheckIn } from '@/api/checkin'
import type { CheckInTodayResp, CheckInStatsResp, CheckInResultResp } from '@/types/checkin'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { showToast } from '@/utils/toast'

const router = useRouter()
const loading = ref(true)
const error = ref('')
const completing = ref(false)
const today = ref<CheckInTodayResp | null>(null)
const stats = ref<CheckInStatsResp | null>(null)
const result = ref<CheckInResultResp | null>(null)
const submittedCount = ref(0)
// 重新打卡状态
const retaking = ref(false)
const retakeData = ref<{
  questions: CheckInTodayResp['questions']
  answerMap: Record<number, string>
  submittedMap: Record<number, boolean>
} | null>(null)
// 重新打卡二次确认
const retakeConfirmRef = ref<InstanceType<typeof ConfirmDialog> | null>(null)
const retakePendingMode = ref<'completed' | 'inprogress' | null>(null)

const startIndex = computed(() => {
  if (retaking.value && retakeData.value) return 0
  if (!today.value) return 0
  const idx = today.value.questions.findIndex(q => !q.answered)
  return idx < 0 ? 0 : idx
})

const finished = computed(() => !!result.value)

const runnerQuestions = computed(() => retaking.value && retakeData.value ? retakeData.value.questions : (today.value?.questions || []))
const runnerAnswers = computed(() => retaking.value && retakeData.value ? retakeData.value.answerMap : (today.value?.answer_map || {}))
const runnerSubmitted = computed(() => {
  if (retaking.value && retakeData.value) return retakeData.value.submittedMap
  if (!today.value) return {}
  return Object.fromEntries(today.value.questions.map(q => [q.id, q.answered]))
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [t, s] = await Promise.all([getCheckInToday(), getCheckInStats()])
    today.value = t
    stats.value = s
    if (t.completed && t.result) result.value = t.result
    submittedCount.value = t.answered_count
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function submitHandler(questionId: number, answer: string, duration: number) {
  const res = await submitCheckInAnswer({ question_id: questionId, answer, duration })
  submittedCount.value++
  return res
}

async function handleFinish() {
  if (completing.value) return
  completing.value = true
  const wasRetaking = retaking.value
  try {
    // 最后一题可能仍在异步上报中，先轮询确认全部落库再完成打卡
    for (let i = 0; i < 10; i++) {
      const t = await getCheckInToday()
      if (t.completed || (t.answered_count >= t.total_count)) break
      await new Promise(r => setTimeout(r, 250))
    }
    const r = await completeCheckIn()
    result.value = r
    // 重打完成：退出重打态并刷新统计与今日数据
    if (wasRetaking) {
      retaking.value = false
      retakeData.value = null
      await load()
    }
    showToast(wasRetaking ? '重新打卡完成' : '今日打卡完成', 'success')
  } catch {
    // toast 由请求层提示
  } finally {
    completing.value = false
  }
}

function startRetake() {
  if (!today.value || !today.value.retake_available) return
  retaking.value = true
  result.value = null
  retakeData.value = {
    questions: today.value.questions.map(q => ({ ...q, answered: false })),
    answerMap: {},
    submittedMap: {},
  }
}

function askRetake() {
  if (!today.value) return
  // 已完成态：直接进入重打（只重置前端快照，不丢失后端汇总）
  if (today.value.completed) {
    startRetake()
    return
  }
  // 未完成态：弹确认，告知会清空已答答案
  retakePendingMode.value = 'inprogress'
  retakeConfirmRef.value?.open()
}

async function onRetakeConfirm() {
  if (retakePendingMode.value === 'inprogress') {
    try {
      await resetCheckIn()
      await load()
      startRetake()
    } catch {
      // toast 由请求层处理
    } finally {
      retakePendingMode.value = null
    }
  }
}

function onRetakeCancel() {
  retakePendingMode.value = null
}

function cancelRetake() {
  retaking.value = false
  retakeData.value = null
  result.value = today.value?.result || null
}

function goBack() {
  router.push('/')
}

function accuracyClass(a: number) {
  if (a >= 80) return 'text-emerald-600'
  if (a >= 60) return 'text-indigo-600'
  return 'text-amber-500'
}
</script>

<template>
  <div>
    <BasePageHeader title="每日打卡" subtitle="每天 10 道真题客观题，完成即打卡，坚持就是胜利" />

    <BaseSkeleton v-if="loading" variant="question" />

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
      <p class="text-sm text-red-500">{{ error }}</p>
      <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重新加载</BaseButton>
    </div>

    <template v-else>
      <!-- 打卡统计 -->
      <div v-if="stats" class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <div class="card-lift rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-orange-50">
            <svg class="h-5 w-5 text-orange-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15.362 5.214A8.252 8.252 0 0112 21 8.25 8.25 0 016.038 7.048 8.287 8.287 0 009 9.6a8.983 8.983 0 013.361-6.867 8.21 8.21 0 003 2.48z" />
            </svg>
          </div>
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ stats.current_streak }}</p>
          <p class="mt-1 text-xs text-gray-400">当前连续打卡</p>
        </div>
        <div class="card-lift rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-50">
            <svg class="h-5 w-5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 18.75h-9m9 0a3 3 0 013 3h-15a3 3 0 013-3m9 0v-3.375c0-.621-.503-1.125-1.125-1.125h-.871M7.5 18.75v-3.375c0-.621.504-1.125 1.125-1.125h.872m5.007 0H9.497m5.007 0a7.454 7.454 0 01-.982-3.172M9.497 14.25a7.454 7.454 0 00.981-3.172M5.25 4.236c-.982.143-1.954.317-2.916.52A6.003 6.003 0 007.73 9.728M5.25 4.236V4.5c0 2.108.966 3.99 2.48 5.228M5.25 4.236V2.721C7.456 2.41 9.71 2.25 12 2.25c2.291 0 4.545.16 6.75.47v1.516M7.73 9.728a6.726 6.726 0 002.748 1.35m8.272-6.842V4.5c0 2.108-.966 3.99-2.48 5.228m2.48-5.492a46.32 46.32 0 012.916.52 6.003 6.003 0 01-5.395 4.972m0 0a6.726 6.726 0 01-2.749 1.35m0 0a6.772 6.772 0 01-3.044 0" />
            </svg>
          </div>
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ stats.max_streak }}</p>
          <p class="mt-1 text-xs text-gray-400">最高连续打卡</p>
        </div>
        <div class="card-lift rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-50">
            <svg class="h-5 w-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
            </svg>
          </div>
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ stats.total_days }}</p>
          <p class="mt-1 text-xs text-gray-400">累计打卡天数</p>
        </div>
        <div class="card-lift rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
          <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-violet-50">
            <svg class="h-5 w-5 text-violet-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <p class="text-2xl font-bold tracking-tight" :class="accuracyClass(stats.avg_accuracy)">{{ stats.avg_accuracy.toFixed(1) }}%</p>
          <p class="mt-1 text-xs text-gray-400">打卡平均正确率</p>
        </div>
      </div>

      <!-- 已完成结果 -->
      <div v-if="finished && result" class="mx-auto max-w-2xl">
        <div class="rounded-2xl border border-gray-100 bg-white p-8 text-center shadow-sm">
          <div class="bg-brand-gradient mx-auto flex h-16 w-16 items-center justify-center rounded-full text-white shadow-lg shadow-indigo-600/30">
            <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 class="mt-4 text-lg font-bold tracking-tight text-gray-900">今日打卡完成</h2>
          <p class="mt-1 text-sm text-gray-500">成绩不计入排名，可重新打卡刷新当日数据</p>
          <div class="mt-6 grid grid-cols-3 gap-3">
            <div class="rounded-xl bg-gray-50 p-4">
              <p class="text-2xl font-bold text-gray-900">{{ result.correct_count }}/{{ result.total_count }}</p>
              <p class="mt-1 text-xs text-gray-400">答对题数</p>
            </div>
            <div class="rounded-xl bg-gray-50 p-4">
              <p class="text-2xl font-bold" :class="accuracyClass(result.accuracy)">{{ result.accuracy.toFixed(1) }}%</p>
              <p class="mt-1 text-xs text-gray-400">正确率</p>
            </div>
            <div class="rounded-xl bg-gray-50 p-4">
              <p class="text-2xl font-bold text-orange-500">{{ stats?.current_streak || 0 }}</p>
              <p class="mt-1 text-xs text-gray-400">连续打卡</p>
            </div>
          </div>
          <div v-if="today?.retake_available" class="mt-6 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1.5 rounded-lg border border-indigo-200 bg-indigo-50/50 px-4 py-2 text-sm font-medium text-indigo-600 transition-all duration-200 hover:border-indigo-300 hover:bg-indigo-50 hover:shadow-sm active:scale-[0.98] focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500/40"
              @click="askRetake"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
              </svg>
              重新打卡
            </button>
            <BaseButton @click="router.push('/')">返回首页</BaseButton>
          </div>
          <div v-else class="mt-6">
            <BaseButton @click="router.push('/')">返回首页</BaseButton>
          </div>
        </div>
      </div>

      <!-- 未完成：作答 -->
      <div v-else-if="(retaking && retakeData) || (today && today.questions.length > 0)">
        <div v-if="!retaking" class="mb-4 flex items-center justify-between">
          <p class="text-xs text-gray-400">本次成绩将更新当日最新数据，不计入排名</p>
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1.5 rounded-lg border border-indigo-200 bg-indigo-50/50 px-3 py-1.5 text-xs font-medium text-indigo-600 transition-all duration-200 hover:border-indigo-300 hover:bg-indigo-50 hover:shadow-sm active:scale-[0.98] focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500/40"
            @click="askRetake"
          >
            <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
            </svg>
            重新打卡
          </button>
        </div>
        <div v-if="retaking" class="mb-4 flex items-center justify-between">
          <p class="text-sm font-medium text-gray-700">重新打卡 · 本次成绩将更新当日数据，不计入排名</p>
          <BaseButton type="ghost" size="sm" @click="cancelRetake">放弃重打</BaseButton>
        </div>
        <PracticeRunner
          :questions="runnerQuestions"
          :initial-answers="runnerAnswers"
          :initial-submitted="runnerSubmitted"
          :start-index="startIndex"
          mode="checkin"
          :title="retaking ? '每日打卡 · 重新作答' : '每日打卡 · 10 道真题'"
          :submit-handler="submitHandler"
          @finish="handleFinish"
          @back="retaking ? cancelRetake : goBack"
        />
      </div>

      <div v-else class="rounded-xl border border-gray-100 bg-white shadow-sm">
        <BaseEmpty title="暂无可用题目" description="题库暂时没有可用的客观题">
          <template #action>
            <BaseButton type="secondary" size="sm" class="mt-4" @click="goBack">返回首页</BaseButton>
          </template>
        </BaseEmpty>
      </div>
    </template>

    <!-- 未完成态重新打卡二次确认 -->
    <ConfirmDialog
      ref="retakeConfirmRef"
      title="重新打卡"
      description="将清空今日已作答的题目，重新开始作答。本次成绩将更新当日最新数据，不计入排名。"
      confirm-text="清空并重打"
      @confirm="onRetakeConfirm"
      @cancel="onRetakeCancel"
    />
  </div>
</template>
