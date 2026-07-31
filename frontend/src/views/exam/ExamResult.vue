<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getExamResult } from '@/api/exam'
import { formatDuration, formatPercent } from '@/utils/format'
import { typeLabel } from '@/utils/question'
import DonutChart from '@/components/charts/DonutChart.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import type { ExamResultResp, SectionAccuracyResp } from '@/types/exam'

const route = useRoute()
const router = useRouter()
const result = ref<ExamResultResp | null>(null)
const loading = ref(true)
const error = ref('')

const percentage = computed(() => {
  if (!result.value || !result.value.total_score) return 0
  return Math.round((result.value.score / result.value.total_score) * 100)
})

const grade = computed(() => {
  const p = percentage.value
  if (p >= 85) return { label: '优秀', color: 'text-emerald-500' }
  if (p >= 70) return { label: '良好', color: 'text-indigo-600' }
  if (p >= 60) return { label: '及格', color: 'text-yellow-500' }
  return { label: '加油', color: 'text-red-500' }
})

const ringColor = computed(() => {
  const p = percentage.value
  if (p >= 85) return '#10b981'
  if (p >= 60) return '#4f46e5'
  return '#ef4444'
})

onMounted(async () => {
  try {
    const id = Number(route.params.id)
    result.value = await getExamResult(id)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

function sectionAccuracyClass(s: SectionAccuracyResp) {
  if (s.accuracy >= 80) return 'bg-emerald-500'
  if (s.accuracy >= 60) return 'bg-indigo-500'
  return 'bg-red-500'
}

function stripHtml(html: string) {
  return html.replace(/<[^>]*>/g, '')
}
</script>

<template>
  <div class="mx-auto max-w-4xl">
    <div v-if="loading">
      <BaseLoading />
    </div>

    <div v-else-if="error" class="rounded-lg bg-white p-6 shadow-sm">
      <p class="text-red-500">{{ error }}</p>
      <BaseButton type="secondary" size="sm" class="mt-4" @click="router.push('/exam')">返回试卷列表</BaseButton>
    </div>

    <template v-else-if="result">
      <!-- 成绩概览 -->
      <div class="mb-6 overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
        <div class="bg-gradient-to-r from-indigo-600 to-indigo-500 px-6 py-5">
          <div class="flex items-center justify-between">
            <div>
              <h1 class="text-lg font-bold text-white">{{ result.template_name }}</h1>
              <p class="mt-1 text-xs text-indigo-200">{{ result.started_at }} ~ {{ result.finished_at }}</p>
            </div>
            <span class="rounded-full bg-white/15 px-3 py-1 text-sm font-bold text-white">{{ grade.label }}</span>
          </div>
        </div>

        <div class="flex flex-col items-center gap-6 p-6 sm:flex-row sm:items-center">
          <DonutChart
            :percentage="percentage"
            :size="140"
            :stroke="14"
            :color="ringColor"
            label="得分率"
          />
          <div class="grid flex-1 grid-cols-3 gap-3">
            <div class="rounded-lg bg-gray-50 p-3 text-center">
              <p class="text-lg font-bold text-gray-900">{{ result.score }}<span class="text-xs text-gray-400"> / {{ result.total_score }}</span></p>
              <p class="mt-0.5 text-xs text-gray-400">得分</p>
            </div>
            <div class="rounded-lg bg-gray-50 p-3 text-center">
              <p class="text-lg font-bold text-emerald-600">{{ result.correct_count }}</p>
              <p class="mt-0.5 text-xs text-gray-400">正确 / {{ result.total_count }}</p>
            </div>
            <div class="rounded-lg bg-gray-50 p-3 text-center">
              <p class="text-lg font-bold text-gray-900">{{ formatDuration(result.duration) }}</p>
              <p class="mt-0.5 text-xs text-gray-400">用时</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 分题型正确率 -->
      <div v-if="result.section_accuracy && result.section_accuracy.length" class="mb-6 rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
        <h2 class="mb-4 text-base font-semibold text-gray-900">题型得分率分析</h2>
        <div class="space-y-3">
          <div v-for="s in result.section_accuracy" :key="s.type">
            <div class="mb-1 flex items-center justify-between text-sm">
              <span class="text-gray-700">{{ typeLabel(s.type) }}</span>
              <span class="text-gray-400">{{ s.correct_count }}/{{ s.total_count }} · <strong :class="s.accuracy >= 60 ? 'text-emerald-600' : 'text-red-500'">{{ formatPercent(s.accuracy) }}</strong></span>
            </div>
            <div class="h-2 overflow-hidden rounded-full bg-gray-100">
              <div
                class="h-full rounded-full transition-all duration-700"
                :class="sectionAccuracyClass(s)"
                :style="{ width: Math.min(100, s.accuracy) + '%' }"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 操作 -->
      <div class="mb-6 flex flex-wrap gap-2">
        <BaseButton @click="router.push('/exam')">返回试卷列表</BaseButton>
        <BaseButton type="secondary" @click="router.push('/exam-records')">查看考试记录</BaseButton>
        <BaseButton type="secondary" @click="router.push('/practice/wrong')">错题重练</BaseButton>
      </div>

      <!-- 答题详情 -->
      <h2 class="mb-3 text-base font-semibold text-gray-900">答题详情</h2>
      <div class="space-y-3">
        <details
          v-for="d in result.answer_details"
          :key="d.question_id"
          class="group rounded-lg border border-gray-200 bg-white shadow-sm"
        >
          <summary class="flex cursor-pointer items-center gap-2 px-4 py-3">
            <span class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-[10px] font-bold text-white" :class="d.is_correct ? 'bg-emerald-500' : 'bg-red-500'">
              {{ d.is_correct ? '✓' : '✗' }}
            </span>
            <span class="min-w-0 flex-1 truncate text-sm text-gray-700">{{ stripHtml(d.content) }}</span>
            <span class="shrink-0 text-xs text-gray-400">{{ d.score }} 分</span>
            <svg class="h-4 w-4 shrink-0 text-gray-400 transition-transform group-open:rotate-180" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
            </svg>
          </summary>
          <div class="border-t border-gray-100 px-4 py-3 text-sm">
            <p class="mb-1 text-gray-600">
              你的答案：<span :class="d.is_correct ? 'text-emerald-600' : 'text-red-600'">{{ d.your_answer || '未作答' }}</span>
            </p>
            <p v-if="!d.is_correct" class="mb-1 text-gray-600">
              正确答案：<span class="font-medium text-emerald-600">{{ d.correct_answer }}</span>
            </p>
            <p v-if="d.analysis" class="text-gray-500"><span class="font-medium text-gray-700">解析：</span>{{ stripHtml(d.analysis) }}</p>
          </div>
        </details>
      </div>
    </template>
  </div>
</template>
