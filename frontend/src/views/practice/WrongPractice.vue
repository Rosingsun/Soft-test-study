<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listWrongQuestions } from '@/api/wrong_question'
import { wrongPracticeSubmit, removeWrongQuestion } from '@/api/wrong_question'
import { getQuestion } from '@/api/question'
import PracticeRunner from '@/components/practice/PracticeRunner.vue'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import type { Question } from '@/types/question'

const route = useRoute()
const router = useRouter()
const singleQid = Number(route.query.qid) || 0
const questions = ref<Question[]>([])
const startIndex = ref(0)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const data = await listWrongQuestions()
    const target = singleQid ? data.filter(q => q.question_id === singleQid) : data
    const loaded: Question[] = []
    const wanted: number[] = []
    for (const w of target) {
      try {
        const q = await getQuestion(w.question_id)
        loaded.push(q)
        wanted.push(q.id)
      } catch {
        // 题目可能已被删除，跳过
      }
    }
    questions.value = loaded
    if (singleQid) {
      const idx = wanted.indexOf(singleQid)
      startIndex.value = idx >= 0 ? idx : 0
    }
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

async function submitHandler(questionId: number, answer: string, duration: number) {
  const res = await wrongPracticeSubmit({ question_id: questionId, answer, duration })
  if (res.is_correct === 1) {
    setTimeout(async () => {
      try {
        await removeWrongQuestion(questionId)
      } catch {
        // 自动移出失败时不打断答题流程
      }
    }, 2000)
  }
  return res
}

function goBack() {
  router.push('/wrong-questions')
}
</script>

<template>
  <div>
    <BasePageHeader
      :title="singleQid ? '错题重做' : '错题练习'"
      subtitle="连续答对会自动移出错题本，作答后可直接继续下一题"
    />

    <BaseSkeleton v-if="loading" variant="question" />

    <div v-else-if="error" class="rounded-lg bg-white p-6 shadow-sm">
      <p class="text-red-500">{{ error }}</p>
      <BaseButton type="secondary" size="sm" class="mt-4" @click="goBack">返回错题本</BaseButton>
    </div>

    <div v-else-if="questions.length === 0" class="rounded-lg bg-white shadow-sm">
      <BaseEmpty title="暂无错题" description="继续保持，错误会越来越少">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="goBack">返回错题本</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <PracticeRunner
      v-else
      :questions="questions"
      :start-index="startIndex"
      mode="wrong"
      :submit-handler="submitHandler"
      @back="goBack"
    />
  </div>
</template>
