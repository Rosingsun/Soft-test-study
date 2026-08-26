<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSubjectStore } from '@/stores/subject'
import { getMaterial, getMaterialContent } from '@/api/material'
import { useAuthStore } from '@/stores/auth'
import { useAiStore } from '@/stores/ai'
import type { StudyMaterialResp, MindMapNode } from '@/types/material'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import ChapterDocNode from '@/components/knowledge/ChapterDocNode.vue'
import ChapterAiGenerateModal from '@/components/knowledge/ChapterAiGenerateModal.vue'

const route = useRoute()
const router = useRouter()
const subjectStore = useSubjectStore()
const auth = useAuthStore()
const aiStore = useAiStore()

const chapterId = computed(() => Number(route.params.id))

const chapter = computed(() => subjectStore.chapters.find(c => c.id === chapterId.value) || null)

const loading = ref(true)
const error = ref('')
const material = ref<StudyMaterialResp | null>(null)
const tree = ref<MindMapNode[]>([])
const keyword = ref('')
const expandMode = ref<'all' | 'none' | null>(null)
const expandKey = ref(0)

function countNodes(nodes: MindMapNode[]): number {
  let total = 0
  for (const n of nodes) {
    total += 1
    if (n.children?.length) total += countNodes(n.children)
  }
  return total
}

const nodeCount = computed(() => countNodes(tree.value))

const hasQuestions = computed(() => !!chapter.value && chapter.value.question_count > 0)

async function ensureChapter() {
  if (subjectStore.chapters.find(c => c.id === chapterId.value)) return
  if (!auth.initialized) await auth.checkAuth()
  const subjectId = auth.selectedSubjectId
  if (!subjectId) return
  await subjectStore.fetchSubSubjects(subjectId)
  for (const ss of subjectStore.subSubjects) {
    await subjectStore.fetchChapters(ss.id)
    if (subjectStore.chapters.find(c => c.id === chapterId.value)) return
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    await ensureChapter()
    const ch = chapter.value
    if (!ch) {
      error.value = '章节不存在或已被删除'
      return
    }
    if (ch.material_id && ch.material_id > 0) {
      const [m, nodes] = await Promise.all([
        getMaterial(ch.material_id).catch(() => null),
        getMaterialContent(ch.material_id),
      ])
      material.value = m
      tree.value = nodes
    }
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function setExpandMode(mode: 'all' | 'none') {
  if (expandMode.value === mode) {
    expandMode.value = null
  } else {
    expandMode.value = mode
  }
  expandKey.value++
}

function goBack() {
  if (chapter.value) {
    router.push(`/sub-subjects/${chapter.value.sub_subject_id}/chapters`)
  } else {
    router.back()
  }
}

function goPractice() {
  if (!chapter.value || chapter.value.question_count === 0) return
  router.push(`/practice/chapter/${chapter.value.id}`)
}

const aiModalVisible = ref(false)

function openAiModal() {
  if (!chapter.value) return
  if (!aiStore.hasConfig) {
    aiStore.syncConfig()
  }
  aiModalVisible.value = true
}

async function onAiGenerated() {
  await load()
}

onMounted(load)
</script>

<template>
  <div>
    <BasePageHeader
      :title="chapter?.name || '章节文档'"
      :subtitle="chapter ? `阅读完文档后点击右侧按钮开始刷题 · 本章共 ${chapter.question_count} 题` : '加载中…'"
    >
      <template #actions>
        <BaseButton type="secondary" @click="goBack">← 返回章节</BaseButton>
        <BaseButton
          v-if="chapter"
          type="secondary"
          class="!border-violet-200 !bg-white !text-violet-700 hover:!bg-violet-50/60 hover:!text-violet-800"
          @click="openAiModal"
        >
          <span class="flex items-center gap-1.5">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456z" />
            </svg>
            AI 出题
          </span>
        </BaseButton>
        <BaseButton
          type="primary"
          :disabled="!chapter || chapter.question_count === 0"
          @click="goPractice"
        >
          <span class="flex items-center gap-1.5">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 010 1.972l-11.54 6.347a1.125 1.125 0 01-1.667-.986V5.653z" />
            </svg>
            开始练习 ({{ chapter?.question_count || 0 }})
          </span>
        </BaseButton>
      </template>
    </BasePageHeader>

    <BaseSkeleton v-if="loading" variant="list" :count="3" />

    <BaseEmpty
      v-else-if="error"
      title="加载失败"
      :description="error"
    >
      <template #action>
        <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重试</BaseButton>
      </template>
    </BaseEmpty>

    <template v-else>
      <!-- 章节概览卡片 -->
      <BaseCard
        v-if="chapter"
        class="mb-4 border-indigo-100 bg-gradient-to-r from-indigo-50/60 via-white to-violet-50/50 p-5"
      >
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white text-sm font-semibold text-indigo-600 shadow-sm ring-1 ring-indigo-100">
            {{ chapter.sort_order || '·' }}
          </div>
          <div class="min-w-0 flex-1">
            <h2 class="text-base font-semibold text-gray-900">{{ chapter.name }}</h2>
            <p class="mt-0.5 text-xs text-gray-500">
              知识文档 · {{ tree.length }} 个主题 / {{ nodeCount }} 个节点 · 建议先通读再刷题
            </p>
          </div>
          <button
            class="group flex items-center gap-1.5 rounded-full bg-white px-3.5 py-1.5 text-sm font-semibold text-indigo-600 shadow-sm ring-1 ring-inset ring-indigo-200 transition-all duration-200 hover:shadow-md hover:ring-indigo-300 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:shadow-sm"
            :disabled="chapter.question_count === 0"
            @click="goPractice"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 010 1.972l-11.54 6.347a1.125 1.125 0 01-1.667-.986V5.653z" />
            </svg>
            练习本章
          </button>
          <button
            class="group flex items-center gap-1.5 rounded-full bg-white px-3.5 py-1.5 text-sm font-semibold text-violet-600 shadow-sm ring-1 ring-inset ring-violet-200 transition-all duration-200 hover:shadow-md hover:ring-violet-300"
            @click="openAiModal"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456z" />
            </svg>
            AI 出题
          </button>
        </div>
      </BaseCard>

      <!-- 无内容提示 -->
      <BaseEmpty
        v-if="tree.length === 0"
        title="该章节暂无文档"
        :description="hasQuestions ? '可直接开始练习' : '尚未上传知识点文档'"
      >
        <template #action>
          <BaseButton
            type="primary"
            :disabled="!hasQuestions"
            class="mt-2"
            @click="goPractice"
          >
            {{ hasQuestions ? '开始练习' : '返回章节' }}
          </BaseButton>
        </template>
      </BaseEmpty>

      <template v-else>
        <!-- 工具栏 -->
        <div class="mb-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <input
            v-model="keyword"
            type="text"
            placeholder="搜索文档内容…"
            class="w-full rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm placeholder-gray-400 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 sm:w-72"
          />
          <div class="flex shrink-0 items-center gap-2">
            <span v-if="keyword.trim()" class="text-xs text-gray-400">匹配 {{ nodeCount }} 个节点</span>
            <button
              class="cursor-pointer rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors duration-200 hover:border-indigo-300 hover:bg-indigo-50/40 hover:text-indigo-700"
              @click="setExpandMode('all')"
            >
              展开全部
            </button>
            <button
              class="cursor-pointer rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors duration-200 hover:border-indigo-300 hover:bg-indigo-50/40 hover:text-indigo-700"
              @click="setExpandMode('none')"
            >
              收起
            </button>
          </div>
        </div>

        <!-- MD 风格文档主体 -->
        <BaseCard class="p-6 sm:p-8">
          <article :key="expandKey" class="space-y-1.5">
            <ChapterDocNode
              v-for="(root, i) in tree"
              :key="i"
              :node="root"
              :depth="0"
              :initial-expand="expandMode"
              :keyword="keyword"
            />
          </article>
        </BaseCard>

        <!-- 底部再放一个明显的练习入口 -->
        <BaseCard
          v-if="chapter && chapter.question_count > 0"
          class="mt-4 border-emerald-100 bg-gradient-to-r from-emerald-50/70 via-white to-indigo-50/50 p-5"
        >
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p class="text-sm font-semibold text-gray-900">已读文档，开始刷题</p>
              <p class="mt-0.5 text-xs text-gray-500">本章共 {{ chapter.question_count }} 道题，做完即可查看解析</p>
            </div>
            <div class="flex flex-wrap gap-2">
              <BaseButton type="secondary" @click="openAiModal">
                <span class="flex items-center gap-1.5">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456z" />
                  </svg>
                  AI 出题
                </span>
              </BaseButton>
              <BaseButton type="primary" @click="goPractice">
                <span class="flex items-center gap-1.5">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  开始练习
                </span>
              </BaseButton>
            </div>
          </div>
        </BaseCard>
      </template>
    </template>

    <ChapterAiGenerateModal
      v-if="chapter"
      v-model:visible="aiModalVisible"
      :chapter="chapter"
      @generated="onAiGenerated"
    />
  </div>
</template>
