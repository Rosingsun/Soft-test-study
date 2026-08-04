<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getExamResult } from '@/api/exam'
import { essayScore, checkEssayScore } from '@/api/ai'
import { useAiStore } from '@/stores/ai'
import { formatDuration, formatPercent } from '@/utils/format'
import { typeLabel } from '@/utils/question'
import DonutChart from '@/components/charts/DonutChart.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import type { ExamResultResp, SectionAccuracyResp } from '@/types/exam'
import type { EssayScoreResp } from '@/types/ai'

const route = useRoute()
const router = useRouter()
const aiStore = useAiStore()
const result = ref<ExamResultResp | null>(null)
const loading = ref(true)
const error = ref('')

// AI 评分状态
const aiScoring = reactive<Record<number, boolean>>({})
const aiScoreResults = reactive<Record<number, EssayScoreResp>>({})
const aiScoreError = reactive<Record<number, string>>({})

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

// 论文 AI 评分
async function handleExamAiScore(detail: { question_id: number; exam_answer_id: number; your_answer: string; type: string }) {
  if (detail.type !== 'essay' || !detail.your_answer) return

  if (!aiStore.hasConfig) {
    if (confirm('尚未配置 AI API Key，是否前往配置？')) {
      router.push('/ai/config')
    }
    return
  }

  const qid = detail.question_id
  const eaid = detail.exam_answer_id

  // 检查是否已有评分
  try {
    const existing = await checkEssayScore({ record_type: 'exam', exam_answer_id: eaid })
    if (existing.has_score && existing.score) {
      if (!confirm('AI 已对此论文评过分，是否还需要再次测评？')) {
        aiScoreResults[qid] = existing.score
        return
      }
    }
  } catch {
    // 忽略检查错误，继续评分
  }

  aiScoring[qid] = true
  aiScoreError[qid] = ''

  try {
    const res = await essayScore({
      api_config: aiStore.getConfig(),
      question_id: detail.question_id,
      user_answer: detail.your_answer,
      record_type: 'exam',
      record_id: result.value!.record_id,
      exam_answer_id: eaid,
    })
    aiScoreResults[qid] = res
  } catch (e) {
    aiScoreError[qid] = (e as Error).message
  } finally {
    aiScoring[qid] = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-4xl">
    <div v-if="loading">
      <BaseLoading />
    </div>

    <div v-else-if="error" class="rounded-xl border border-red-100 bg-red-50 p-6">
      <p class="text-sm text-red-600">{{ error }}</p>
      <BaseButton type="secondary" size="sm" class="mt-4" @click="router.push('/exam')">返回试卷列表</BaseButton>
    </div>

    <template v-else-if="result">
      <!-- 成绩概览 -->
      <div class="mb-6 overflow-hidden rounded-2xl border border-gray-100 bg-white shadow-sm">
        <div class="bg-brand-gradient relative overflow-hidden px-6 py-6">
          <div class="absolute -right-8 -top-8 h-32 w-32 rounded-full bg-white/10" />
          <div class="absolute -bottom-10 right-20 h-28 w-28 rounded-full bg-white/10" />
          <div class="relative flex items-center justify-between">
            <div>
              <h1 class="text-lg font-bold tracking-tight text-white">{{ result.template_name }}</h1>
              <p class="mt-1 text-xs text-indigo-100">{{ result.started_at }} ~ {{ result.finished_at }}</p>
            </div>
            <span class="rounded-full bg-white/20 px-3.5 py-1 text-sm font-bold text-white ring-1 ring-inset ring-white/30">{{ grade.label }}</span>
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
          <div class="grid flex-1 grid-cols-2 gap-3 sm:grid-cols-4">
            <div class="rounded-xl bg-gray-50 p-3.5 text-center">
              <p class="text-lg font-bold text-gray-900">{{ result.score }}<span class="text-xs text-gray-400"> / {{ result.total_score }}</span></p>
              <p class="mt-0.5 text-xs text-gray-400">得分</p>
            </div>
            <div class="rounded-xl bg-gray-50 p-3.5 text-center">
              <p class="text-lg font-bold text-emerald-600">{{ formatPercent(result.accuracy) }}</p>
              <p class="mt-0.5 text-xs text-gray-400">正确率</p>
            </div>
            <div class="rounded-xl bg-gray-50 p-3.5 text-center">
              <p class="text-lg font-bold text-gray-900">{{ result.correct_count }}</p>
              <p class="mt-0.5 text-xs text-gray-400">正确 / {{ result.total_count }}</p>
            </div>
            <div class="rounded-xl bg-gray-50 p-3.5 text-center">
              <p class="text-lg font-bold text-gray-900">{{ formatDuration(result.duration) }}</p>
              <p class="mt-0.5 text-xs text-gray-400">用时</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 分题型正确率 -->
      <div v-if="result.section_accuracy && result.section_accuracy.length" class="mb-6 rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
        <h2 class="mb-4 text-base font-semibold tracking-tight text-gray-900">题型得分率分析</h2>
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
      <h2 class="mb-3 text-base font-semibold tracking-tight text-gray-900">答题详情</h2>
      <div class="space-y-3">
        <details
          v-for="d in result.answer_details"
          :key="d.question_id"
          class="group rounded-xl border border-gray-100 bg-white shadow-sm transition-shadow hover:shadow-md"
        >
          <summary class="flex cursor-pointer items-center gap-2.5 px-4 py-3.5">
            <span class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-[10px] font-bold text-white" :class="d.type === 'essay' ? 'bg-indigo-400' : (d.is_correct ? 'bg-emerald-500' : 'bg-red-500')">
              {{ d.type === 'essay' ? '✓' : (d.is_correct ? '✓' : '✗') }}
            </span>
            <span class="min-w-0 flex-1 truncate text-sm text-gray-700">{{ stripHtml(d.content) }}</span>
            <span class="shrink-0 text-xs text-gray-400">{{ d.type === 'essay' ? '已提交' : d.score + ' 分' }}</span>
            <svg class="h-4 w-4 shrink-0 text-gray-400 transition-transform group-open:rotate-180" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
            </svg>
          </summary>
          <div class="border-t border-gray-100 px-4 py-3 text-sm">
            <p class="mb-1 text-gray-600">
              <template v-if="d.type === 'essay'">你的作答：</template>
              <template v-else>你的答案：</template>
              <span :class="d.type === 'essay' ? 'text-indigo-600' : (d.is_correct ? 'text-emerald-600' : 'text-red-600')">{{ d.your_answer || '未作答' }}</span>
            </p>
            <p v-if="d.type !== 'essay' && !d.is_correct" class="mb-1 text-gray-600">
              正确答案：<span class="font-medium text-emerald-600">{{ d.correct_answer }}</span>
            </p>
            <p v-if="d.type !== 'essay' && d.analysis" class="text-gray-500"><span class="font-medium text-gray-700">解析：</span>{{ stripHtml(d.analysis) }}</p>

            <!-- 论文 AI 评分 -->
            <div v-if="d.type === 'essay' && d.your_answer" class="mt-3 border-t border-gray-200 pt-3">
              <button
                v-if="!aiScoring[d.question_id] && !aiScoreResults[d.question_id]"
                class="cursor-pointer rounded-lg bg-indigo-600 px-4 py-1.5 text-sm text-white hover:bg-indigo-700 transition-colors"
                @click="handleExamAiScore(d)"
              >
                AI 评分
              </button>
              <span v-if="aiScoring[d.question_id]" class="text-sm text-indigo-600">AI 正在评分中...</span>
              <div v-if="aiScoreError[d.question_id]" class="mt-2 text-sm text-red-500">{{ aiScoreError[d.question_id] }}</div>

              <div v-if="aiScoreResults[d.question_id]" class="mt-3 space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-sm font-medium text-gray-700">AI 评分结果</span>
                  <button
                    class="cursor-pointer text-xs text-indigo-600 hover:text-indigo-800"
                    @click="handleExamAiScore(d)"
                    :disabled="aiScoring[d.question_id]"
                  >
                    {{ aiScoring[d.question_id] ? '评分中...' : '重新评分' }}
                  </button>
                </div>
                <!-- 总分 -->
                <div class="flex items-center gap-3">
                  <span class="text-2xl font-bold text-indigo-600">{{ aiScoreResults[d.question_id].total_score }}</span>
                  <span class="text-xs text-gray-400">/ 75 分</span>
                  <div class="h-2 flex-1 overflow-hidden rounded-full bg-gray-200">
                    <div
                      class="h-full rounded-full bg-indigo-500 transition-all duration-500"
                      :style="{ width: (aiScoreResults[d.question_id].total_score / 75 * 100) + '%' }"
                    />
                  </div>
                </div>
                <!-- 分维度评分 -->
                <div class="grid grid-cols-2 gap-2">
                  <div class="rounded-lg bg-gray-50 p-2">
                    <div class="flex items-center justify-between text-xs">
                      <span class="text-gray-500">论点与立意</span>
                      <span class="font-medium text-indigo-600">{{ aiScoreResults[d.question_id].argument_score }}/20</span>
                    </div>
                    <div class="mt-1 h-1.5 overflow-hidden rounded-full bg-gray-200">
                      <div class="h-full rounded-full bg-indigo-400" :style="{ width: (aiScoreResults[d.question_id].argument_score / 20 * 100) + '%' }" />
                    </div>
                  </div>
                  <div class="rounded-lg bg-gray-50 p-2">
                    <div class="flex items-center justify-between text-xs">
                      <span class="text-gray-500">结构与逻辑</span>
                      <span class="font-medium text-indigo-600">{{ aiScoreResults[d.question_id].structure_score }}/20</span>
                    </div>
                    <div class="mt-1 h-1.5 overflow-hidden rounded-full bg-gray-200">
                      <div class="h-full rounded-full bg-indigo-400" :style="{ width: (aiScoreResults[d.question_id].structure_score / 20 * 100) + '%' }" />
                    </div>
                  </div>
                  <div class="rounded-lg bg-gray-50 p-2">
                    <div class="flex items-center justify-between text-xs">
                      <span class="text-gray-500">语言表达</span>
                      <span class="font-medium text-indigo-600">{{ aiScoreResults[d.question_id].language_score }}/20</span>
                    </div>
                    <div class="mt-1 h-1.5 overflow-hidden rounded-full bg-gray-200">
                      <div class="h-full rounded-full bg-indigo-400" :style="{ width: (aiScoreResults[d.question_id].language_score / 20 * 100) + '%' }" />
                    </div>
                  </div>
                  <div class="rounded-lg bg-gray-50 p-2">
                    <div class="flex items-center justify-between text-xs">
                      <span class="text-gray-500">深度与广度</span>
                      <span class="font-medium text-indigo-600">{{ aiScoreResults[d.question_id].depth_score }}/15</span>
                    </div>
                    <div class="mt-1 h-1.5 overflow-hidden rounded-full bg-gray-200">
                      <div class="h-full rounded-full bg-indigo-400" :style="{ width: (aiScoreResults[d.question_id].depth_score / 15 * 100) + '%' }" />
                    </div>
                  </div>
                </div>
                <!-- 评语 -->
                <div class="rounded-lg bg-gray-50 p-3">
                  <p class="text-xs font-medium text-gray-500 mb-1">AI 评语</p>
                  <p class="text-sm text-gray-700 leading-relaxed whitespace-pre-wrap">{{ aiScoreResults[d.question_id].comment }}</p>
                </div>
              </div>
            </div>
          </div>
        </details>
      </div>
    </template>
  </div>
</template>
