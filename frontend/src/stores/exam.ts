import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { StartExamResp } from '@/types/exam'

export const useExamStore = defineStore('exam', () => {
  const aiExamData = ref<StartExamResp | null>(null)

  function setAiExamData(data: StartExamResp) {
    aiExamData.value = data
  }

  function getAndClearAiExamData(): StartExamResp | null {
    const data = aiExamData.value
    aiExamData.value = null
    return data
  }

  return { aiExamData, setAiExamData, getAndClearAiExamData }
})
