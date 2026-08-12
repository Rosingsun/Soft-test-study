<script setup lang="ts">
import { ref, computed } from 'vue'
import { sanitizeHtml } from '@/utils/sanitize'
import { renderMarkdown } from '@/utils/markdown'
import { extractKnowledgePoints, addKnowledgePoint } from '@/api/knowledge'
import { useAiStore } from '@/stores/ai'
import { showToast } from '@/utils/toast'
import { useRouter } from 'vue-router'
import type { KnowledgePointItem } from '@/types/knowledge'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'

const props = withDefaults(defineProps<{
  modelValue: boolean
  questionId?: number
  questionContent?: string
  questionType?: string
  questionAnswer?: string
  questionAnalysis?: string
  subjectId?: number
}>(), {})

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'added', items: { name: string; status: string }[]): void
}>()

const router = useRouter()
const aiStore = useAiStore()

const points = ref<KnowledgePointItem[]>([])
const loading = ref(false)
const adding = ref<Set<number>>(new Set())
const added = ref<Set<number>>(new Set())
const expanded = ref<Set<number>>(new Set())
const duplicateInfo = ref<{ index: number; existing_name: string } | null>(null)

const isOpen = computed({
  get: () => props.modelValue,
  set: v => emit('update:modelValue', v),
})

function close() {
  isOpen.value = false
}

async function loadPoints() {
  if (!aiStore.hasConfig) {
    if (confirm('尚未配置 AI API Key，是否前往配置？')) {
      router.push('/ai/config')
    }
    close()
    return
  }
  if (!props.questionContent) {
    showToast('题目内容为空', 'error')
    return
  }
  loading.value = true
  points.value = []
  added.value = new Set()
  expanded.value = new Set()
  try {
    const res = await extractKnowledgePoints({
      api_config: aiStore.getConfig(),
      question_content: props.questionContent,
      question_type: props.questionType || '',
      question_answer: props.questionAnswer || '',
      question_analysis: props.questionAnalysis || '',
      subject_id: props.subjectId,
    })
    points.value = res.points
  } catch (e) {
    showToast((e as Error).message || 'AI 提取失败', 'error')
  } finally {
    loading.value = false
  }
}

function toggleExpand(i: number) {
  if (expanded.value.has(i)) expanded.value.delete(i)
  else expanded.value.add(i)
  expanded.value = new Set(expanded.value)
}

function isExpanded(i: number) {
  return expanded.value.has(i)
}

function needsFold(desc: string): boolean {
  return desc.split('\n').length > 5
}

function renderHtml(desc: string): string {
  return sanitizeHtml(renderMarkdown(desc))
}

async function addOne(i: number) {
  const p = points.value[i]
  if (!p || adding.value.has(i) || added.value.has(i)) return
  adding.value.add(i)
  adding.value = new Set(adding.value)
  try {
    const res = await addKnowledgePoint({
      name: p.name,
      description: p.description,
      subject_id: props.subjectId,
      source_question_id: props.questionId,
    })
    added.value.add(i)
    added.value = new Set(added.value)
    showToast('已加入：' + p.name, 'success')
    emit('added', [{ name: p.name, status: res.status }])
  } catch (e) {
    const err = e as Error & { code?: number; data?: { existing?: { name: string } } }
    if (err.code === 10009 && err.data?.existing) {
      duplicateInfo.value = { index: i, existing_name: err.data.existing.name }
    } else {
      showToast(err.message || '加入失败', 'error')
    }
  } finally {
    adding.value.delete(i)
    adding.value = new Set(adding.value)
  }
}

async function addAll() {
  const pending = points.value
    .map((_, i) => i)
    .filter(i => !added.value.has(i))
  for (const i of pending) {
    await addOne(i)
  }
}

function onModalOpen(v: boolean) {
  if (v) loadPoints()
}
</script>

<template>
  <BaseModal v-model="isOpen" title="AI 提取的知识点" width="max-w-2xl" @update:model-value="onModalOpen">
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 3" :key="i" class="rounded-xl border border-gray-100 bg-gray-50/50 p-4">
        <div class="skeleton mb-2 h-4 w-32" />
        <div class="skeleton h-3 w-full" />
        <div class="skeleton mt-1.5 h-3 w-3/4" />
      </div>
    </div>

    <BaseEmpty v-else-if="!points.length" title="AI 未提取到有效知识点" description="可尝试重新回答本题" />
    <div v-else class="space-y-3">
      <div
        v-for="(p, i) in points"
        :key="i"
        class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-colors hover:border-indigo-200"
      >
        <div class="mb-2 flex items-start justify-between gap-3">
          <div class="flex min-w-0 items-start gap-2">
            <span class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-lg bg-violet-50 text-xs font-semibold text-violet-600 ring-1 ring-inset ring-violet-600/20">
              {{ i + 1 }}
            </span>
            <h4 class="text-sm font-semibold tracking-tight text-gray-900">{{ p.name }}</h4>
          </div>
          <BaseButton
            v-if="!added.has(i)"
            type="primary"
            size="sm"
            :loading="adding.has(i)"
            @click="addOne(i)"
          >
            {{ adding.has(i) ? '加入中' : '加入' }}
          </BaseButton>
          <span
            v-else
            class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-600 ring-1 ring-inset ring-emerald-600/20"
          >
            <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
            已加入
          </span>
        </div>
        <div
          class="prose-sm text-xs text-gray-600"
          :class="isExpanded(i) ? '' : (needsFold(p.description) ? 'max-h-24 overflow-hidden' : '')"
          v-html="renderHtml(p.description)"
        />
        <button
          v-if="needsFold(p.description)"
          class="mt-1.5 cursor-pointer text-xs font-medium text-indigo-600 transition-colors hover:text-indigo-800"
          @click="toggleExpand(i)"
        >
          {{ isExpanded(i) ? '收起' : '展开全文' }}
        </button>
      </div>
    </div>

    <template v-if="!loading && points.length">
      <div class="mt-5 flex items-center justify-between border-t border-gray-100 pt-4">
        <span class="text-xs text-gray-400">已加入 {{ added.size }} / {{ points.length }} 条</span>
        <div class="flex items-center gap-2">
          <BaseButton type="secondary" size="sm" @click="close">关闭</BaseButton>
          <BaseButton
            type="primary"
            size="sm"
            :disabled="added.size === points.length"
            @click="addAll"
          >
            全部加入
          </BaseButton>
        </div>
      </div>
    </template>

    <div
      v-if="duplicateInfo"
      class="fixed inset-0 z-[110] flex items-center justify-center p-4"
      @click.self="duplicateInfo = null"
    >
      <div class="absolute inset-0 bg-gray-900/60 backdrop-blur-sm" />
      <div class="relative w-full max-w-md overflow-hidden rounded-2xl bg-white p-6 shadow-2xl">
        <h4 class="mb-2 text-base font-semibold tracking-tight text-gray-900">该知识点已存在</h4>
        <p class="mb-4 text-sm text-gray-500">
          「<span class="font-medium text-gray-800">{{ duplicateInfo.existing_name }}</span>」已存在于你的知识点中。可前往「学习 → 知识点」查看。
        </p>
        <div class="flex justify-end gap-2">
          <BaseButton type="secondary" size="sm" @click="duplicateInfo = null">我知道了</BaseButton>
          <BaseButton
            type="primary"
            size="sm"
            @click="router.push('/knowledge'); close(); duplicateInfo = null"
          >
            前往查看
          </BaseButton>
        </div>
      </div>
    </div>
  </BaseModal>
</template>
