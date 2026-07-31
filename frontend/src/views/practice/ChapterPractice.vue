<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubjectStore } from '@/stores/subject'
import { getQuestionsByChapter } from '@/api/question'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import type { Question } from '@/types/question'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const subjectStore = useSubjectStore()
const chapterId = Number(route.params.chapterId)
const questions = ref<Question[]>([])
const loading = ref(true)
const error = ref('')
const chapterName = ref('')

onMounted(async () => {
  try {
    questions.value = await getQuestionsByChapter(chapterId, auth.selectedDifficulty || undefined)
    const chapter = subjectStore.chapters.find(c => c.id === chapterId)
    chapterName.value = chapter?.name || '章节练习'
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

function goBack() {
  router.back()
}
</script>

<template>
  <div>
    <BasePageHeader title="章节练习" :subtitle="chapterName || `章节 #${chapterId}`" />

    <div v-if="loading">
      <BaseLoading />
    </div>

    <div v-else-if="error" class="rounded-lg bg-white p-6 shadow-sm">
      <p class="text-red-500">{{ error }}</p>
      <BaseButton type="secondary" size="sm" class="mt-4" @click="goBack">返回</BaseButton>
    </div>

    <div v-else-if="questions.length === 0" class="rounded-lg bg-white shadow-sm">
      <BaseEmpty title="该章节暂无题目" description="请稍后再来试试">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="goBack">返回</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <PracticeRunner v-else :questions="questions" mode="chapter" @back="goBack" />
  </div>
</template>
