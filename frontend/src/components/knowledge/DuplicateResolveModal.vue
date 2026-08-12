<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { sanitizeHtml } from '@/utils/sanitize'
import { renderMarkdown } from '@/utils/markdown'
import { resolveDuplicate } from '@/api/knowledge'
import { showToast } from '@/utils/toast'
import type { ExistingKnowledgePoint } from '@/types/knowledge'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseButton from '@/components/common/BaseButton.vue'

type Action = 'keep_old' | 'replace' | 'add_new'

const props = defineProps<{
  modelValue: boolean
  existing: ExistingKnowledgePoint | null
  incoming: { name: string; description?: string; source_question_id?: number } | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'resolved', action: Action, result: { user_kp_id: number; status: string }): void
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: v => emit('update:modelValue', v),
})

const loading = ref(false)
const action = ref<Action>('keep_old')
const newName = ref('')

watch(
  () => props.modelValue,
  v => {
    if (v) {
      action.value = 'keep_old'
      newName.value = props.incoming?.name || props.existing?.name || ''
    }
  },
)

function renderHtml(desc: string): string {
  return sanitizeHtml(renderMarkdown(desc || ''))
}

const incomingName = computed(() => props.incoming?.name || props.existing?.name || '')

async function submit() {
  if (!props.existing) return
  if (action.value === 'add_new' && !newName.value.trim()) {
    showToast('请输入新名称', 'error')
    return
  }
  loading.value = true
  try {
    const res = await resolveDuplicate(props.existing.user_kp_id, {
      action: action.value,
      new_name: newName.value,
      new_description: props.incoming?.description || '',
    })
    showToast(
      action.value === 'keep_old' ? '已保留旧版' : action.value === 'replace' ? '已替换为新版' : '已添加为新条目',
      'success',
    )
    emit('resolved', action.value, { user_kp_id: res.user_kp_id, status: res.status })
    isOpen.value = false
  } catch (e) {
    showToast((e as Error).message || '操作失败', 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <BaseModal v-model="isOpen" title="该知识点已存在" width="max-w-3xl">
    <div v-if="existing" class="space-y-4">
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="rounded-xl border border-gray-100 bg-gray-50/60 p-4">
          <div class="mb-2 flex items-center gap-2">
            <span class="inline-flex items-center gap-1 rounded-full bg-gray-200/70 px-2 py-0.5 text-[10px] font-medium text-gray-600">
              旧版
            </span>
            <span v-if="existing.is_important" class="text-amber-500">
              <svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" /></svg>
            </span>
            <span v-if="existing.is_mastered" class="text-emerald-500">
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" /></svg>
            </span>
            <span class="text-xs text-gray-400">加入于 {{ existing.created_at }}</span>
          </div>
          <h4 class="mb-2 text-sm font-semibold text-gray-900">{{ existing.name }}</h4>
          <div
            v-if="existing.description"
            class="prose-sm max-h-48 overflow-y-auto text-xs text-gray-600"
            v-html="renderHtml(existing.description)"
          />
          <p v-else class="text-xs text-gray-400">（无描述）</p>
        </div>

        <div class="rounded-xl border border-indigo-100 bg-indigo-50/30 p-4">
          <div class="mb-2 flex items-center gap-2">
            <span class="inline-flex items-center gap-1 rounded-full bg-indigo-100 px-2 py-0.5 text-[10px] font-medium text-indigo-700">
              新版（来自 AI）
            </span>
          </div>
          <h4 class="mb-2 text-sm font-semibold text-gray-900">{{ incomingName }}</h4>
          <div
            v-if="incoming?.description"
            class="prose-sm max-h-48 overflow-y-auto text-xs text-gray-600"
            v-html="renderHtml(incoming.description)"
          />
          <p v-else class="text-xs text-gray-400">（无描述）</p>
        </div>
      </div>

      <div class="space-y-2 rounded-xl bg-gray-50/60 p-3">
        <label class="flex cursor-pointer items-start gap-3 rounded-lg p-2 transition-colors hover:bg-white">
          <input v-model="action" type="radio" value="keep_old" class="mt-1 cursor-pointer" />
          <div>
            <div class="text-sm font-medium text-gray-800">保留旧版</div>
            <div class="text-xs text-gray-500">关闭弹窗，已存在的知识点保持不变</div>
          </div>
        </label>
        <label class="flex cursor-pointer items-start gap-3 rounded-lg p-2 transition-colors hover:bg-white">
          <input v-model="action" type="radio" value="replace" class="mt-1 cursor-pointer" />
          <div>
            <div class="text-sm font-medium text-gray-800">用新版描述替换</div>
            <div class="text-xs text-gray-500">覆盖原有描述，名称保持不变</div>
          </div>
        </label>
        <label class="flex cursor-pointer items-start gap-3 rounded-lg p-2 transition-colors hover:bg-white">
          <input v-model="action" type="radio" value="add_new" class="mt-1 cursor-pointer" />
          <div class="flex-1">
            <div class="text-sm font-medium text-gray-800">仍添加为新条目</div>
            <div class="mb-1.5 text-xs text-gray-500">系统会自动追加「(副本 YYYY-MM-DD)」后缀</div>
            <input
              v-model="newName"
              type="text"
              placeholder="新条目名称"
              class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
            />
          </div>
        </label>
      </div>

      <div class="flex justify-end gap-2 border-t border-gray-100 pt-3">
        <BaseButton type="secondary" size="sm" :disabled="loading" @click="isOpen = false">取消</BaseButton>
        <BaseButton type="primary" size="sm" :loading="loading" @click="submit">确认</BaseButton>
      </div>
    </div>
  </BaseModal>
</template>
