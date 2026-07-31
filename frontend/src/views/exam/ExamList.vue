<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubjectStore } from '@/stores/subject'
import { listExamTemplates } from '@/api/exam'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import type { ExamTemplateResp } from '@/types/exam'

const auth = useAuthStore()
const subjectStore = useSubjectStore()
const router = useRouter()
const templates = ref<ExamTemplateResp[]>([])
const loading = ref(true)
const error = ref('')
const subjectFilter = ref(0)

const filteredTemplates = computed(() => {
  if (!subjectFilter.value) return templates.value
  return templates.value.filter(t => t.subject_id === subjectFilter.value)
})

const subjectOptions = computed(() => {
  const ids = new Set(templates.value.map(t => t.subject_id))
  return subjectStore.subjects.filter(s => ids.has(s.id))
})

onMounted(async () => {
  try {
    templates.value = await listExamTemplates()
    subjectFilter.value = auth.selectedSubjectId || 0
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

function goExam(t: ExamTemplateResp) {
  router.push(`/exam/${t.id}`)
}
</script>

<template>
  <div>
    <BasePageHeader title="模拟考试" subtitle="全真模拟，检验备考成果">
      <template #actions>
        <button
          class="cursor-pointer rounded-lg bg-white border border-gray-300 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50"
          @click="router.push('/exam-records')"
        >
          考试记录
        </button>
      </template>
    </BasePageHeader>

    <div class="mb-5 flex flex-wrap items-center gap-2">
      <button
        class="cursor-pointer rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
        :class="subjectFilter === 0
          ? 'bg-indigo-600 text-white'
          : 'bg-white text-gray-600 border border-gray-300 hover:bg-gray-50'"
        @click="subjectFilter = 0"
      >
        全部科目
      </button>
      <button
        v-for="s in subjectOptions"
        :key="s.id"
        class="cursor-pointer rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
        :class="subjectFilter === s.id
          ? 'bg-indigo-600 text-white'
          : 'bg-white text-gray-600 border border-gray-300 hover:bg-gray-50'"
        @click="subjectFilter = s.id"
      >
        {{ s.short_name || s.name }}
      </button>
    </div>

    <div v-if="loading">
      <BaseLoading />
    </div>

    <div v-else-if="error" class="rounded-lg bg-white p-6 shadow-sm">
      <p class="text-red-500">{{ error }}</p>
    </div>

    <div v-else-if="filteredTemplates.length === 0" class="rounded-lg bg-white shadow-sm">
      <BaseEmpty title="暂无可用试卷" description="题目数量达标后会自动生成模拟试卷" />
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="t in filteredTemplates"
        :key="t.id"
        class="flex flex-col rounded-lg border border-gray-200 bg-white p-5 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md"
      >
        <div class="mb-3 flex items-start justify-between gap-2">
          <h3 class="text-base font-semibold text-gray-900">{{ t.name }}</h3>
          <BaseBadge v-if="t.year">{{ t.year }}年</BaseBadge>
        </div>
        <p class="mb-4 text-xs text-gray-400">
          {{ subjectStore.subjects.find(s => s.id === t.subject_id)?.name || `科目 #${t.subject_id}` }}
        </p>

        <div class="mb-4 grid grid-cols-3 gap-2 text-center">
          <div class="rounded-lg bg-gray-50 py-2">
            <p class="text-sm font-bold text-gray-800">{{ t.question_count }}</p>
            <p class="text-[10px] text-gray-400">题目</p>
          </div>
          <div class="rounded-lg bg-gray-50 py-2">
            <p class="text-sm font-bold text-gray-800">{{ t.total_score }}</p>
            <p class="text-[10px] text-gray-400">总分</p>
          </div>
          <div class="rounded-lg bg-gray-50 py-2">
            <p class="text-sm font-bold text-gray-800">{{ t.duration }}分</p>
            <p class="text-[10px] text-gray-400">时长</p>
          </div>
        </div>

        <button
          class="mt-auto cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
          @click="goExam(t)"
        >
          开始考试
        </button>
      </div>
    </div>
  </div>
</template>
