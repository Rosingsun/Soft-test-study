<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSubjectStore } from '@/stores/subject'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'

const store = useSubjectStore()
const route = useRoute()
const router = useRouter()
const loading = ref(true)
const error = ref('')
const subSubjectName = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const id = Number(route.params.id)
    await store.fetchChapters(id)
    const sub = store.subSubjects.find(s => s.id === id)
    if (sub) {
      subSubjectName.value = sub.name
    } else if (store.chapters.length > 0) {
      const subjectId = store.chapters[0].subject_id
      await store.fetchSubSubjects(subjectId)
      const found = store.subSubjects.find(s => s.id === id)
      if (found) {
        subSubjectName.value = found.name
      }
    }
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <BasePageHeader
      :title="subSubjectName || '章节列表'"
      :subtitle="'选择章节开始练习'"
    >
      <template #actions>
        <BaseButton type="secondary" @click="router.back()">
          ← 返回
        </BaseButton>
      </template>
    </BasePageHeader>

    <BaseLoading v-if="loading" />

    <BaseEmpty
      v-else-if="error"
      :title="'加载失败'"
      :description="error"
    >
      <template #action>
        <BaseButton @click="load">
          重试
        </BaseButton>
      </template>
    </BaseEmpty>

    <BaseEmpty
      v-else-if="store.chapters.length === 0"
      :title="'暂无章节'"
      :description="'该子科目下还没有章节'"
    />

    <div v-else class="space-y-3">
      <router-link
        v-for="ch in store.chapters"
        :key="ch.id"
        :to="`/practice/chapter/${ch.id}`"
        class="block"
      >
        <BaseCard hover class="flex cursor-pointer items-center justify-between gap-3">
          <div class="flex min-w-0 items-center gap-3">
            <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-50 text-sm font-semibold text-indigo-600">
              {{ ch.sort_order || '' }}
            </div>
            <h3 class="truncate text-sm font-medium text-gray-900">{{ ch.name }}</h3>
          </div>
          <span class="shrink-0 text-sm font-medium text-indigo-600">开始练习 →</span>
        </BaseCard>
      </router-link>
    </div>
  </div>
</template>
