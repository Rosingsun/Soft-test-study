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

function goBackToList() {
  const levelId = store.currentSubject?.level_id
  router.push(levelId ? { path: '/subjects', query: { level_id: levelId } } : '/subjects')
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const id = Number(route.params.id)
    await store.fetchSubSubjects(id)
    if (store.subSubjects.length === 1) {
      router.replace(`/sub-subjects/${store.subSubjects[0].id}/chapters`)
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
      :title="store.currentSubject?.name || '子科目'"
      :subtitle="'选择你要练习的子科目'"
    >
      <template #actions>
        <BaseButton type="secondary" @click="goBackToList">
          ← 返回科目列表
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
      v-else-if="store.subSubjects.length === 0"
      :title="'暂无子科目'"
      :description="'该科目下还没有子科目'"
    />

    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <router-link
        v-for="sub in store.subSubjects"
        :key="sub.id"
        :to="`/sub-subjects/${sub.id}/chapters`"
        class="block"
      >
        <BaseCard hover class="flex h-full cursor-pointer items-center justify-between gap-2">
          <h3 class="min-w-0 truncate text-base font-semibold text-gray-900">{{ sub.name }}</h3>
          <svg class="h-5 w-5 shrink-0 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </BaseCard>
      </router-link>
    </div>
  </div>
</template>
