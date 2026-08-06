<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useSubjectStore } from '@/stores/subject'
import { listMaterials, downloadMaterial } from '@/api/material'
import { showToast } from '@/utils/toast'
import { formatFileSize, timeAgo } from '@/utils/format'
import type { StudyMaterialResp } from '@/types/material'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseButton from '@/components/common/BaseButton.vue'

const subjectStore = useSubjectStore()

const materials = ref<StudyMaterialResp[]>([])
const loading = ref(true)
const error = ref('')
const activeSubjectId = ref<number | null>(null)
const preview = ref<StudyMaterialResp | null>(null)
const downloadingId = ref<number | null>(null)

const filterTabs = computed(() => {
  return [
    { label: '全部', value: 0 },
    ...subjectStore.subjects.map(s => ({ label: s.short_name || s.name, value: s.id })),
  ]
})

const filteredMaterials = computed(() => {
  if (!activeSubjectId.value) return materials.value
  return materials.value.filter(m => m.subject_id === activeSubjectId.value)
})

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    materials.value = await listMaterials()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function openPreview(m: StudyMaterialResp) {
  preview.value = m
}

function closePreview() {
  preview.value = null
}

async function onDownload(m: StudyMaterialResp) {
  if (downloadingId.value) return
  downloadingId.value = m.id
  try {
    await downloadMaterial(m.id)
    m.download_count += 1
    showToast('已开始下载', 'success')
  } catch {
    // 失败提示已由下载函数统一处理
  } finally {
    downloadingId.value = null
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') closePreview()
}

onMounted(async () => {
  window.addEventListener('keydown', onKeydown)
  if (subjectStore.levels.length === 0) await subjectStore.fetchLevels()
  if (subjectStore.subjects.length === 0) await subjectStore.ensureSubjects(subjectStore.levels[0]?.id || 1)
  await loadData()
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div>
    <BasePageHeader title="学习资料" subtitle="思维导图学习笔记，点击卡片在线预览，支持下载 xmind 源文件">
      <template #actions>
        <span class="hidden items-center gap-1.5 rounded-full bg-indigo-50 px-3 py-1.5 text-xs font-medium text-indigo-600 sm:flex">
          <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M7.5 7.5h9m-9 3H12m-4.5 3H15" />
          </svg>
          {{ materials.length }} 份资料
        </span>
      </template>
    </BasePageHeader>

    <!-- 科目筛选 -->
    <div v-if="materials.length > 0 && !loading" class="mb-5 flex flex-wrap gap-2">
      <button
        v-for="tab in filterTabs"
        :key="tab.value"
        class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
        :class="activeSubjectId === tab.value
          ? 'border-indigo-500 bg-indigo-50 text-indigo-700 shadow-sm'
          : 'border-gray-200 bg-white text-gray-600 hover:border-indigo-300 hover:bg-indigo-50/40'"
        @click="activeSubjectId = tab.value"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 加载骨架：贴合图片卡片布局 -->
    <div v-if="loading" class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="i in 6"
        :key="i"
        class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm"
      >
        <div class="skeleton aspect-[16/9] w-full" />
        <div class="p-4">
          <div class="skeleton mb-2 h-4 w-3/5" />
          <div class="skeleton mb-3 h-3 w-full" />
          <div class="skeleton mb-3 h-3 w-2/3" />
          <div class="skeleton h-8 w-full rounded-lg" />
        </div>
      </div>
    </div>

    <!-- 加载失败 -->
    <div v-else-if="error" class="rounded-xl border border-red-100 bg-red-50 p-6">
      <p class="text-sm text-red-600">{{ error }}</p>
    </div>

    <!-- 空状态 -->
    <div v-else-if="filteredMaterials.length === 0" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty
        :title="materials.length === 0 ? '暂无学习资料' : '该科目暂无资料'"
        description="资料整理中，敬请期待"
      />
    </div>

    <!-- 资料卡片网格 -->
    <div v-else class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
      <article
        v-for="m in filteredMaterials"
        :key="m.id"
        class="card-lift group flex flex-col overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm transition-all duration-200"
      >
        <!-- 封面预览图 -->
        <button
          class="relative block aspect-[16/9] w-full cursor-pointer overflow-hidden bg-gray-100"
          @click="openPreview(m)"
        >
          <img
            v-if="m.cover_url"
            :src="m.cover_url"
            :alt="m.title"
            loading="lazy"
            class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
          />
          <div v-else class="flex h-full w-full flex-col items-center justify-center gap-1.5 bg-gradient-to-br from-indigo-50 to-violet-50">
            <svg class="h-8 w-8 text-indigo-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
            </svg>
            <span class="text-xs text-indigo-400">暂无预览图</span>
          </div>

          <span v-if="m.subject_name" class="absolute left-2 top-2 rounded-full bg-white/90 px-2.5 py-0.5 text-xs font-medium text-indigo-600 shadow-sm backdrop-blur">
            {{ m.subject_name }}
          </span>

          <div class="absolute inset-0 flex items-center justify-center bg-gray-900/0 opacity-0 transition-all duration-200 group-hover:bg-gray-900/25 group-hover:opacity-100">
            <span class="flex items-center gap-1 rounded-full bg-white/95 px-3 py-1.5 text-xs font-medium text-gray-700 shadow-sm">
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              点击预览
            </span>
          </div>
        </button>

        <!-- 内容区 -->
        <div class="flex flex-1 flex-col p-4">
          <button class="mb-1 cursor-pointer text-left" @click="openPreview(m)">
            <h3 class="line-clamp-1 text-sm font-semibold tracking-tight text-gray-900 transition-colors group-hover:text-indigo-600">{{ m.title }}</h3>
          </button>
          <p class="mb-3 line-clamp-2 min-h-[2rem] text-xs leading-relaxed text-gray-500">
            {{ m.description || '暂无简介' }}
          </p>

          <div class="mt-auto flex items-center justify-between border-t border-gray-50 pt-3">
            <div class="flex items-center gap-3 text-xs text-gray-400">
              <span class="flex items-center gap-1">
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" />
                </svg>
                {{ m.download_count }}
              </span>
              <span>{{ formatFileSize(m.file_size) }}</span>
              <span class="hidden sm:block">{{ timeAgo(m.created_at) }}</span>
            </div>
            <BaseButton type="primary" size="sm" :loading="downloadingId === m.id" @click.stop="onDownload(m)">
              下载
            </BaseButton>
          </div>
        </div>
      </article>
    </div>

    <!-- 预览弹窗（lightbox） -->
    <Teleport to="body">
      <Transition name="lightbox">
        <div v-if="preview" class="fixed inset-0 z-[100] flex flex-col bg-gray-900/85 backdrop-blur-sm">
          <!-- 顶部栏 -->
          <div class="flex items-center justify-between px-4 py-3 sm:px-6">
            <span class="flex items-center gap-2 text-sm font-medium text-white/90">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              资料预览
            </span>
            <button
              class="cursor-pointer rounded-lg p-1.5 text-white/70 transition-colors hover:bg-white/10 hover:text-white"
              @click="closePreview"
            >
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- 预览大图 -->
          <div class="flex flex-1 items-center justify-center overflow-auto p-4 sm:p-6" @click.self="closePreview">
            <img
              v-if="preview.cover_url"
              :src="preview.cover_url"
              :alt="preview.title"
              class="max-h-full max-w-full rounded-xl bg-white/5 object-contain shadow-2xl"
            />
            <div v-else class="flex flex-col items-center gap-3 rounded-xl bg-white/5 p-16 text-center">
              <svg class="h-14 w-14 text-white/25" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
              </svg>
              <p class="text-sm text-white/40">该资料暂无预览图，请下载 xmind 查看</p>
            </div>
          </div>

          <!-- 底部信息条 -->
          <div class="border-t border-white/10 bg-gray-900/70 px-4 py-3.5 sm:px-6">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div class="min-w-0">
                <p class="truncate text-sm font-semibold text-white">{{ preview.title }}</p>
                <p class="mt-0.5 text-xs text-gray-400">
                  {{ preview.subject_name || '通用资料' }}
                  <span class="mx-1.5 text-gray-600">·</span>
                  {{ formatFileSize(preview.file_size) }}
                  <span class="mx-1.5 text-gray-600">·</span>
                  {{ preview.download_count }} 次下载
                </p>
                <p v-if="preview.description" class="mt-1 line-clamp-2 max-w-xl text-xs leading-relaxed text-gray-300">
                  {{ preview.description }}
                </p>
              </div>
              <div class="flex shrink-0 items-center gap-2">
                <span class="hidden text-xs text-gray-400 md:block">{{ timeAgo(preview.created_at) }}</span>
                <BaseButton type="primary" size="lg" :loading="downloadingId === preview.id" @click="onDownload(preview)">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" />
                  </svg>
                  下载 xmind
                </BaseButton>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.lightbox-enter-active,
.lightbox-leave-active {
  transition: opacity 0.2s ease;
}
.lightbox-enter-from,
.lightbox-leave-to {
  opacity: 0;
}
</style>
