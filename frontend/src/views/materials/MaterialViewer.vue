<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getMaterial, getMaterialContent, downloadMaterial, fetchMaterialBlob } from '@/api/material'
import { showToast } from '@/utils/toast'
import { formatFileSize } from '@/utils/format'
import {
  MATERIAL_TYPE_META,
  type MaterialFileType,
  type StudyMaterialResp,
  type MindMapNode,
} from '@/types/material'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseTabs from '@/components/common/BaseTabs.vue'
import MindMapTree from '@/components/material/MindMapTree.vue'
import MindMapGraph from '@/components/material/MindMapGraph.vue'

const route = useRoute()
const router = useRouter()

const material = ref<StudyMaterialResp | null>(null)
const tree = ref<MindMapNode[]>([])
const pdfUrl = ref('')
const loading = ref(true)
const error = ref('')
const downloading = ref(false)
const keyword = ref('')
const expandMode = ref<'all' | 'none' | null>(null)
const expandKey = ref(0)
const viewMode = ref<'tree' | 'graph'>('graph')

const fileType = computed<MaterialFileType>(() =>
  (material.value?.file_type as MaterialFileType) || 'xmind',
)
const typeMeta = computed(() => MATERIAL_TYPE_META[fileType.value])
const isPdf = computed(() => fileType.value === 'pdf')

async function load() {
  loading.value = true
  error.value = ''
  const id = Number(route.params.id)
  try {
    const m = await getMaterial(id)
    material.value = m
    if (fileType.value === 'xmind') {
      tree.value = await getMaterialContent(id)
    } else if (fileType.value === 'pdf') {
      // PDF：带 JWT 拉取 blob → objectURL → iframe 原生渲染
      const blob = await fetchMaterialBlob(id)
      pdfUrl.value = URL.createObjectURL(blob)
    }
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function cleanupPdfUrl() {
  if (pdfUrl.value) {
    URL.revokeObjectURL(pdfUrl.value)
    pdfUrl.value = ''
  }
}

onUnmounted(cleanupPdfUrl)

function countNodes(nodes: MindMapNode[]): number {
  let total = 0
  for (const n of nodes) {
    total += 1
    if (n.children?.length) total += countNodes(n.children)
  }
  return total
}

const nodeCount = computed(() => countNodes(tree.value))

const filteredTree = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return tree.value
  return pruneTree(tree.value, kw)
})

function pruneTree(nodes: MindMapNode[], kw: string): MindMapNode[] {
  const result: MindMapNode[] = []
  for (const n of nodes) {
    const children = n.children ? pruneTree(n.children, kw) : []
    const selfHit = n.title.toLowerCase().includes(kw)
    if (selfHit || children.length > 0) {
      result.push({ title: n.title, children })
    }
  }
  return result
}

function setExpandMode(mode: 'all' | 'none') {
  if (expandMode.value === mode) {
    expandMode.value = null
  } else {
    expandMode.value = mode
  }
  expandKey.value++
}

async function onDownload() {
  if (!material.value || downloading.value) return
  downloading.value = true
  try {
    await downloadMaterial(material.value.id)
    material.value.download_count += 1
    showToast('已开始下载', 'success')
  } catch {
    // 失败提示已由下载函数统一处理
  } finally {
    downloading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <BasePageHeader
      :title="material?.title || '资料详情'"
      :subtitle="material
        ? `${typeMeta?.label || '文件'} · ${material.subject_name || '通用资料'}${fileType === 'xmind' ? ` · ${nodeCount} 个节点` : ''} · ${formatFileSize(material.file_size)}`
        : ''"
    >
      <template #actions>
        <BaseButton type="secondary" @click="router.back()">← 返回</BaseButton>
        <BaseButton v-if="material" type="primary" :loading="downloading" @click="onDownload">
          下载 {{ typeMeta?.label || '文件' }}
        </BaseButton>
      </template>
    </BasePageHeader>

    <!-- 加载骨架 -->
    <div v-if="loading">
      <div class="mb-4 flex items-center gap-3">
        <div class="skeleton h-9 w-64 rounded-lg" />
        <div class="skeleton h-9 w-24 rounded-lg" />
        <div class="skeleton h-9 w-24 rounded-lg" />
      </div>
      <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
        <div class="skeleton mb-3 h-5 w-1/3 rounded" />
        <div class="space-y-2.5">
          <div v-for="i in 10" :key="i" class="skeleton h-4 rounded" :style="{ width: `${95 - i * 6}%` }" />
        </div>
      </div>
    </div>

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重试</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <template v-else>
      <!-- xmind：视图切换 + 搜索 + 展开/收起工具栏 -->
      <div v-if="fileType === 'xmind'" class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex items-center gap-3">
          <BaseTabs
            v-model="viewMode"
            :tabs="[
              { label: '思维导图', value: 'graph' },
              { label: '树形', value: 'tree' },
            ]"
            size="sm"
          />
          <BaseInput v-model="keyword" placeholder="搜索节点标题…" class="w-full sm:w-64" />
        </div>
        <div v-if="viewMode === 'tree'" class="flex shrink-0 items-center gap-2">
          <span v-if="keyword.trim()" class="text-xs text-gray-400">匹配 {{ countNodes(filteredTree) }} 个节点</span>
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

      <!-- xmind 思维导图（分支图） -->
      <BaseCard v-if="fileType === 'xmind' && filteredTree.length && viewMode === 'graph'" class="p-0">
        <MindMapGraph :nodes="filteredTree" :keyword="keyword" />
      </BaseCard>

      <!-- xmind 树形 -->
      <BaseCard v-else-if="fileType === 'xmind' && filteredTree.length && viewMode === 'tree'" class="overflow-x-auto p-6">
        <ul :key="expandKey" class="min-w-max space-y-1">
          <MindMapTree
            v-for="(root, i) in filteredTree"
            :key="i"
            :node="root"
            :depth="0"
            :initial-expand="expandMode"
            :keyword="keyword"
          />
        </ul>
      </BaseCard>

      <div v-else-if="fileType === 'xmind'" class="rounded-xl border border-gray-100 bg-white shadow-sm">
        <BaseEmpty title="未找到匹配内容" description="换个关键词试试" />
      </div>

      <!-- PDF：浏览器原生预览 -->
      <div v-else-if="isPdf && pdfUrl" class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
        <iframe
          :src="pdfUrl"
          class="h-[calc(100vh-16rem)] min-h-[32rem] w-full"
          title="PDF 在线预览"
        />
      </div>

      <!-- word / ppt：仅支持下载 -->
      <div v-else class="rounded-xl border border-gray-100 bg-white shadow-sm">
        <BaseEmpty
          title="该类型暂不支持在线预览"
          :description="`${typeMeta?.label || '文件'} 文件请下载后使用本地办公软件打开`"
        >
          <template #action>
            <BaseButton type="primary" size="sm" class="mt-4" :loading="downloading" @click="onDownload">
              下载 {{ typeMeta?.label || '文件' }}
            </BaseButton>
          </template>
        </BaseEmpty>
      </div>
    </template>
  </div>
</template>
