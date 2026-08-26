<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubjectStore } from '@/stores/subject'
import { useExamStore } from '@/stores/exam'
import { listExamTemplates } from '@/api/exam'
import { startAiExam } from '@/api/ai'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import type { ExamTemplateResp } from '@/types/exam'

const auth = useAuthStore()
const subjectStore = useSubjectStore()
const examStore = useExamStore()
const router = useRouter()
const templates = ref<ExamTemplateResp[]>([])
const loading = ref(true)
const error = ref('')

const aiModalOpen = ref(false)
const aiCount = ref(10)
const aiDuration = ref(20)
const aiGenerating = ref(false)
const aiError = ref('')

const filteredTemplates = computed(() => templates.value.filter(t => t.subject_id === auth.selectedSubjectId))

const countOptions = [5, 10, 15, 20, 30, 50, 75]

onMounted(async () => {
  try {
    templates.value = await listExamTemplates()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

function goExam(t: ExamTemplateResp) {
  router.push({ path: `/exam/${t.id}`, query: { source: 'template' } })
}

function openAiModal() {
  aiCount.value = 10
  aiDuration.value = 20
  aiError.value = ''
  aiModalOpen.value = true
}

async function startAiExamHandler() {
  if (!auth.selectedSubjectId) {
    aiError.value = '请选择科目'
    return
  }
  aiGenerating.value = true
  aiError.value = ''
  try {
    const data = await startAiExam({
      subject_id: auth.selectedSubjectId,
      count: aiCount.value,
      duration: aiDuration.value,
    })
    examStore.setAiExamData(data)
    aiModalOpen.value = false
    router.push(`/exam/${data.record_id}?source=ai`)
  } catch (e) {
    aiError.value = (e as Error).message
  } finally {
    aiGenerating.value = false
  }
}
</script>

<template>
  <div>
    <BasePageHeader title="模拟考试" subtitle="全真模拟，检验备考成果">
      <template #actions>
        <button
          class="cursor-pointer rounded-lg bg-white border border-gray-300 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50"
          @click="router.push('/exam-records')"
        >
          考试记录
        </button>
      </template>
    </BasePageHeader>

    <BaseSkeleton v-if="loading" variant="grid" :count="6" />

    <div v-else-if="error" class="rounded-xl border border-red-100 bg-red-50 p-6">
      <p class="text-sm text-red-600">{{ error }}</p>
    </div>

    <div v-else-if="filteredTemplates.length === 0" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="暂无可用试卷" description="题目数量达标后会自动生成模拟试卷" />
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <!-- AI 智能组卷卡片 -->
      <div class="card-lift flex flex-col rounded-xl border-2 border-emerald-200 bg-gradient-to-br from-emerald-50/50 to-white p-5 shadow-sm">
        <div class="mb-3 flex items-start justify-between gap-2">
          <div class="flex items-center gap-2">
            <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-100">
              <svg class="h-4 w-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5" />
              </svg>
            </div>
            <h3 class="text-base font-semibold text-gray-900">AI 智能组卷</h3>
          </div>
          <BaseBadge>AI</BaseBadge>
        </div>
        <p class="mb-4 text-xs text-gray-500">从你在「AI 练习」中已生成的题目里随机抽题组卷，无需实时调用 AI</p>

        <div class="mb-4 grid grid-cols-3 gap-2 text-center">
          <div class="rounded-lg bg-emerald-50 py-2">
            <p class="text-sm font-bold text-emerald-700">已有</p>
            <p class="text-[10px] text-gray-400">AI 题库</p>
          </div>
          <div class="rounded-lg bg-emerald-50 py-2">
            <p class="text-sm font-bold text-emerald-700">随机</p>
            <p class="text-[10px] text-gray-400">抽题</p>
          </div>
          <div class="rounded-lg bg-emerald-50 py-2">
            <p class="text-sm font-bold text-emerald-700">即时</p>
            <p class="text-[10px] text-gray-400">组卷</p>
          </div>
        </div>

        <button
          class="mt-auto cursor-pointer rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-emerald-700"
          @click="openAiModal"
        >
          AI 智能组卷
        </button>
      </div>

      <div
        v-for="t in filteredTemplates"
        :key="t.id"
        class="card-lift flex flex-col rounded-xl border border-gray-100 bg-white p-5 shadow-sm"
      >
        <div class="mb-3 flex items-start justify-between gap-2">
          <h3 class="text-base font-semibold tracking-tight text-gray-900">{{ t.name }}</h3>
          <BaseBadge v-if="t.year" type="info">{{ t.year }}年</BaseBadge>
        </div>
        <p class="mb-4 text-xs text-gray-400">
          {{ subjectStore.subjects.find(s => s.id === t.subject_id)?.name || `科目 #${t.subject_id}` }}
        </p>

        <div class="mb-4 grid grid-cols-3 gap-2 text-center">
          <div class="rounded-xl bg-gray-50 py-2.5">
            <p class="text-sm font-bold text-gray-800">{{ t.question_count }}</p>
            <p class="text-[10px] text-gray-400">题目</p>
          </div>
          <div class="rounded-xl bg-gray-50 py-2.5">
            <p class="text-sm font-bold text-gray-800">{{ t.total_score }}</p>
            <p class="text-[10px] text-gray-400">总分</p>
          </div>
          <div class="rounded-xl bg-gray-50 py-2.5">
            <p class="text-sm font-bold text-gray-800">{{ t.duration }}分</p>
            <p class="text-[10px] text-gray-400">时长</p>
          </div>
        </div>

        <button
          class="bg-brand-gradient mt-auto cursor-pointer rounded-lg px-4 py-2 text-sm font-semibold text-white shadow-sm transition-all duration-200 hover:shadow-md hover:brightness-110 active:scale-[0.98]"
          @click="goExam(t)"
        >
          开始考试
        </button>
      </div>
    </div>

    <!-- AI 组卷弹窗（从已有 AI 题库抽题） -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="aiModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="aiModalOpen = false">
          <div class="w-full max-w-md rounded-xl bg-white shadow-2xl" @click.stop>
            <div class="flex items-center justify-between border-b border-gray-100 px-5 py-4">
              <h2 class="text-lg font-semibold text-gray-900">AI 智能组卷</h2>
              <button class="cursor-pointer text-gray-400 hover:text-gray-600" @click="aiModalOpen = false">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div class="space-y-4 px-5 py-4">
              <div class="rounded-lg bg-blue-50 px-3 py-2 text-xs text-blue-700">
                从你在「AI 练习」中已生成的题目库里，按科目随机抽取题目组成试卷。如需新题，请先去「AI 练习」生成。
              </div>

              <div v-if="aiError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{{ aiError }}</div>

              <div>
                <label class="mb-1.5 block text-sm font-medium text-gray-700">题目数量</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="n in countOptions"
                    :key="n"
                    class="cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors"
                    :class="aiCount === n
                      ? 'border-emerald-500 bg-emerald-50 text-emerald-600'
                      : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
                    @click="aiCount = n"
                  >
                    {{ n }} 题
                  </button>
                </div>
              </div>

              <div>
                <label class="mb-1.5 block text-sm font-medium text-gray-700">考试时长（分钟）</label>
                <input
                  v-model.number="aiDuration"
                  type="number"
                  min="5"
                  max="180"
                  class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none"
                />
              </div>
            </div>

            <div class="flex justify-end gap-2 border-t border-gray-100 px-5 py-4">
              <button
                class="cursor-pointer rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm text-gray-600 hover:bg-gray-50"
                @click="aiModalOpen = false"
              >
                取消
              </button>
              <button
                class="cursor-pointer rounded-lg bg-emerald-600 px-4 py-2 text-sm text-white hover:bg-emerald-700 disabled:opacity-50"
                :disabled="aiGenerating"
                @click="startAiExamHandler"
              >
                {{ aiGenerating ? '组卷中...' : '开始考试' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
