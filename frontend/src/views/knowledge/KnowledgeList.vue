<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { sanitizeHtml } from '@/utils/sanitize'
import { renderMarkdown } from '@/utils/markdown'
import { showToast } from '@/utils/toast'
import {
  listKnowledgePoints,
  getKnowledgeStats,
  updateKnowledgePoint,
  removeKnowledgePoint,
} from '@/api/knowledge'
import { getSubjectsByLevel } from '@/api/subject'
import { useAuthStore } from '@/stores/auth'
import type { UserKnowledgePointResp, KnowledgePointStatsResp, KnowledgeFilter } from '@/types/knowledge'
import type { Subject } from '@/types/subject'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseTabs from '@/components/common/BaseTabs.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

const router = useRouter()
const auth = useAuthStore()

const loading = ref(true)
const error = ref('')
const list = ref<UserKnowledgePointResp[]>([])
const total = ref(0)
const stats = ref<KnowledgePointStatsResp | null>(null)
const filter = ref<KnowledgeFilter>('all')
const keyword = ref('')
const subjectId = ref<number>(0)
const subjects = ref<Subject[]>([])
const confirmDialog = ref<InstanceType<typeof ConfirmDialog> | null>(null)
const removeTargetId = ref<number | null>(null)
const toggling = ref<Set<number>>(new Set())

const filterTabs = computed(() => [
  { label: '全部', value: 'all' as const, badge: stats.value?.total ?? 0 },
  { label: '★ 重点', value: 'important' as const, badge: stats.value?.important ?? 0 },
  { label: '✓ 已掌握', value: 'mastered' as const, badge: stats.value?.mastered ?? 0 },
  { label: '待复习', value: 'unmastered' as const, badge: stats.value?.unmastered ?? 0 },
])

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [page, s] = await Promise.all([
      listKnowledgePoints({
        filter: filter.value === 'all' ? undefined : filter.value,
        subject_id: subjectId.value || undefined,
        keyword: keyword.value.trim() || undefined,
        page: 1,
        page_size: 50,
      }),
      getKnowledgeStats(),
    ])
    list.value = page.list
    total.value = page.total
    stats.value = s
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function loadSubjects() {
  if (auth.selectedLevelId) {
    try {
      subjects.value = await getSubjectsByLevel(auth.selectedLevelId)
    } catch {
      subjects.value = []
    }
  }
}

onMounted(async () => {
  await loadSubjects()
  await load()
})

watch([filter, subjectId], () => load())

let keywordDebounce: ReturnType<typeof setTimeout> | null = null
watch(keyword, () => {
  if (keywordDebounce) clearTimeout(keywordDebounce)
  keywordDebounce = setTimeout(load, 350)
})

function renderHtml(desc: string) {
  return sanitizeHtml(renderMarkdown(desc || ''))
}

async function toggleImportant(item: UserKnowledgePointResp) {
  if (toggling.value.has(item.id)) return
  toggling.value.add(item.id)
  toggling.value = new Set(toggling.value)
  const next = !item.is_important
  try {
    await updateKnowledgePoint(item.id, { is_important: next })
    item.is_important = next
  } catch (e) {
    showToast((e as Error).message || '操作失败', 'error')
  } finally {
    toggling.value.delete(item.id)
    toggling.value = new Set(toggling.value)
  }
}

async function toggleMastered(item: UserKnowledgePointResp) {
  if (toggling.value.has(item.id)) return
  toggling.value.add(item.id)
  toggling.value = new Set(toggling.value)
  const next = !item.is_mastered
  try {
    await updateKnowledgePoint(item.id, { is_mastered: next })
    item.is_mastered = next
    if (next) item.mastered_at = new Date().toISOString()
    showToast(next ? '已标记为掌握' : '已取消掌握', 'success')
    if (stats.value) {
      stats.value.mastered += next ? 1 : -1
      stats.value.unmastered = stats.value.total - stats.value.mastered
    }
  } catch (e) {
    showToast((e as Error).message || '操作失败', 'error')
  } finally {
    toggling.value.delete(item.id)
    toggling.value = new Set(toggling.value)
  }
}

function askRemove(id: number) {
  removeTargetId.value = id
  confirmDialog.value?.open()
}

async function confirmRemove() {
  const id = removeTargetId.value
  if (!id) return
  try {
    await removeKnowledgePoint(id)
    showToast('已移除', 'success')
    list.value = list.value.filter(v => v.id !== id)
    total.value = Math.max(0, total.value - 1)
    if (stats.value) {
      stats.value.total = Math.max(0, stats.value.total - 1)
      stats.value.unmastered = stats.value.total - stats.value.mastered
    }
  } catch (e) {
    showToast((e as Error).message || '删除失败', 'error')
  } finally {
    removeTargetId.value = null
  }
}

function goQuestion(questionId: number) {
  if (!questionId) return
  router.push(`/practice/random`)
}
</script>

<template>
  <div>
    <BasePageHeader title="我的知识点" subtitle="AI 智能提取的考点碎片，集中沉淀便于长期复习" />

    <div v-if="stats" class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div class="rounded-xl border border-indigo-100 bg-gradient-to-br from-indigo-50/80 to-white p-4 shadow-sm">
        <p class="text-2xl font-bold tracking-tight text-indigo-600">{{ stats.total }}</p>
        <p class="mt-1 text-xs text-gray-500">知识点总数</p>
      </div>
      <div class="rounded-xl border border-amber-100 bg-amber-50/50 p-4 shadow-sm">
        <p class="text-2xl font-bold tracking-tight text-amber-500">{{ stats.important }}</p>
        <p class="mt-1 text-xs text-gray-500">重点收藏</p>
      </div>
      <div class="rounded-xl border border-emerald-100 bg-emerald-50/50 p-4 shadow-sm">
        <p class="text-2xl font-bold tracking-tight text-emerald-600">{{ stats.mastered }}</p>
        <p class="mt-1 text-xs text-gray-500">已掌握</p>
      </div>
      <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
        <p class="text-2xl font-bold tracking-tight text-gray-900">{{ stats.recent_week }}</p>
        <p class="mt-1 text-xs text-gray-400">本周新增</p>
      </div>
    </div>

    <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <BaseTabs v-model="filter" :tabs="filterTabs" variant="gradient" size="sm" />
      <div class="flex flex-wrap items-center gap-2">
        <BaseSelect v-model="subjectId" class="w-40">
          <option :value="0">全部科目</option>
          <option v-for="s in subjects" :key="s.id" :value="s.id">{{ s.short_name || s.name }}</option>
        </BaseSelect>
        <BaseInput v-model="keyword" placeholder="搜索知识点…" class="w-56" />
      </div>
    </div>

    <BaseSkeleton v-if="loading" variant="list" :count="4" />

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重新加载</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else-if="!list.length" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="还没有知识点" description="在练习时点击解析面板右上角的「获取知识点」按钮，AI 会帮你提炼核心考点">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="router.push('/practice/random')">
            去练习
          </BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <BaseCard
        v-for="item in list"
        :key="item.id"
        hover
        :class="[
          'flex flex-col gap-3 transition-opacity',
          item.is_mastered ? 'opacity-70' : '',
        ]"
      >
        <div class="flex items-start gap-2">
          <button
            v-if="item.is_important"
            class="text-amber-500 transition-transform hover:scale-110"
            title="取消重点"
            @click="toggleImportant(item)"
          >
            <svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" /></svg>
          </button>
          <button
            v-else
            class="text-gray-300 transition-colors hover:text-amber-500"
            title="标记为重点"
            @click="toggleImportant(item)"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.563.563 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z" />
            </svg>
          </button>
          <h3 class="flex-1 text-sm font-semibold tracking-tight text-gray-900">
            {{ item.knowledge.name }}
          </h3>
          <span
            v-if="item.is_mastered"
            class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-[10px] font-medium text-emerald-600 ring-1 ring-inset ring-emerald-600/20"
          >
            <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" /></svg>
            已掌握
          </span>
        </div>

        <div
          v-if="item.knowledge.description"
          class="prose-sm text-xs text-gray-600 max-h-24 overflow-hidden"
          v-html="renderHtml(item.knowledge.description)"
        />

        <div class="flex items-center justify-between text-[11px] text-gray-400">
          <span>加入于 {{ item.created_at }}</span>
          <a
            v-if="item.source_question_id"
            class="cursor-pointer text-indigo-500 transition-colors hover:text-indigo-700"
            @click.prevent="goQuestion(item.source_question_id)"
          >来源题 →</a>
        </div>

        <div class="flex items-center gap-1.5 border-t border-gray-50 pt-3">
          <BaseButton
            type="ghost"
            size="sm"
            :disabled="toggling.has(item.id)"
            @click="toggleMastered(item)"
          >
            <svg
              class="h-3.5 w-3.5"
              :class="item.is_mastered ? 'text-emerald-500' : 'text-gray-400'"
              :fill="item.is_mastered ? 'currentColor' : 'none'"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ item.is_mastered ? '已掌握' : '标记掌握' }}
          </BaseButton>
          <BaseButton type="ghost" size="sm" class="ml-auto" @click="askRemove(item.id)">
            <svg class="h-3.5 w-3.5 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
            </svg>
            删除
          </BaseButton>
        </div>
      </BaseCard>
    </div>

    <ConfirmDialog
      ref="confirmDialog"
      title="确认删除"
      description="删除后该知识点将从你的列表中移除，全局知识点仍保留供其他用户使用。"
      confirm-text="确认删除"
      danger
      @confirm="confirmRemove"
    />
  </div>
</template>
