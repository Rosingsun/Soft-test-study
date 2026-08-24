import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { UserInfo } from '@/types/user'
import * as authApi from '@/api/auth'
import { useSubjectStore } from '@/stores/subject'
import router from '@/router'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserInfo | null>(null)
  const isLoggedIn = ref(false)
  const initialized = ref(false)

  const selectedLevelId = ref(0)
  const selectedSubjectId = ref(0)
  const selectedDifficulty = ref('')

  function syncSelection() {
    if (user.value) {
      selectedLevelId.value = user.value.level_id || 0
      selectedSubjectId.value = user.value.subject_id || 0
      selectedDifficulty.value = user.value.difficulty || ''
    }
  }

  function getHomeRoute(): string {
    // 默认进入「科目导航-综合知识」
    return '/subjects'
  }

  async function login(username: string, password: string) {
    const res = await authApi.login({ username, password })
    localStorage.setItem('access_token', res.access_token)
    isLoggedIn.value = true
    user.value = res.user_info
    syncSelection()
    await loadUserSubjects()
  }

  async function register(username: string, password: string, confirmPassword: string, email: string, levelId: number, subjectId: number, difficulty?: string, inviteCode?: string) {
    const res = await authApi.register({ username, password, confirm_password: confirmPassword, email, level_id: levelId, subject_id: subjectId, difficulty, invite_code: inviteCode })
    localStorage.setItem('access_token', res.access_token)
    isLoggedIn.value = true
    user.value = res.user_info
    syncSelection()
    await loadUserSubjects()
  }

  async function fetchUserInfo() {
    try {
      user.value = await authApi.getUserInfo()
      isLoggedIn.value = true
      syncSelection()
    } catch {
      logout()
    }
  }

  async function loadUserSubjects() {
    if (!user.value || !user.value.level_id) return
    const subjectStore = useSubjectStore()
    if (subjectStore.levels.length === 0) {
      await subjectStore.fetchLevels()
    }
    await subjectStore.ensureSubjects(user.value.level_id)
  }

  async function checkAuth() {
    const token = localStorage.getItem('access_token')
    if (token) {
      await fetchUserInfo()
      await loadUserSubjects()
    }
    initialized.value = true
  }

  async function updateSubjectLevel(levelId: number, subjectId: number) {
    if (!user.value) return
    await authApi.updateProfile({ level_id: levelId, subject_id: subjectId })
    selectedLevelId.value = levelId
    selectedSubjectId.value = subjectId
    await fetchUserInfo()
    await loadUserSubjects()
  }

  async function updateDifficulty(difficulty: string) {
    if (!user.value) return
    await authApi.updateProfile({ difficulty })
    selectedDifficulty.value = difficulty
    await fetchUserInfo()
  }

  async function updateProfile(payload: {
    nickname?: string
    avatar?: string
    level_id?: number | null
    subject_id?: number | null
    difficulty?: string | null
  }) {
    await authApi.updateProfile(payload)
    await fetchUserInfo()
    if (payload.difficulty !== undefined) {
      selectedDifficulty.value = user.value?.difficulty || ''
    }
    if (payload.level_id !== undefined) {
      selectedLevelId.value = payload.level_id || user.value?.level_id || 0
    }
    if (payload.subject_id !== undefined) {
      selectedSubjectId.value = payload.subject_id || user.value?.subject_id || 0
    }
  }

  function logout() {
    localStorage.removeItem('access_token')
    user.value = null
    isLoggedIn.value = false
    selectedLevelId.value = 0
    selectedSubjectId.value = 0
    selectedDifficulty.value = ''
    router.push('/login')
  }

  return { user, isLoggedIn, initialized, selectedLevelId, selectedSubjectId, selectedDifficulty, getHomeRoute, login, register, fetchUserInfo, checkAuth, updateSubjectLevel, updateDifficulty, updateProfile, logout }
})
