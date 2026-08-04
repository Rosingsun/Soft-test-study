<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listExamRecords } from '@/api/exam'
import { formatDuration, formatPercent, timeAgo } from '@/utils/format'
import LineChart from '@/components/charts/LineChart.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import type { ExamRecordResp } from '@/types/exam'

const router = useRouter()
const records = ref<ExamRecordResp[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    records.value = await listExamRecords()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

const finishedRecords = computed(() => records.value.filter(r => r.status === 'finished'))

const trendData = computed(() =>
  finishedRecords.value
    .map(r => ({
      label: (r.started_at || '').slice(5),
      value: percent(r.score, r.total_score),
      hint: `正确率 ${formatPercent(r.accuracy)}`,
    }))
    .reverse(),
)

function percent(score: number, total: number) {
  if (!total) return 0
  return Math.round((score / total) * 100)
}

function accuracyClass(r: ExamRecordResp) {
  if (r.accuracy >= 80) return 'text-emerald-600'
  if (r.accuracy >= 60) return 'text-indigo-600'
  return 'text-red-500'
}

function statusBadge(r: ExamRecordResp) {
  if (r.status === 'finished') return { type: 'success' as const, label: '已完成' }
  if (r.status === 'in_progress') return { type: 'warning' as const, label: '进行中' }
  return { type: 'default' as const, label: r.status }
}

function goDetail(r: ExamRecordResp) {
  if (r.status === 'finished') {
    router.push(`/exam/${r.id}/result`)
  } else {
    router.push(`/exam/${r.template_id}`)
  }
}
</script>

<template>
  <div>
    <BasePageHeader title="考试记录" subtitle="回顾每一次模拟考试的成绩">
      <template #actions>
        <BaseButton type="secondary" size="sm" @click="router.push('/exam')">去考试</BaseButton>
      </template>
    </BasePageHeader>

    <div v-if="loading">
      <BaseLoading />
    </div>

    <div v-else-if="error" class="rounded-lg bg-white p-6 shadow-sm">
      <p class="text-red-500">{{ error }}</p>
    </div>

    <div v-else-if="records.length === 0" class="rounded-lg bg-white shadow-sm">
      <BaseEmpty title="暂无考试记录" description="完成一次模拟考试后，记录会展示在这里">
        <template #action>
          <BaseButton size="sm" class="mt-4" @click="router.push('/exam')">去参加考试</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else class="space-y-3">
      <div v-if="finishedRecords.length" class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
        <div class="mb-4 flex flex-wrap items-baseline justify-between gap-2">
          <h2 class="text-base font-semibold text-gray-900">成绩趋势</h2>
          <p class="text-xs text-gray-400">最近 {{ finishedRecords.length }} 次已完成考试 · 得分率</p>
        </div>
        <LineChart :data="trendData" :height="200" />
      </div>

      <div
        v-for="r in records"
        :key="r.id"
        class="flex cursor-pointer flex-col gap-3 rounded-lg border border-gray-200 bg-white p-5 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md sm:flex-row sm:items-center"
        @click="goDetail(r)"
      >
        <div class="min-w-0 flex-1">
          <div class="mb-1 flex flex-wrap items-center gap-2">
            <h3 class="truncate text-sm font-semibold text-gray-900">{{ r.template_name }}</h3>
            <BaseBadge :type="statusBadge(r).type">{{ statusBadge(r).label }}</BaseBadge>
          </div>
          <p class="text-xs text-gray-400">{{ timeAgo(r.created_at || r.started_at) }} · 用时 {{ formatDuration(r.duration) }}</p>
        </div>

        <template v-if="r.status === 'finished'">
          <div class="flex items-center gap-4">
            <div class="w-24">
              <div class="mb-1 flex items-center justify-between text-xs">
                <span class="text-gray-400">得分率</span>
                <span :class="percent(r.score, r.total_score) >= 60 ? 'text-emerald-600' : 'text-red-500'">{{ percent(r.score, r.total_score) }}%</span>
              </div>
              <div class="h-1.5 overflow-hidden rounded-full bg-gray-100">
                <div
                  class="h-full rounded-full"
                  :class="percent(r.score, r.total_score) >= 60 ? 'bg-emerald-500' : 'bg-red-500'"
                  :style="{ width: percent(r.score, r.total_score) + '%' }"
                />
              </div>
            </div>
            <div class="text-center">
              <p class="text-sm font-bold text-gray-900">{{ r.score }}<span class="text-xs font-normal text-gray-400"> / {{ r.total_score }}</span></p>
              <p class="mt-0.5 text-xs" :class="accuracyClass(r)">{{ formatPercent(r.accuracy) }}</p>
            </div>
            <BaseButton size="sm" type="secondary">查看详情</BaseButton>
          </div>
        </template>
        <BaseButton v-else size="sm">继续考试</BaseButton>
      </div>
    </div>
  </div>
</template>
