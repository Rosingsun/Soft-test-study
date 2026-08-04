<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listFolders, createFolder, listFavorites, removeFavorite } from '@/api/bookmark'
import { getQuestion } from '@/api/question'
import { typeLabel } from '@/utils/question'
import type { FavoriteFolder, QuestionFavoriteResp, Question } from '@/types/question'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseInput from '@/components/common/BaseInput.vue'

const router = useRouter()

const folders = ref<FavoriteFolder[]>([])
const currentFolderId = ref<number | null>(null)
const favorites = ref<QuestionFavoriteResp[]>([])
const questions = ref<Record<number, Question>>({})
const loading = ref(true)
const error = ref('')
const showCreateFolder = ref(false)
const newFolderName = ref('')
const creatingFolder = ref(false)

const typeBadges: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'default'> = {
  single: 'info',
  multi: 'success',
  judge: 'warning',
  fill: 'default',
  short: 'default',
  comprehensive: 'info',
}

async function loadFolders() {
  folders.value = await listFolders()
}

async function loadFavorites() {
  loading.value = true
  error.value = ''
  try {
    favorites.value = currentFolderId.value
      ? await listFavorites(currentFolderId.value)
      : await listFavorites()
    const ids = [...new Set(favorites.value.map(f => f.question_id))]
    const qs = await Promise.all(ids.map(id => getQuestion(id).catch(() => null)))
    questions.value = {}
    ids.forEach((id, idx) => {
      if (qs[idx]) questions.value[id] = qs[idx] as Question
    })
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function handleCreateFolder() {
  const name = newFolderName.value.trim()
  if (!name || creatingFolder.value) return
  creatingFolder.value = true
  try {
    await createFolder({ name })
    newFolderName.value = ''
    showCreateFolder.value = false
    await loadFolders()
  } catch {
  } finally {
    creatingFolder.value = false
  }
}

async function handleRemove(questionId: number) {
  try {
    await removeFavorite(questionId)
    await loadFavorites()
  } catch {
  }
}

function stripHtml(html: string) {
  return html.replace(/<[^>]*>/g, '')
}

onMounted(async () => {
  try {
    await loadFolders()
    await loadFavorites()
  } catch (e) {
    error.value = (e as Error).message
    loading.value = false
  }
})
</script>

<template>
  <div>
    <BasePageHeader title="我的收藏" subtitle="收藏夹帮你整理高频考点，随时回看复习">
      <template #actions>
        <BaseButton @click="showCreateFolder = true">
          新建收藏夹
        </BaseButton>
      </template>
    </BasePageHeader>

    <div class="mb-6 flex flex-wrap items-center gap-2">
      <button
        type="button"
        class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
        :class="!currentFolderId
          ? 'border-indigo-500 bg-indigo-50 text-indigo-700 shadow-sm'
          : 'border-gray-200 bg-white text-gray-600 hover:border-indigo-300 hover:text-indigo-600'"
        @click="currentFolderId = null; loadFavorites()"
      >
        全部
      </button>
      <button
        v-for="folder in folders"
        :key="folder.id"
        type="button"
        class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all duration-200"
        :class="currentFolderId === folder.id
          ? 'border-indigo-500 bg-indigo-50 text-indigo-700 shadow-sm'
          : 'border-gray-200 bg-white text-gray-600 hover:border-indigo-300 hover:text-indigo-600'"
        @click="currentFolderId = folder.id; loadFavorites()"
      >
        {{ folder.name }}
      </button>
    </div>

    <div v-if="loading">
      <BaseLoading />
    </div>

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="loadFavorites">
            重新加载
          </BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else-if="favorites.length === 0" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="暂无收藏题目" description="去题库中收藏值得反复练习的题目吧">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="router.push('/practice/random')">
            去练习
          </BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else class="space-y-3">
      <BaseCard v-for="fav in favorites" :key="fav.id" class="flex flex-col gap-3">
        <div class="flex flex-wrap items-center gap-2">
          <BaseBadge :type="typeBadges[questions[fav.question_id]?.type]">
            {{ typeLabel(questions[fav.question_id]?.type || '') || '题目' }}
          </BaseBadge>
          <span class="text-xs text-gray-400">收藏于 {{ fav.created_at }}</span>
        </div>
        <p class="text-sm text-gray-700 line-clamp-2">
          {{ stripHtml(questions[fav.question_id]?.content || '') || '题目内容加载中' }}
        </p>
        <div class="flex items-center gap-2">
          <router-link
            v-if="questions[fav.question_id]?.chapter_id"
            :to="`/practice/chapter/${questions[fav.question_id]?.chapter_id}`"
            class="text-xs text-indigo-600 hover:underline"
          >
            去练习 →
          </router-link>
          <BaseButton type="ghost" size="sm" class="ml-auto" @click="handleRemove(fav.question_id)">
            取消收藏
          </BaseButton>
        </div>
      </BaseCard>
    </div>

    <BaseModal v-model="showCreateFolder" title="新建收藏夹">
      <div class="flex flex-col gap-4">
        <BaseInput v-model="newFolderName" placeholder="请输入收藏夹名称" @keyup.enter="handleCreateFolder" />
        <div class="flex justify-end gap-2">
          <BaseButton type="secondary" size="sm" @click="showCreateFolder = false">
            取消
          </BaseButton>
          <BaseButton size="sm" :loading="creatingFolder" @click="handleCreateFolder">
            创建
          </BaseButton>
        </div>
      </div>
    </BaseModal>
  </div>
</template>
