import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ExamLevel, Subject, SubSubject, Chapter } from '@/types/subject'
import * as subjectApi from '@/api/subject'

export const useSubjectStore = defineStore('subject', () => {
  const levels = ref<ExamLevel[]>([])
  const currentLevel = ref<ExamLevel | null>(null)
  const subjects = ref<Subject[]>([])
  const currentSubject = ref<Subject | null>(null)
  const subSubjects = ref<SubSubject[]>([])
  const chapters = ref<Chapter[]>([])

  async function fetchLevels() {
    levels.value = await subjectApi.getExamLevels()
  }

  async function fetchSubjectsByLevel(levelId: number) {
    subjects.value = await subjectApi.getSubjectsByLevel(levelId)
    currentLevel.value = levels.value.find(l => l.id === levelId) || null
  }

  async function ensureSubjects(levelId: number) {
    if (subjects.value.length === 0 || subjects.value[0]?.level_id !== levelId) {
      await fetchSubjectsByLevel(levelId)
    }
  }

  async function fetchSubSubjects(subjectId: number) {
    subSubjects.value = await subjectApi.getSubSubjects(subjectId)
    currentSubject.value = subjects.value.find(s => s.id === subjectId) || null
  }

  async function fetchChapters(subSubjectId: number) {
    chapters.value = await subjectApi.getChapters(subSubjectId)
  }

  function reset() {
    currentLevel.value = null
    subjects.value = []
    currentSubject.value = null
    subSubjects.value = []
    chapters.value = []
  }

  return {
    levels, currentLevel, subjects, currentSubject, subSubjects, chapters,
    fetchLevels, fetchSubjectsByLevel, ensureSubjects, fetchSubSubjects, fetchChapters, reset,
  }
})
