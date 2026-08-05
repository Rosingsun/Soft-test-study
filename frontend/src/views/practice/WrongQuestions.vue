<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { listWrongQuestions, removeWrongQuestion } from '@/api/wrong_question'
import { typeLabel } from '@/utils/question'
import { timeAgo } from '@/utils/format'
import type { WrongQuestionResp } from '@/types/question'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'

const auth = useAuthStore()
const router = useRouter()
const list = ref<WrongQuestionResp[]>([])
const loading = ref(true)
const error = ref('')

const subtitle = computed(() => (list.value.length ? `共 ${list.value.length} 道错题` : ''))

async function loadWrongQuestions() {
  loading.value = true
  error.value = ''
  try {
    const subjectId = auth.selectedSubjectId || undefined
    list.value = await listWrongQuestions(subjectId)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(loadWrongQuestions)
watch(() => auth.selectedSubjectId, loadWrongQuestions)

async function handleRemove(questionId: number) {
  try {
    await removeWrongQuestion(questionId)
    list.value = list.value.filter(w => w.question_id !== questionId)
  } catch {}
}

function stripHtml(html: string) {
  return html.replace(/<[^>]*>/g, '')
}

const typeBadges: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'default'> = {
  single: 'info',
  multi: 'success',
  judge: 'warning',
  fill: 'default',
  short: 'default',
  comprehensive: 'info',
}
</script>

<template>
  <div>
    <BasePageHeader title="错题本" :subtitle="subtitle">
      <template #actions>
        <BaseButton v-if="list.length" @click="router.push('/practice/wrong')">
          错题练习
        </BaseButton>
      </template>
    </BasePageHeader>

    <BaseSkeleton v-if="loading" variant="list" :count="4" />

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="loadWrongQuestions">
            重新加载
          </BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else-if="list.length === 0" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="暂无错题" description="继续保持，错误会越来越少">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="router.push('/practice/random')">
            去练习
          </BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <BaseCard
        v-for="w in list"
        :key="w.id"
        class="flex flex-col gap-3"
      >
        <div class="flex flex-wrap items-center gap-2">
          <BaseBadge :type="typeBadges[w.type]">
            {{ typeLabel(w.type) }}
          </BaseBadge>
          <span class="text-xs text-gray-400">
            错 {{ w.wrong_count }} · 对 {{ w.correct_count }}
          </span>
        </div>
        <p class="line-clamp-3 min-h-0 flex-1 text-sm leading-6 text-gray-700">{{ stripHtml(w.content) }}</p>
        <div class="flex items-center gap-2 border-t border-gray-50 pt-3">
          <span class="min-w-0 flex-1 truncate text-xs text-gray-400">{{ timeAgo(w.last_wrong_at) }}</span>
          <BaseButton type="secondary" size="sm" @click="router.push(`/practice/wrong?qid=${w.question_id}`)">
            重做
          </BaseButton>
          <BaseButton type="ghost" size="sm" @click="handleRemove(w.question_id)">
            移出
          </BaseButton>
        </div>
      </BaseCard>
    </div>
  </div>
</template>
