<script setup lang="ts">
import { ref, computed } from 'vue'
import { sanitizeHtml } from '@/utils/sanitize'
import { renderMarkdown } from '@/utils/markdown'
import { extractKnowledgePoints, addKnowledgePoint } from '@/api/knowledge'
import { analyzeQuestion } from '@/api/ai'
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
// AI 兜底分析:当知识库提取失败时使用,提供对题目本身的整体解析
const analyzing = ref(false)
const analysis = ref('')

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
  // 重新提取时清空上一次的兜底分析,避免误用旧结果
  analysis.value = ''
  try {
    const res = await extractKnowledgePoints({
      api_config: aiStore.getConfig(),
      question_content: props.questionContent,
      question_type: props.questionType || '',
      subject_id: props.subjectId,
    })
    points.value = res.points
  } catch (e) {
    showToast((e as Error).message || 'AI 提取失败', 'error')
  } finally {
    loading.value = false
  }
}

// 兜底入口:当 AI 未提取到有效知识点时,直接对题目做整体解析(知识点/正确答案分析/巩固建议)。
// 后端 /ai/analyze 要求 user_answer,这里用占位文字传过去,prompt 中"用户答案哪里错了"会被自然跳过。
async function handleAnalyze() {
  if (!aiStore.hasConfig) {
    if (confirm('尚未配置 AI API Key，是否前往配置？')) {
      router.push('/ai/config')
    }
    return
  }
  if (!props.questionContent) {
    showToast('题目内容为空', 'error')
    return
  }
  analyzing.value = true
  analysis.value = ''
  try {
    const res = await analyzeQuestion({
      api_config: aiStore.getConfig(),
      question_content: props.questionContent,
      question_type: props.questionType || '',
      question_answer: props.questionAnswer || '',
      user_answer: '（用户尚未作答，请直接对题目本身做整体解析）',
    })
    analysis.value = res.analysis
  } catch (e) {
    showToast((e as Error).message || 'AI 分析失败', 'error')
  } finally {
    analyzing.value = false
  }
}

// 在分析结果页点"重新提取",清空分析并走一次 extract
function retryExtract() {
  analysis.value = ''
  loadPoints()
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

    <!-- AI 兜底分析进行中 -->
    <div v-else-if="analyzing" class="space-y-3">
      <div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4">
        <div class="skeleton mb-2 h-4 w-40" />
        <div class="skeleton h-3 w-full" />
        <div class="skeleton mt-1.5 h-3 w-11/12" />
        <div class="skeleton mt-1.5 h-3 w-3/4" />
      </div>
      <p class="text-center text-xs text-indigo-600">
        <span class="inline-block h-3 w-3 animate-spin rounded-full border-2 border-indigo-600 border-t-transparent align-middle" />
        AI 正在解析题目…
      </p>
    </div>

    <!-- AI 兜底分析结果:知识库提取失败时的备选视图 -->
    <div v-else-if="analysis" class="space-y-3">
      <div class="flex items-center gap-2 rounded-lg border border-indigo-100 bg-indigo-50/40 px-3 py-2 text-xs text-indigo-700">
        <svg class="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
        </svg>
        <span>AI 对题目的整体解析（未能拆出可加入知识库的条目时使用）</span>
      </div>
      <div
        class="prose-sm max-h-[60vh] overflow-y-auto whitespace-pre-line rounded-xl border border-gray-100 bg-white p-4 text-sm leading-relaxed text-gray-700"
        v-html="renderHtml(analysis)"
      />
    </div>

    <!-- 空状态:知识库提取失败 + 提供 AI 兜底分析入口 -->
    <BaseEmpty
      v-else-if="!points.length"
      title="AI 未提取到有效知识点"
      description="可点击下方按钮，让 AI 直接对题目做整体解析"
    >
      <template #action>
        <BaseButton type="primary" size="sm" @click="handleAnalyze">
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
          </svg>
          AI分析
        </BaseButton>
      </template>
    </BaseEmpty>

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

    <!-- 知识库提取结果底部操作 -->
    <template v-if="!loading && !analyzing && !analysis && points.length">
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

    <!-- AI 兜底分析结果底部操作:可重新走知识库提取 -->
    <template v-else-if="!loading && !analyzing && analysis">
      <div class="mt-5 flex items-center justify-end gap-2 border-t border-gray-100 pt-4">
        <BaseButton type="secondary" size="sm" @click="close">关闭</BaseButton>
        <BaseButton type="primary" size="sm" @click="retryExtract">重新提取</BaseButton>
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
