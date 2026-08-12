<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubjectStore } from '@/stores/subject'
import { showToast } from '@/utils/toast'

const auth = useAuthStore()
const subjectStore = useSubjectStore()
const router = useRouter()
const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const email = ref('')
const inviteCode = ref('')
const selectedLevelId = ref<number | null>(null)
const selectedSubjectId = ref<number | null>(null)
const selectedDifficulty = ref('')
const error = ref('')
const loading = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const step = ref(1)

const isAdminBypass = computed(() => username.value.trim().toLowerCase() === 'ross')

const usernameError = computed(() => {
  if (!username.value) return ''
  if (username.value.length < 3) return '用户名至少3个字符'
  if (!/^[a-zA-Z0-9_\u4e00-\u9fa5]+$/.test(username.value)) return '只能包含字母、数字、下划线和中文'
  return ''
})

const passwordRules = computed(() => {
  const pw = password.value
  return {
    length: pw.length >= 8,
    digit: /[0-9]/.test(pw),
    charCount: pw.length,
  }
})

const passwordStrength = computed(() => {
  const { length, digit, charCount } = passwordRules.value
  let score = 0
  if (length) score++
  if (digit) score++
  if (charCount >= 12) score++
  if (!password.value) return { level: 0, label: '', color: '', width: 0 }
  if (score <= 1) return { level: 1, label: '弱', color: 'bg-red-400', width: 25 }
  if (score === 2) return { level: 2, label: '中', color: 'bg-amber-400', width: 60 }
  return { level: 3, label: '强', color: 'bg-emerald-400', width: 100 }
})

const passwordMatch = computed(() => {
  if (!confirmPassword.value) return ''
  return password.value === confirmPassword.value ? '' : '两次密码不一致'
})

const emailError = computed(() => {
  if (!email.value) return ''
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value) ? '' : '邮箱格式不正确'
})

const inviteCodeError = computed(() => {
  if (isAdminBypass.value) return ''
  if (!inviteCode.value) return ''
  if (inviteCode.value.length < 6 || inviteCode.value.length > 32) return '邀请码格式不正确'
  return ''
})

const canProceedStep2 = computed(() => selectedLevelId.value !== null && selectedSubjectId.value !== null)

const selectedLevelName = computed(() =>
  subjectStore.levels.find(l => l.id === selectedLevelId.value)?.name || ''
)
const selectedSubjectName = computed(() =>
  subjectStore.subjects.find(s => s.id === selectedSubjectId.value)?.name || ''
)

const canSubmit = computed(() => {
  const { length, digit } = passwordRules.value
  const inviteOk = isAdminBypass.value
    ? true
    : inviteCode.value.length >= 6 && inviteCode.value.length <= 32
  return (
    username.value.length >= 3 &&
    !usernameError.value &&
    length && digit &&
    !passwordMatch.value &&
    !emailError.value &&
    email.value &&
    inviteOk
  )
})

onMounted(async () => {
  if (subjectStore.levels.length === 0) {
    await subjectStore.fetchLevels()
  }
})

function selectLevel(levelId: number) {
  selectedLevelId.value = levelId
  selectedSubjectId.value = null
  subjectStore.fetchSubjectsByLevel(levelId)
}

function selectSubject(subjectId: number) {
  selectedSubjectId.value = subjectId
  step.value = 2
}

async function handleRegister() {
  if (!canSubmit.value || loading.value) return
  try {
    loading.value = true
    error.value = ''
    await auth.register(
      username.value,
      password.value,
      confirmPassword.value,
      email.value,
      selectedLevelId.value!,
      selectedSubjectId.value!,
      selectedDifficulty.value || undefined,
      isAdminBypass.value ? undefined : inviteCode.value.trim(),
    )
    showToast('注册成功，已自动登录')
    router.push(auth.getHomeRoute())
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden bg-gray-50 px-4 py-12">
    <!-- 背景装饰 -->
    <div class="pointer-events-none absolute inset-0">
      <div class="bg-brand-gradient absolute -top-32 left-1/2 h-80 w-[42rem] -translate-x-1/2 rounded-full opacity-15 blur-3xl" />
      <div class="absolute -bottom-24 -left-24 h-72 w-72 rounded-full bg-violet-300 opacity-20 blur-3xl" />
      <div class="absolute -right-24 top-1/3 h-72 w-72 rounded-full bg-indigo-300 opacity-20 blur-3xl" />
    </div>

    <div class="relative w-full max-w-md">
      <div class="mb-6 flex flex-col items-center">
        <div class="bg-brand-gradient mb-3 flex h-12 w-12 items-center justify-center rounded-2xl text-white shadow-lg shadow-indigo-600/30">
          <svg class="h-7 w-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold tracking-tight text-gray-900">创建账号</h1>
        <p class="mt-1.5 text-sm text-gray-500">开启你的软考备考之旅</p>
      </div>

      <div class="rounded-2xl border border-gray-100 bg-white/90 p-8 shadow-xl shadow-gray-900/5 backdrop-blur">

      <div v-if="step === 1" class="space-y-5">
        <p class="text-sm font-medium text-gray-700">选择你的考试目标</p>

        <div v-if="subjectStore.levels.length" class="flex gap-2 flex-wrap">
          <button
            v-for="level in subjectStore.levels"
            :key="level.id"
            type="button"
            class="cursor-pointer rounded-xl px-4 py-2 text-sm font-medium border transition-colors"
            :class="selectedLevelId === level.id
              ? 'bg-indigo-600 text-white border-indigo-600'
              : 'bg-white text-gray-600 border-gray-200 hover:border-gray-300 hover:text-gray-800'"
            @click="selectLevel(level.id)"
          >
            {{ level.name }}
          </button>
        </div>

        <div v-if="subjectStore.subjects.length" class="grid grid-cols-2 gap-2">
          <button
            v-for="sub in subjectStore.subjects"
            :key="sub.id"
            type="button"
            class="cursor-pointer rounded-xl px-3 py-2.5 text-sm border transition-colors text-left"
            :class="selectedSubjectId === sub.id
              ? 'bg-indigo-50 text-indigo-700 border-indigo-300'
              : 'bg-white text-gray-600 border-gray-200 hover:border-gray-300 hover:text-gray-800'"
            @click="selectSubject(sub.id)"
          >
            <div class="font-medium">{{ sub.short_name || sub.name }}</div>
          </button>
        </div>

        <div class="flex justify-end pt-2">
          <button
            type="button"
            :disabled="!canProceedStep2"
            class="rounded-xl px-5 py-2.5 text-sm font-semibold text-white transition-all duration-200"
            :class="canProceedStep2 ? 'bg-brand-gradient shadow-md shadow-indigo-600/25 hover:shadow-lg hover:brightness-110' : 'cursor-not-allowed bg-indigo-300'"
            @click="step = 2"
          >
            下一步
          </button>
        </div>
      </div>

      <form v-else @submit.prevent="handleRegister" class="space-y-5">
        <div class="mb-1 flex items-center justify-between">
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1 text-sm text-gray-400 transition-colors hover:text-indigo-600"
            @click="step = 1"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
            返回修改
          </button>
          <span class="text-sm text-indigo-600">{{ selectedLevelName }} · {{ selectedSubjectName }}</span>
        </div>

        <div>
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">用户名</label>
          <input
            v-model="username"
            type="text"
            maxlength="50"
            required
            placeholder="3-50个字符，支持字母、数字、下划线"
            class="block w-full rounded-xl border bg-gray-50 px-4 py-2.5 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:bg-white focus:outline-none focus:ring-2"
            :class="usernameError ? 'border-red-200 focus:border-red-400 focus:ring-red-100' : 'border-gray-200 focus:border-indigo-400 focus:ring-indigo-100'"
          />
          <p v-if="usernameError" class="mt-1.5 text-xs text-red-500">{{ usernameError }}</p>
        </div>

        <div>
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">邮箱</label>
          <input
            v-model="email"
            type="email"
            required
            placeholder="请输入邮箱地址"
            class="block w-full rounded-xl border bg-gray-50 px-4 py-2.5 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:bg-white focus:outline-none focus:ring-2"
            :class="emailError ? 'border-red-200 focus:border-red-400 focus:ring-red-100' : 'border-gray-200 focus:border-indigo-400 focus:ring-indigo-100'"
          />
          <p v-if="emailError" class="mt-1.5 text-xs text-red-500">{{ emailError }}</p>
        </div>

        <div>
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">密码</label>
          <div class="relative">
            <input
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              maxlength="128"
              required
              placeholder="至少8位，包含数字"
              class="block w-full rounded-xl border bg-gray-50 px-4 py-2.5 pr-11 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:bg-white focus:outline-none focus:ring-2"
              :class="password && passwordStrength.level === 1 ? 'border-red-200 focus:border-red-400 focus:ring-red-100' : 'border-gray-200 focus:border-indigo-400 focus:ring-indigo-100'"
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 cursor-pointer text-xs font-medium text-gray-400 transition-colors hover:text-gray-600"
              @click="showPassword = !showPassword"
            >
              {{ showPassword ? '隐藏' : '显示' }}
            </button>
          </div>

          <div v-if="password" class="mt-2">
            <div class="h-1 w-full overflow-hidden rounded-full bg-gray-100">
              <div
                class="h-full rounded-full transition-all duration-300"
                :class="passwordStrength.color"
                :style="{ width: passwordStrength.width + '%' }"
              ></div>
            </div>
            <ul class="mt-2 space-y-1">
              <li class="flex items-center gap-1.5 text-xs" :class="passwordRules.length ? 'text-emerald-600' : 'text-gray-400'">
                <svg class="h-3.5 w-3.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path v-if="passwordRules.length" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                  <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
                </svg>
                至少 8 个字符
              </li>
              <li class="flex items-center gap-1.5 text-xs" :class="passwordRules.digit ? 'text-emerald-600' : 'text-gray-400'">
                <svg class="h-3.5 w-3.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path v-if="passwordRules.digit" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                  <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
                </svg>
                至少包含一个数字 (0-9)
              </li>
            </ul>
          </div>
        </div>

        <div>
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">确认密码</label>
          <div class="relative">
            <input
              v-model="confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              maxlength="128"
              required
              placeholder="再次输入密码"
              class="block w-full rounded-xl border bg-gray-50 px-4 py-2.5 pr-11 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:bg-white focus:outline-none focus:ring-2"
              :class="confirmPassword && passwordMatch ? 'border-red-200 focus:border-red-400 focus:ring-red-100' : 'border-gray-200 focus:border-indigo-400 focus:ring-indigo-100'"
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 cursor-pointer text-xs font-medium text-gray-400 transition-colors hover:text-gray-600"
              @click="showConfirmPassword = !showConfirmPassword"
            >
              {{ showConfirmPassword ? '隐藏' : '显示' }}
            </button>
          </div>
          <p v-if="confirmPassword && passwordMatch" class="mt-1.5 text-xs text-red-500">{{ passwordMatch }}</p>
        </div>

        <div v-if="!isAdminBypass">
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">
            邀请码
            <span class="ml-1 normal-case text-gray-400 font-normal">由管理员发放，必填</span>
          </label>
          <input
            v-model="inviteCode"
            type="text"
            maxlength="32"
            required
            placeholder="请输入 12 位邀请码"
            autocomplete="off"
            class="block w-full rounded-xl border bg-gray-50 px-4 py-2.5 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:bg-white focus:outline-none focus:ring-2 uppercase tracking-wider"
            :class="inviteCodeError ? 'border-red-200 focus:border-red-400 focus:ring-red-100' : 'border-gray-200 focus:border-indigo-400 focus:ring-indigo-100'"
          />
          <p v-if="inviteCodeError" class="mt-1.5 text-xs text-red-500">{{ inviteCodeError }}</p>
        </div>

        <div class="border-t border-gray-100 pt-4">
          <p class="mb-3 text-sm font-medium text-gray-700">答题偏好 <span class="text-gray-400 font-normal">（可跳过，后续可在页面右上角修改）</span></p>

          <p class="mb-2 text-xs text-gray-500">答题强度</p>
          <div class="flex gap-2">
            <button
              v-for="item in [{ value: 'easy', label: '简单' }, { value: 'medium', label: '中等' }, { value: 'hard', label: '困难' }]"
              :key="item.value"
              type="button"
              class="cursor-pointer rounded-xl px-3 py-1.5 text-xs font-medium border transition-colors"
              :class="selectedDifficulty === item.value
                ? 'bg-indigo-600 text-white border-indigo-600'
                : 'bg-white text-gray-600 border-gray-200 hover:border-gray-300'"
              @click="selectedDifficulty = item.value"
            >
              {{ item.label }}
            </button>
          </div>
        </div>

        <div v-if="error" class="rounded-xl border border-red-200 bg-red-50 px-4 py-3">
          <p class="text-sm text-red-600">{{ error }}</p>
        </div>

        <button
          type="submit"
          :disabled="!canSubmit || loading"
          class="w-full rounded-xl px-4 py-2.5 text-sm font-semibold text-white transition-all duration-200"
          :class="canSubmit && !loading ? 'bg-brand-gradient shadow-md shadow-indigo-600/25 hover:shadow-lg hover:brightness-110 active:scale-[0.99]' : 'cursor-not-allowed bg-indigo-300'"
        >
          <span v-if="loading" class="inline-flex items-center justify-center gap-2">
            <span class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
            注册中...
          </span>
          <span v-else>注 册</span>
        </button>
      </form>

      <div class="mt-6 text-center">
        <p class="text-sm text-gray-400">
          已有账号？
          <router-link to="/login" class="font-semibold text-indigo-600 transition-colors hover:text-indigo-700">立即登录</router-link>
        </p>
      </div>
      </div>

      <p class="mt-6 text-center text-xs text-gray-400">软考学系 · 你的智能备考助手</p>
    </div>
  </div>
</template>
