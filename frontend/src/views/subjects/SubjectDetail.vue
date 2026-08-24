<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSubjectStore } from '@/stores/subject'
import { getChapterProgress } from '@/api/stats'
import type { ChapterProgressResp } from '@/types/stats'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseProgressBar from '@/components/common/BaseProgressBar.vue'

const store = useSubjectStore()
const route = useRoute()
const router = useRouter()
const loading = ref(true)
const error = ref('')
const subSubjectName = ref('')
const progressMap = ref<Map<number, ChapterProgressResp>>(new Map())
const keyword = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const id = Number(route.params.id)
    await store.fetchChapters(id)
    const sub = store.subSubjects.find(s => s.id === id)
    if (sub) {
      subSubjectName.value = sub.name
    } else if (store.chapters.length > 0) {
      const subjectId = store.chapters[0].subject_id
      await store.fetchSubSubjects(subjectId)
      const found = store.subSubjects.find(s => s.id === id)
      if (found) {
        subSubjectName.value = found.name
      }
    }
    loadProgress()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function loadProgress() {
  try {
    const list = await getChapterProgress()
    progressMap.value = new Map(list.map(p => [p.chapter_id, p]))
  } catch {
    progressMap.value = new Map()
  }
}

const stats = computed(() => {
  const chapters = store.chapters
  const totalQuestions = chapters.reduce((sum, c) => sum + (c.question_count || 0), 0)
  let practicedChapters = 0
  let practicedTotal = 0
  let correctTotal = 0
  for (const ch of chapters) {
    const p = progressMap.value.get(ch.id)
    if (p && p.total_count > 0) {
      practicedChapters++
      practicedTotal += p.total_count
      correctTotal += p.correct_count
    }
  }
  return {
    chapterCount: chapters.length,
    totalQuestions,
    practicedChapters,
    accuracy: practicedTotal > 0 ? Math.round((correctTotal / practicedTotal) * 100) : null,
  }
})

const filteredChapters = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return store.chapters
  return store.chapters.filter(
    c => c.name.toLowerCase().includes(kw) || String(c.sort_order || '') === kw,
  )
})

function progressOf(chapterId: number): { practiced: number; accuracy: number | null; percent: number } {
  const p = progressMap.value.get(chapterId)
  if (!p || p.total_count === 0) {
    return { practiced: 0, accuracy: null, percent: 0 }
  }
  const total = Math.max(1, progressQuestionTotal(chapterId))
  return {
    practiced: p.total_count,
    accuracy: Math.round(p.accuracy),
    percent: Math.min(100, Math.round((p.total_count / total) * 100)),
  }
}

function progressQuestionTotal(chapterId: number): number {
  const ch = store.chapters.find(c => c.id === chapterId)
  return ch?.question_count || 0
}

function accuracyClass(accuracy: number | null): string {
  if (accuracy === null) return 'text-gray-400'
  if (accuracy >= 80) return 'text-emerald-600'
  if (accuracy >= 60) return 'text-amber-600'
  return 'text-red-500'
}

function goDocument(ch: { id: number }) {
  router.push(`/chapters/${ch.id}`)
}

onMounted(load)
</script>

<template>
  <div>
    <BasePageHeader
      :title="subSubjectName || '章节列表'"
      :subtitle="'选择章节开始练习，进度与正确率实时同步'"
    >
      <template #actions>
        <BaseButton type="secondary" @click="router.back()">
          ← 返回
        </BaseButton>
      </template>
    </BasePageHeader>

    <BaseSkeleton v-if="loading" variant="grid" :count="6" />

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="load">
            重试
          </BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <BaseEmpty
      v-else-if="store.chapters.length === 0"
      title="暂无章节"
      description="该子科目下还没有章节"
    />

    <template v-else>
      <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <div class="rounded-xl border border-indigo-100 bg-indigo-50/50 p-4 shadow-sm">
          <p class="text-2xl font-bold tracking-tight text-indigo-600">{{ stats.chapterCount }}</p>
          <p class="mt-1 text-xs text-gray-500">章节数</p>
        </div>
        <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ stats.totalQuestions }}</p>
          <p class="mt-1 text-xs text-gray-500">题目总数</p>
        </div>
        <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
          <p class="text-2xl font-bold tracking-tight text-gray-900">{{ stats.practicedChapters }}<span class="text-base font-medium text-gray-400"> / {{ stats.chapterCount }}</span></p>
          <p class="mt-1 text-xs text-gray-500">已练章节</p>
        </div>
        <div class="rounded-xl border border-emerald-100 bg-emerald-50/50 p-4 shadow-sm">
          <p :class="['text-2xl font-bold tracking-tight', accuracyClass(stats.accuracy)]">
            {{ stats.accuracy === null ? '—' : `${stats.accuracy}%` }}
          </p>
          <p class="mt-1 text-xs text-gray-500">综合正确率</p>
        </div>
      </div>

      <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <BaseInput v-model="keyword" placeholder="搜索章节名称或序号…" class="w-full sm:w-64" />
        <p v-if="keyword.trim()" class="text-xs text-gray-400">
          匹配 {{ filteredChapters.length }} / {{ store.chapters.length }} 个章节
        </p>
      </div>

      <BaseEmpty
        v-if="filteredChapters.length === 0"
        title="未找到匹配章节"
        description="换个关键词试试"
      />

      <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <template v-for="ch in filteredChapters" :key="ch.id">
          <BaseCard
            v-if="ch.question_count > 0"
            hover
            class="group flex h-full flex-col justify-between gap-4"
            @click="goDocument(ch)"
          >
            <div class="flex items-start gap-3">
              <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-50 text-sm font-semibold text-indigo-600">
                {{ ch.sort_order || '·' }}
              </div>
              <h3 class="line-clamp-2 text-sm font-semibold leading-5 tracking-tight text-gray-900">{{ ch.name }}</h3>
              <div class="ml-auto flex shrink-0 items-center gap-1.5">
                <span
                  v-if="ch.material_id"
                  class="flex items-center gap-1 rounded-full bg-violet-50 px-2.5 py-1 text-xs font-medium text-violet-600 ring-1 ring-inset ring-violet-600/20"
                  title="本章已绑定知识点文档"
                >
                  <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  有文档
                </span>
                <svg class="h-5 w-5 shrink-0 text-gray-300 transition-colors duration-200 group-hover:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </div>
            <div class="space-y-2.5">
              <BaseProgressBar :value="progressOf(ch.id).percent" size="sm" />
              <div class="flex items-center justify-between text-xs">
                <span class="text-gray-400">{{ ch.question_count }} 题</span>
                <span class="font-medium text-indigo-600">已练 {{ progressOf(ch.id).practiced }} / {{ ch.question_count }}</span>
                <span :class="['font-medium', accuracyClass(progressOf(ch.id).accuracy)]">
                  {{ progressOf(ch.id).accuracy === null ? '未作答' : `正确率 ${progressOf(ch.id).accuracy}%` }}
                </span>
              </div>
            </div>
          </BaseCard>
          <BaseCard
            v-else
            :hover="!!ch.material_id"
            class="flex h-full flex-col justify-between gap-4"
            :class="ch.material_id ? 'cursor-pointer' : 'opacity-60'"
            @click="ch.material_id ? goDocument(ch) : undefined"
          >
            <div class="flex items-start gap-3">
              <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-sm font-semibold text-gray-400">
                {{ ch.sort_order || '·' }}
              </div>
              <h3 class="line-clamp-2 text-sm font-semibold leading-5 tracking-tight text-gray-900">{{ ch.name }}</h3>
              <span
                v-if="ch.material_id"
                class="ml-auto flex shrink-0 items-center gap-1 rounded-full bg-violet-50 px-2.5 py-1 text-xs font-medium text-violet-600 ring-1 ring-inset ring-violet-600/20"
              >
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                有文档
              </span>
            </div>
            <div class="flex items-center justify-between text-xs">
              <span class="rounded-full bg-gray-100 px-2 py-0.5 font-medium text-gray-400 ring-1 ring-inset ring-gray-200">暂无题目</span>
              <span class="text-gray-400">{{ progressOf(ch.id).practiced > 0 ? `历史已练 ${progressOf(ch.id).practiced} 题` : '' }}</span>
            </div>
          </BaseCard>
        </template>
      </div>
    </template>
  </div>
</template>
