<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useSubjectStore } from '@/stores/subject'
import { changePassword, sendEmailCode, verifyEmailCode } from '@/api/auth'
import { getStatsOverview, getDailyStats } from '@/api/stats'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseTabs from '@/components/common/BaseTabs.vue'
import BarChart from '@/components/charts/BarChart.vue'
import { showToast } from '@/utils/toast'
import type { StatsOverviewResp, DailyStatsResp } from '@/types/stats'
import type { EmailPurpose } from '@/types/user'

const auth = useAuthStore()
const subjectStore = useSubjectStore()
const overview = ref<StatsOverviewResp | null>(null)
const dailyStats = ref<DailyStatsResp[]>([])
const loading = ref(true)
const error = ref('')

type TabKey = 'stats' | 'security'
const activeTab = ref<TabKey>('stats')

const showEdit = ref(false)
const nickname = ref('')
const avatar = ref('')
const levelId = ref<number>(0)
const subjectId = ref<number>(0)
const difficulty = ref('')
const saving = ref(false)

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const pwdSaving = ref(false)
const pwdError = ref('')
const showOldPwd = ref(false)
const showNewPwd = ref(false)
const showConfirmPwd = ref(false)

// 邮箱绑定弹窗
const showEmailModal = ref(false)
const emailPurpose = ref<EmailPurpose>('verify')
const emailTarget = ref('')
const emailCodeDigits = ref<string[]>(['', '', '', '', '', ''])
const emailCodeRefs = ref<(HTMLInputElement | null)[]>([])
const emailSending = ref(false)
const emailVerifying = ref(false)
const emailCooldown = ref(0)
const emailError = ref('')
let emailCooldownTimer: ReturnType<typeof setInterval> | null = null

const difficultyOptions = [
  { value: 'easy', label: '简单' },
  { value: 'medium', label: '中等' },
  { value: 'hard', label: '困难' },
]
const difficultyLabelMap: Record<string, { label: string; type: 'success' | 'warning' | 'danger' }> = {
  easy: { label: '简单', type: 'success' },
  medium: { label: '中等', type: 'warning' },
  hard: { label: '困难', type: 'danger' },
}

const roleLabel = computed(() => auth.user?.role === 'admin' ? '管理员' : '学生')

const sortedDaily = computed(() => [...dailyStats.value].sort((a, b) => b.date.localeCompare(a.date)))

const chartData = computed(() => {
  const now = new Date()
  const map = new Map(dailyStats.value.map(d => [d.date.slice(0, 10), d]))
  const days: { label: string; value: number; hint: string }[] = []
  for (let i = 29; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    const count = map.get(key)?.total_count || 0
    days.push({
      label: `${d.getMonth() + 1}/${d.getDate()}`,
      value: count,
      hint: '题',
    })
  }
  return days
})

const pwdStrength = computed(() => {
  const p = newPassword.value
  if (!p) return { level: 0, label: '', color: 'bg-gray-200' }
  let score = 0
  if (p.length >= 6) score++
  if (p.length >= 10) score++
  if (/[A-Z]/.test(p) && /[a-z]/.test(p)) score++
  if (/\d/.test(p)) score++
  if (/[^\w\s]/.test(p)) score++
  const level = Math.min(4, Math.max(0, score - 1))
  if (level <= 1) return { level, label: '弱', color: 'bg-red-400' }
  if (level === 2) return { level, label: '中等', color: 'bg-amber-400' }
  return { level, label: '强', color: 'bg-emerald-500' }
})

watch(() => auth.user, u => {
  if (u && levelId.value === 0) levelId.value = u.level_id || 0
})

const emailCode = computed(() => emailCodeDigits.value.join(''))
const isEmailCodeComplete = computed(() => emailCodeDigits.value.every(d => d !== ''))
const isEmailValid = computed(() => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailTarget.value.trim()))
const isEmailCooldown = computed(() => emailCooldown.value > 0)

function openEmailModal(purpose: EmailPurpose) {
  emailPurpose.value = purpose
  emailTarget.value = auth.user?.email || ''
  emailCodeDigits.value = ['', '', '', '', '', '']
  emailError.value = ''
  emailCooldown.value = 0
  showEmailModal.value = true
  nextTick(() => {
    emailCodeRefs.value[0]?.focus()
  })
}

function closeEmailModal() {
  showEmailModal.value = false
}

async function onSendCode() {
  if (!isEmailValid.value) {
    emailError.value = '请输入有效的邮箱地址'
    return
  }
  emailSending.value = true
  emailError.value = ''
  try {
    await sendEmailCode({ email: emailTarget.value.trim(), purpose: emailPurpose.value })
    showToast('验证码已发送，请查收邮箱', 'success')
    startEmailCooldown()
  } catch {
    // toast 由 request 层提示
  } finally {
    emailSending.value = false
  }
}

function startEmailCooldown() {
  emailCooldown.value = 60
  if (emailCooldownTimer) clearInterval(emailCooldownTimer)
  emailCooldownTimer = setInterval(() => {
    emailCooldown.value -= 1
    if (emailCooldown.value <= 0 && emailCooldownTimer) {
      clearInterval(emailCooldownTimer)
      emailCooldownTimer = null
    }
  }, 1000)
}

function onCodeInput(idx: number, e: Event) {
  const target = e.target as HTMLInputElement
  const v = target.value.replace(/\D/g, '')
  emailCodeDigits.value[idx] = v.slice(-1)
  if (v && idx < 5) {
    emailCodeRefs.value[idx + 1]?.focus()
  }
}

function onCodeKeydown(idx: number, e: KeyboardEvent) {
  if (e.key === 'Backspace' && !emailCodeDigits.value[idx] && idx > 0) {
    emailCodeRefs.value[idx - 1]?.focus()
  }
}

function onCodePaste(e: ClipboardEvent) {
  const text = e.clipboardData?.getData('text') || ''
  const digits = text.replace(/\D/g, '').slice(0, 6).split('')
  if (!digits.length) return
  e.preventDefault()
  for (let i = 0; i < 6; i++) {
    emailCodeDigits.value[i] = digits[i] || ''
  }
  const nextEmpty = emailCodeDigits.value.findIndex(d => !d)
  const focusIdx = nextEmpty === -1 ? 5 : nextEmpty
  emailCodeRefs.value[focusIdx]?.focus()
}

async function onConfirmEmail() {
  if (!isEmailValid.value) {
    emailError.value = '请输入有效的邮箱地址'
    return
  }
  if (!isEmailCodeComplete.value) {
    emailError.value = '请输入完整的 6 位验证码'
    return
  }
  emailVerifying.value = true
  emailError.value = ''
  try {
    await verifyEmailCode({
      email: emailTarget.value.trim(),
      code: emailCode.value,
      purpose: emailPurpose.value,
    })
    showToast(emailPurpose.value === 'change' ? '邮箱换绑成功' : '邮箱验证成功', 'success')
    await auth.fetchUserInfo()
    closeEmailModal()
  } catch {
    // toast 由 request 层提示
  } finally {
    emailVerifying.value = false
  }
}

function scrollToEmailCard() {
  activeTab.value = 'security'
  nextTick(() => {
    const el = document.getElementById('email-binding-card')
    el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}

async function openEdit() {
  nickname.value = auth.user?.nickname || ''
  avatar.value = auth.user?.avatar || ''
  difficulty.value = auth.user?.difficulty || ''
  levelId.value = auth.user?.level_id || 0
  subjectId.value = auth.user?.subject_id || 0

  if (subjectStore.levels.length === 0) {
    try {
      await subjectStore.fetchLevels()
    } catch {
      showToast('加载考试级别失败', 'error')
    }
  }

  if (levelId.value && (subjectStore.subjects.length === 0 || subjectStore.subjects[0]?.level_id !== levelId.value)) {
    try {
      await subjectStore.fetchSubjectsByLevel(levelId.value)
    } catch {
      showToast('加载科目失败', 'error')
    }
  }

  showEdit.value = true
}

watch(levelId, async newId => {
  if (!newId) return
  subjectId.value = 0
  if (subjectStore.levels.length === 0) {
    try { await subjectStore.fetchLevels() } catch { /* noop */ }
  }
  try {
    await subjectStore.fetchSubjectsByLevel(newId)
  } catch {
    showToast('加载科目失败', 'error')
  }
})

async function saveProfile() {
  if (!nickname.value.trim()) {
    showToast('请输入昵称', 'error')
    return
  }
  saving.value = true
  try {
    await auth.updateProfile({
      nickname: nickname.value.trim(),
      avatar: avatar.value.trim() || undefined,
      difficulty: difficulty.value || null,
      level_id: levelId.value || null,
      subject_id: subjectId.value || null,
    })
    showEdit.value = false
    showToast('保存成功', 'success')
  } catch {
    // toast 由 request 层提示
  } finally {
    saving.value = false
  }
}

function resetPwdForm() {
  oldPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
  pwdError.value = ''
  showOldPwd.value = false
  showNewPwd.value = false
  showConfirmPwd.value = false
}

async function savePassword() {
  pwdError.value = ''
  if (!oldPassword.value || !newPassword.value || !confirmPassword.value) {
    pwdError.value = '请填写完整密码信息'
    return
  }
  if (newPassword.value.length < 6) {
    pwdError.value = '新密码至少 6 位'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    pwdError.value = '两次输入的新密码不一致'
    return
  }
  if (newPassword.value === oldPassword.value) {
    pwdError.value = '新密码不能与原密码相同'
    return
  }
  pwdSaving.value = true
  try {
    await changePassword({ old_password: oldPassword.value, new_password: newPassword.value })
    showToast('密码修改成功', 'success')
    resetPwdForm()
  } catch {
    // toast 由 request 层提示
  } finally {
    pwdSaving.value = false
  }
}

function accuracyClass(a: number) {
  if (a >= 80) return 'text-emerald-500'
  if (a >= 60) return 'text-indigo-600'
  return 'text-yellow-500'
}

function avatarBgClass(seed: string) {
  const colors = [
    'bg-indigo-100 text-indigo-600',
    'bg-violet-100 text-violet-600',
    'bg-emerald-100 text-emerald-600',
    'bg-amber-100 text-amber-600',
    'bg-rose-100 text-rose-600',
  ]
  let hash = 0
  for (let i = 0; i < seed.length; i++) hash = (hash * 31 + seed.charCodeAt(i)) >>> 0
  return colors[hash % colors.length]
}

onMounted(async () => {
  try {
    const [ov, daily] = await Promise.all([getStatsOverview(), getDailyStats(30)])
    overview.value = ov
    dailyStats.value = daily || []
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => {
  if (emailCooldownTimer) {
    clearInterval(emailCooldownTimer)
    emailCooldownTimer = null
  }
})
</script>

<template>
  <div class="mx-auto max-w-5xl">
    <BasePageHeader title="个人中心" subtitle="管理个人信息、账号安全与学习数据" />

    <BaseSkeleton v-if="loading" variant="detail" />

    <div v-else-if="error">
      <BaseCard>
        <div class="flex items-center gap-3 text-red-500">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
          </svg>
          <p class="text-sm">{{ error }}</p>
        </div>
      </BaseCard>
    </div>

    <template v-else-if="auth.user">
      <!-- 顶部个人信息卡 -->
      <BaseCard :padding="false" class="mb-6">
        <div class="bg-brand-gradient relative h-28 overflow-hidden rounded-t-xl sm:h-32">
          <div class="absolute -right-10 -top-10 h-40 w-40 rounded-full bg-white/10" />
          <div class="absolute -bottom-16 right-24 h-48 w-48 rounded-full bg-white/10" />
          <div class="absolute -left-8 top-1/2 h-24 w-24 -translate-y-1/2 rounded-full bg-white/5" />
        </div>
        <div class="px-6 pb-6 sm:px-8">
          <div class="-mt-12 flex flex-col items-center gap-4 sm:-mt-14 sm:flex-row sm:items-end">
            <div
              class="relative z-10 flex h-24 w-24 shrink-0 items-center justify-center overflow-hidden rounded-2xl border-4 border-white text-2xl font-bold shadow-md sm:h-28 sm:w-28"
              :class="auth.user.avatar ? 'bg-white' : avatarBgClass(auth.user.username || auth.user.nickname || 'U')"
            >
              <img v-if="auth.user.avatar" :src="auth.user.avatar" :alt="auth.user.nickname" class="h-full w-full object-cover" />
              <template v-else>{{ (auth.user.nickname || auth.user.username || 'U').charAt(0).toUpperCase() }}</template>
            </div>
            <div class="flex-1 text-center sm:text-left">
              <div class="flex flex-wrap items-center justify-center gap-2 sm:justify-start">
                <h1 class="text-xl font-bold tracking-tight text-gray-900">{{ auth.user.nickname || auth.user.username }}</h1>
                <BaseBadge :type="auth.user.role === 'admin' ? 'info' : 'default'">{{ roleLabel }}</BaseBadge>
              </div>
              <button
                type="button"
                class="mt-1 inline-flex cursor-pointer items-center gap-1.5 rounded-md transition-colors hover:bg-gray-100/60"
                :class="!auth.user.email_verified ? 'pr-1.5' : ''"
                @click="scrollToEmailCard"
              >
                <span class="text-sm text-gray-500">{{ auth.user.email }}</span>
                <BaseBadge v-if="auth.user.email_verified" type="success" dot>已验证</BaseBadge>
                <BaseBadge v-else type="warning" dot>未验证</BaseBadge>
              </button>
              <div class="mt-3 flex flex-wrap justify-center gap-1.5 sm:justify-start">
                <BaseBadge type="info">{{ auth.user.level_name || '未选择级别' }}</BaseBadge>
                <BaseBadge type="success">{{ auth.user.subject_name || '未选择科目' }}</BaseBadge>
                <BaseBadge v-if="auth.user.difficulty && difficultyLabelMap[auth.user.difficulty]" :type="difficultyLabelMap[auth.user.difficulty].type">
                  难度偏好：{{ difficultyLabelMap[auth.user.difficulty].label }}
                </BaseBadge>
              </div>
            </div>
            <div class="flex flex-wrap justify-center gap-2 sm:justify-end">
              <BaseButton type="secondary" size="sm" @click="openEdit">
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                </svg>
                编辑资料
              </BaseButton>
              <BaseButton type="secondary" size="sm" @click="activeTab = 'security'">
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
                </svg>
                修改密码
              </BaseButton>
            </div>
          </div>
        </div>
      </BaseCard>

      <!-- 分段切换 -->
      <div class="mb-6 flex justify-center sm:justify-start">
        <BaseTabs
          v-model="activeTab"
          :tabs="[
            { label: '学习数据', value: 'stats' },
            { label: '账号安全', value: 'security' },
          ]"
        />
      </div>

      <!-- 学习数据 -->
      <template v-if="activeTab === 'stats'">
        <!-- 概览卡片 -->
        <div v-if="overview" class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
          <BaseCard hover>
            <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-50">
              <svg class="h-5 w-5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
              </svg>
            </div>
            <p class="text-2xl font-bold tracking-tight text-gray-900">{{ overview.total_practiced }}</p>
            <p class="mt-1 text-xs text-gray-400">累计答题</p>
          </BaseCard>
          <BaseCard hover>
            <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-50">
              <svg class="h-5 w-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <p class="text-2xl font-bold tracking-tight" :class="accuracyClass(overview.accuracy)">{{ overview.accuracy.toFixed(1) }}%</p>
            <p class="mt-1 text-xs text-gray-400">总体正确率</p>
          </BaseCard>
          <BaseCard hover>
            <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-violet-50">
              <svg class="h-5 w-5 text-violet-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0021 11.25v7.5" />
              </svg>
            </div>
            <p class="text-2xl font-bold tracking-tight text-gray-900">{{ overview.study_days }}</p>
            <p class="mt-1 text-xs text-gray-400">累计学习天数</p>
          </BaseCard>
          <BaseCard hover>
            <div class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-red-50">
              <svg class="h-5 w-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
              </svg>
            </div>
            <p class="text-2xl font-bold tracking-tight text-gray-900">{{ overview.wrong_count }}</p>
            <p class="mt-1 text-xs text-gray-400">待复习错题</p>
          </BaseCard>
        </div>

        <!-- 答题趋势图 -->
        <BaseCard class="mb-6">
          <div class="mb-4 flex items-center justify-between">
            <div>
              <h2 class="text-base font-semibold tracking-tight text-gray-900">近 30 天答题趋势</h2>
              <p class="mt-0.5 text-xs text-gray-400">每日答题量分布</p>
            </div>
            <BaseBadge type="info">共 {{ dailyStats.reduce((s, d) => s + d.total_count, 0) }} 题</BaseBadge>
          </div>
          <BarChart :data="chartData" :height="160" />
        </BaseCard>

        <!-- 每日学习记录 -->
        <BaseCard>
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-base font-semibold tracking-tight text-gray-900">每日学习记录</h2>
            <span class="text-xs text-gray-400">最近 30 天</span>
          </div>
          <BaseEmpty
            v-if="sortedDaily.length === 0"
            title="暂无学习记录"
            description="完成第一次练习，记录就会出现在这里"
          >
            <template #action>
              <BaseButton type="primary" @click="$router.push('/practice/random')">开始练习</BaseButton>
            </template>
          </BaseEmpty>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-100 text-left text-xs text-gray-500">
                  <th class="pb-2.5 font-medium">日期</th>
                  <th class="pb-2.5 font-medium">答题数</th>
                  <th class="pb-2.5 font-medium">正确数</th>
                  <th class="pb-2.5 font-medium">错误数</th>
                  <th class="pb-2.5 font-medium">正确率</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in sortedDaily"
                  :key="row.date"
                  class="border-b border-gray-50 transition-colors hover:bg-gray-50/60"
                >
                  <td class="py-2.5 text-gray-700">{{ row.date }}</td>
                  <td class="py-2.5 text-gray-700">{{ row.total_count }}</td>
                  <td class="py-2.5 text-emerald-600">{{ row.correct_count }}</td>
                  <td class="py-2.5 text-red-500">{{ row.incorrect_count }}</td>
                  <td class="py-2.5">
                    <span v-if="row.total_count === 0" class="text-xs text-gray-400">-</span>
                    <span
                      v-else
                      class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset"
                      :class="(row.correct_count / row.total_count * 100) >= 60
                        ? 'bg-emerald-50 text-emerald-700 ring-emerald-600/20'
                        : 'bg-red-50 text-red-700 ring-red-600/20'"
                    >
                      {{ (row.correct_count / row.total_count * 100).toFixed(0) }}%
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </BaseCard>
      </template>

      <!-- 账号安全 -->
      <template v-else>
        <div class="grid gap-6 lg:grid-cols-5">
          <!-- 账号信息 -->
          <BaseCard class="lg:col-span-2">
            <div class="mb-4 flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-50">
                <svg class="h-5 w-5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 9h3.75M15 12h3.75M15 15h3.75M4.5 19.5h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5zm6-10.125a1.875 1.875 0 11-3.75 0 1.875 1.875 0 013.75 0zm1.294 6.336a6.721 6.721 0 01-3.17.789 6.721 6.721 0 01-3.168-.789 3.376 3.376 0 016.338 0z" />
                </svg>
              </div>
              <div>
                <h2 class="text-base font-semibold tracking-tight text-gray-900">账号信息</h2>
                <p class="text-xs text-gray-400">基础账户信息</p>
              </div>
            </div>
            <dl class="space-y-3.5">
              <div class="flex items-center justify-between gap-3 text-sm">
                <dt class="text-gray-500">用户名</dt>
                <dd class="font-medium text-gray-800">{{ auth.user.username }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3 text-sm">
                <dt class="shrink-0 text-gray-500">邮箱</dt>
                <dd class="flex min-w-0 items-center justify-end gap-2">
                  <span class="truncate font-medium text-gray-800">{{ auth.user.email }}</span>
                  <BaseBadge v-if="auth.user.email_verified" type="success" dot>已验证</BaseBadge>
                  <BaseBadge v-else type="warning" dot>未验证</BaseBadge>
                </dd>
              </div>
              <div class="flex items-center justify-between gap-3 text-sm">
                <dt class="text-gray-500">角色</dt>
                <dd>
                  <BaseBadge :type="auth.user.role === 'admin' ? 'info' : 'default'">{{ roleLabel }}</BaseBadge>
                </dd>
              </div>
              <div class="flex items-center justify-between gap-3 text-sm">
                <dt class="text-gray-500">所属级别</dt>
                <dd class="font-medium text-gray-800">{{ auth.user.level_name || '-' }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3 text-sm">
                <dt class="text-gray-500">所属科目</dt>
                <dd class="font-medium text-gray-800">{{ auth.user.subject_name || '-' }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3 text-sm">
                <dt class="text-gray-500">账号 ID</dt>
                <dd class="font-mono text-gray-800">#{{ auth.user.id }}</dd>
              </div>
            </dl>
          </BaseCard>

          <!-- 修改密码 -->
          <BaseCard class="lg:col-span-3">
            <div class="mb-4 flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-50">
                <svg class="h-5 w-5 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
                </svg>
              </div>
              <div>
                <h2 class="text-base font-semibold tracking-tight text-gray-900">修改密码</h2>
                <p class="text-xs text-gray-400">建议每 3 个月更换一次密码</p>
              </div>
            </div>

            <div v-if="pwdError" class="mb-4 flex items-start gap-2 rounded-lg border border-red-100 bg-red-50/70 px-3.5 py-2.5 text-sm text-red-600">
              <svg class="mt-0.5 h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
              </svg>
              <span>{{ pwdError }}</span>
            </div>

            <div class="space-y-4">
              <div>
                <label class="mb-1.5 block text-sm font-medium text-gray-700">原密码</label>
                <div class="relative">
                  <input
                    v-model="oldPassword"
                    :type="showOldPwd ? 'text' : 'password'"
                    placeholder="请输入原密码"
                    autocomplete="current-password"
                    class="w-full rounded-lg border border-gray-300 bg-white px-3.5 py-2 pr-10 text-sm text-gray-800 shadow-sm transition-colors placeholder:text-gray-400 hover:border-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                  />
                  <button
                    type="button"
                    class="absolute inset-y-0 right-0 flex cursor-pointer items-center px-3 text-gray-400 hover:text-gray-600"
                    @click="showOldPwd = !showOldPwd"
                  >
                    <svg v-if="showOldPwd" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                      <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
                    </svg>
                  </button>
                </div>
              </div>
              <div>
                <label class="mb-1.5 block text-sm font-medium text-gray-700">新密码</label>
                <div class="relative">
                  <input
                    v-model="newPassword"
                    :type="showNewPwd ? 'text' : 'password'"
                    placeholder="至少 6 位，建议字母+数字+符号"
                    autocomplete="new-password"
                    class="w-full rounded-lg border border-gray-300 bg-white px-3.5 py-2 pr-10 text-sm text-gray-800 shadow-sm transition-colors placeholder:text-gray-400 hover:border-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                  />
                  <button
                    type="button"
                    class="absolute inset-y-0 right-0 flex cursor-pointer items-center px-3 text-gray-400 hover:text-gray-600"
                    @click="showNewPwd = !showNewPwd"
                  >
                    <svg v-if="showNewPwd" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                      <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
                    </svg>
                  </button>
                </div>
                <div v-if="newPassword" class="mt-2 flex items-center gap-2">
                  <div class="flex h-1.5 flex-1 gap-1">
                    <div
                      v-for="i in 4"
                      :key="i"
                      class="flex-1 rounded-full transition-colors"
                      :class="i <= pwdStrength.level ? pwdStrength.color : 'bg-gray-100'"
                    />
                  </div>
                  <span class="text-xs" :class="pwdStrength.level >= 3 ? 'text-emerald-600' : pwdStrength.level >= 2 ? 'text-amber-600' : 'text-red-500'">
                    {{ pwdStrength.label }}
                  </span>
                </div>
              </div>
              <div>
                <label class="mb-1.5 block text-sm font-medium text-gray-700">确认新密码</label>
                <div class="relative">
                  <input
                    v-model="confirmPassword"
                    :type="showConfirmPwd ? 'text' : 'password'"
                    placeholder="请再次输入新密码"
                    autocomplete="new-password"
                    class="w-full rounded-lg border border-gray-300 bg-white px-3.5 py-2 pr-10 text-sm text-gray-800 shadow-sm transition-colors placeholder:text-gray-400 hover:border-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                  />
                  <button
                    type="button"
                    class="absolute inset-y-0 right-0 flex cursor-pointer items-center px-3 text-gray-400 hover:text-gray-600"
                    @click="showConfirmPwd = !showConfirmPwd"
                  >
                    <svg v-if="showConfirmPwd" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                      <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
                    </svg>
                  </button>
                </div>
                <p v-if="confirmPassword && newPassword !== confirmPassword" class="mt-1 text-xs text-red-500">
                  两次输入的密码不一致
                </p>
              </div>
              <div class="flex items-center justify-end gap-2 border-t border-gray-50 pt-4">
                <BaseButton type="secondary" @click="resetPwdForm">重置</BaseButton>
                <BaseButton type="primary" :loading="pwdSaving" @click="savePassword">确认修改</BaseButton>
              </div>
            </div>
          </BaseCard>
        </div>

        <!-- 邮箱绑定 -->
        <BaseCard id="email-binding-card" class="mt-6">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex items-start gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-violet-50">
                <svg class="h-5 w-5 text-violet-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
                </svg>
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <h2 class="text-base font-semibold tracking-tight text-gray-900">邮箱绑定</h2>
                  <BaseBadge v-if="auth.user.email_verified" type="success" dot>已验证</BaseBadge>
                  <BaseBadge v-else type="warning" dot>未验证</BaseBadge>
                </div>
                <p class="mt-1 text-xs leading-relaxed text-gray-500">
                  验证邮箱可用于账号安全提醒、找回密码等重要操作。
                </p>
                <p class="mt-1.5 text-sm text-gray-700">
                  当前邮箱：<span class="font-medium text-gray-900">{{ auth.user.email }}</span>
                </p>
              </div>
            </div>
            <div class="flex flex-wrap gap-2 sm:flex-nowrap sm:justify-end">
              <BaseButton
                v-if="!auth.user.email_verified"
                type="primary"
                @click="openEmailModal('verify')"
              >
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                立即验证
              </BaseButton>
              <BaseButton type="secondary" @click="openEmailModal('change')">
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
                </svg>
                换绑邮箱
              </BaseButton>
            </div>
          </div>
        </BaseCard>
      </template>
    </template>

    <!-- 编辑资料弹窗 -->
    <BaseModal v-model="showEdit" title="编辑个人资料" width="max-w-xl">
      <div class="space-y-4">
        <div class="flex items-center gap-4 rounded-xl border border-gray-100 bg-gray-50/50 p-4">
          <div
            class="flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-xl text-xl font-bold"
            :class="avatar ? 'bg-white' : avatarBgClass(nickname || auth.user?.username || 'U')"
          >
            <img v-if="avatar" :src="avatar" :alt="nickname" class="h-full w-full object-cover" />
            <template v-else>{{ (nickname || auth.user?.username || 'U').charAt(0).toUpperCase() }}</template>
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-semibold text-gray-800">头像预览</p>
            <p class="mt-0.5 truncate text-xs text-gray-500">
              {{ avatar || '未设置，将使用昵称首字母作为默认头像' }}
            </p>
          </div>
        </div>

        <BaseInput v-model="nickname" label="昵称" placeholder="请输入昵称" />

        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">头像链接</label>
          <BaseInput v-model="avatar" placeholder="请输入头像图片 URL（可选）" />
          <p class="mt-1 text-xs text-gray-400">支持 jpg / png 格式，留空则使用首字母头像</p>
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">考试级别</label>
            <BaseSelect v-model="levelId">
              <option :value="0">请选择</option>
              <option v-for="lv in subjectStore.levels" :key="lv.id" :value="lv.id">{{ lv.name }}</option>
            </BaseSelect>
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">所属科目</label>
            <BaseSelect v-model="subjectId" :disabled="!levelId">
              <option :value="0">请选择</option>
              <option v-for="s in subjectStore.subjects" :key="s.id" :value="s.id">{{ s.name }}</option>
            </BaseSelect>
          </div>
        </div>

        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">难度偏好</label>
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              v-for="opt in difficultyOptions"
              :key="opt.value"
              class="cursor-pointer rounded-full border px-4 py-1.5 text-sm font-medium transition-all"
              :class="difficulty === opt.value
                ? 'border-indigo-500 bg-indigo-50 text-indigo-700 shadow-sm'
                : 'border-gray-200 bg-white text-gray-600 hover:border-indigo-300 hover:bg-indigo-50/40'"
              @click="difficulty = difficulty === opt.value ? '' : opt.value"
            >
              {{ opt.label }}
            </button>
          </div>
          <p class="mt-1.5 text-xs text-gray-400">点击已选中的难度可取消选择</p>
        </div>

        <div class="flex justify-end gap-2 border-t border-gray-50 pt-4">
          <BaseButton type="secondary" @click="showEdit = false">取消</BaseButton>
          <BaseButton type="primary" :loading="saving" @click="saveProfile">保存</BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- 邮箱绑定弹窗 -->
    <BaseModal
      v-model="showEmailModal"
      :title="emailPurpose === 'change' ? '换绑邮箱' : '验证邮箱'"
      width="max-w-lg"
    >
      <div class="space-y-5">
        <div class="rounded-xl border border-violet-100 bg-violet-50/40 p-3.5 text-sm leading-relaxed text-gray-600">
          <p v-if="emailPurpose === 'change'">
            <span class="font-medium text-gray-800">换绑说明：</span>请输入新邮箱，验证码将发送至该邮箱。验证通过后，原邮箱将被替换。
          </p>
          <p v-else>
            <span class="font-medium text-gray-800">验证说明：</span>点击「获取验证码」，6 位数字验证码将发送到下方邮箱，5 分钟内有效。
          </p>
        </div>

        <div v-if="emailError" class="flex items-start gap-2 rounded-lg border border-red-100 bg-red-50/70 px-3.5 py-2.5 text-sm text-red-600">
          <svg class="mt-0.5 h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
          </svg>
          <span>{{ emailError }}</span>
        </div>

        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">
            {{ emailPurpose === 'change' ? '新邮箱地址' : '邮箱地址' }}
          </label>
          <div class="flex gap-2">
            <input
              v-model="emailTarget"
              type="email"
              :placeholder="emailPurpose === 'change' ? '请输入新邮箱' : '请输入邮箱'"
              autocomplete="email"
              :readonly="emailPurpose === 'verify'"
              :class="[
                'flex-1 rounded-lg border bg-white px-3.5 py-2 text-sm text-gray-800 shadow-sm transition-colors placeholder:text-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20',
                emailPurpose === 'verify'
                  ? 'cursor-not-allowed border-gray-200 bg-gray-50 text-gray-500'
                  : 'border-gray-300 hover:border-gray-400',
              ]"
            />
            <button
              type="button"
              class="inline-flex shrink-0 cursor-pointer items-center justify-center gap-1.5 rounded-lg px-3.5 py-2 text-sm font-medium transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500/40 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50 disabled:active:scale-100"
              :class="isEmailCooldown
                ? 'border border-gray-200 bg-gray-50 text-gray-400'
                : 'border border-indigo-200 bg-indigo-50 text-indigo-600 hover:border-indigo-300 hover:bg-indigo-100'"
              :disabled="emailSending || isEmailCooldown || !isEmailValid"
              @click="onSendCode"
            >
              <svg v-if="emailSending" class="h-3.5 w-3.5 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              <svg v-else-if="isEmailCooldown" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <svg v-else class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
              </svg>
              <span>{{ isEmailCooldown ? `${emailCooldown}s 后重试` : (emailSending ? '发送中' : '获取验证码') }}</span>
            </button>
          </div>
        </div>

        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700">邮箱验证码</label>
          <div class="flex gap-2 sm:gap-2.5" @paste="onCodePaste">
            <input
              v-for="(digit, idx) in emailCodeDigits"
              :key="idx"
              :ref="el => { emailCodeRefs[idx] = el as HTMLInputElement | null }"
              type="text"
              inputmode="numeric"
              maxlength="1"
              :value="digit"
              class="h-12 w-10 flex-1 cursor-text rounded-lg border-2 border-gray-200 bg-white text-center text-lg font-bold text-gray-800 shadow-sm transition-all placeholder:text-gray-300 hover:border-gray-300 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 sm:h-14 sm:text-xl"
              placeholder="·"
              @input="e => onCodeInput(idx, e)"
              @keydown="e => onCodeKeydown(idx, e)"
            />
          </div>
          <p class="mt-2 text-xs text-gray-400">
            验证码 6 位数字，5 分钟内有效。开发环境可查看后端日志获取验证码。
          </p>
        </div>

        <div class="flex justify-end gap-2 border-t border-gray-50 pt-4">
          <BaseButton type="secondary" @click="closeEmailModal">取消</BaseButton>
          <BaseButton
            type="primary"
            :loading="emailVerifying"
            :disabled="!isEmailValid || !isEmailCodeComplete"
            @click="onConfirmEmail"
          >
            {{ emailPurpose === 'change' ? '确认换绑' : '确认验证' }}
          </BaseButton>
        </div>
      </div>
    </BaseModal>
  </div>
</template>
