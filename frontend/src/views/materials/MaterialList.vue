<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSubjectStore } from '@/stores/subject'
import { listMaterials, downloadMaterial } from '@/api/material'
import { showToast } from '@/utils/toast'
import { formatFileSize, timeAgo } from '@/utils/format'
import type { StudyMaterialResp } from '@/types/material'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseButton from '@/components/common/BaseButton.vue'

const router = useRouter()
const subjectStore = useSubjectStore()

const materials = ref<StudyMaterialResp[]>([])
const loading = ref(true)
const error = ref('')
const activeSubjectId = ref<number | null>(null)
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

function openViewer(m: StudyMaterialResp) {
  router.push(`/materials/${m.id}/view`)
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

onMounted(async () => {
  if (subjectStore.levels.length === 0) await subjectStore.fetchLevels()
  if (subjectStore.subjects.length === 0) await subjectStore.ensureSubjects(subjectStore.levels[0]?.id || 1)
  await loadData()
})
</script>

<template>
  <div>
    <BasePageHeader title="学习资料" subtitle="思维导图学习笔记，点击卡片在线阅读，支持下载 xmind 源文件">
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
          @click="openViewer(m)"
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
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.042A8.967 8.967 0 009 3.75a8.967 8.967 0 00-3 2.292m5.842.018a9.24 9.24 0 015.816-2.208c1.74 0 3.39.36 4.886 1.006A9.196 9.196 0 0112 21a9.196 9.196 0 01-6.544-2.658c-1.408-1.407-2.28-3.35-2.456-5.47m14.5 0a9.077 9.077 0 01-2.62 4.406M21 12a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.66 0 3-4.03 3-9s-1.34-9-3-9m0 18c-1.66 0-3-4.03-3-9s1.34-9 3-9" />
              </svg>
              在线阅读
            </span>
          </div>
        </button>

        <!-- 内容区 -->
        <div class="flex flex-1 flex-col p-4">
          <button class="mb-1 cursor-pointer text-left" @click="openViewer(m)">
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
  </div>
</template>

