<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { Subject } from '@/types/subject'
import { useSubjectStore } from '@/stores/subject'
import { useAuthStore } from '@/stores/auth'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseButton from '@/components/common/BaseButton.vue'

const store = useSubjectStore()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const loading = ref(true)
const error = ref('')

function subjectIcon(subject: Subject) {
  return subject.icon || subject.name.charAt(0)
}

function maybeRedirect(levelId: number) {
  if (auth.isLoggedIn && auth.selectedSubjectId) {
    const matched = store.subjects.find(s => s.id === auth.selectedSubjectId)
    if (matched && matched.level_id === levelId) {
      router.replace(`/subjects/${auth.selectedSubjectId}`)
    }
  }
}

async function loadSubjects(levelId: number) {
  loading.value = true
  error.value = ''
  try {
    await store.fetchSubjectsByLevel(levelId)
    maybeRedirect(levelId)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function init() {
  if (!auth.initialized) {
    await auth.checkAuth()
  }
  if (store.levels.length === 0) {
    try {
      await store.fetchLevels()
    } catch {
      /* noop */
    }
  }
  const levelId = auth.selectedLevelId || Number(route.query.level_id)
  if (!levelId) {
    loading.value = false
    return
  }
  await loadSubjects(levelId)
}

onMounted(init)

watch(() => route.query.level_id, (id) => {
  if (id) {
    loadSubjects(Number(id))
  }
})
</script>

<template>
  <div>
    <BasePageHeader
      :title="'科目导航'"
      :subtitle="store.currentLevel?.name || '选择考试科目'"
    >
      <template v-if="!auth.isLoggedIn" #actions>
        <BaseButton type="secondary" @click="router.push('/')">
          返回选择等级
        </BaseButton>
      </template>
    </BasePageHeader>

    <BaseSkeleton v-if="loading" variant="grid" :count="6" />

    <BaseEmpty
      v-else-if="error"
      :title="'加载失败'"
      :description="error"
    >
      <template #action>
        <BaseButton @click="loadSubjects(auth.selectedLevelId || Number(route.query.level_id))">
          重试
        </BaseButton>
      </template>
    </BaseEmpty>

    <BaseEmpty
      v-else-if="!auth.selectedLevelId"
      :title="'请先选择考试等级'"
      :description="'请在右上角选择你的考试等级，开始备考'"
    />

    <BaseEmpty
      v-else-if="store.subjects.length === 0"
      :title="'该等级暂无考试科目'"
      :description="'换个等级试试吧'"
    />

    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <router-link
        v-for="subject in store.subjects"
        :key="subject.id"
        :to="`/subjects/${subject.id}`"
        class="block"
      >
        <BaseCard hover class="h-full cursor-pointer">
          <div class="flex items-start gap-3">
            <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-indigo-50 text-xl text-indigo-600">
              {{ subjectIcon(subject) }}
            </div>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="text-base font-semibold text-gray-900">{{ subject.name }}</h3>
                <BaseBadge v-if="subject.short_name" type="info">
                  {{ subject.short_name }}
                </BaseBadge>
              </div>
              <p v-if="subject.description" class="mt-1 line-clamp-2 text-sm text-gray-500">
                {{ subject.description }}
              </p>
            </div>
          </div>
        </BaseCard>
      </router-link>
    </div>
  </div>
</template>
