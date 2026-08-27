<script setup lang="ts">
import { ref, computed, onBeforeUnmount, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { sendResetPasswordCode, resetPasswordByCode } from '@/api/auth'
import { showToast } from '@/utils/toast'

const router = useRouter()

const step = ref<1 | 2>(1)

const email = ref('')
const codeDigits = ref<string[]>(['', '', '', '', '', ''])
const codeRefs = ref<(HTMLInputElement | null)[]>([])

const newPassword = ref('')
const confirmPassword = ref('')
const showNewPwd = ref(false)
const showConfirmPwd = ref(false)

const sending = ref(false)
const sendingError = ref('')
const submitting = ref(false)
const submitError = ref('')
const codeCooldown = ref(0)
let codeCooldownTimer: ReturnType<typeof setInterval> | null = null

const isEmailValid = computed(() => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim()))
const isCodeComplete = computed(() => codeDigits.value.every(d => d !== ''))
const isPasswordValid = computed(() => newPassword.value.length >= 8 && /\d/.test(newPassword.value))
const isPasswordMatch = computed(() => newPassword.value && newPassword.value === confirmPassword.value)

const passwordRules = computed(() => ({
  length: newPassword.value.length >= 8,
  digit: /\d/.test(newPassword.value),
}))

const passwordStrength = computed(() => {
  const p = newPassword.value
  if (!p) return { level: 0, label: '', color: 'bg-gray-200' }
  let score = 0
  if (p.length >= 8) score++
  if (p.length >= 12) score++
  if (/[A-Z]/.test(p) && /[a-z]/.test(p)) score++
  if (/\d/.test(p)) score++
  if (/[^\w\s]/.test(p)) score++
  const level = Math.min(4, Math.max(0, score - 1))
  if (level <= 1) return { level, label: '弱', color: 'bg-red-400' }
  if (level === 2) return { level, label: '中等', color: 'bg-amber-400' }
  return { level, label: '强', color: 'bg-emerald-500' }
})

const canSubmit = computed(
  () => isEmailValid.value && isCodeComplete.value && isPasswordValid.value && isPasswordMatch.value && !submitting.value
)

const isCooldown = computed(() => codeCooldown.value > 0)

function startCooldown() {
  codeCooldown.value = 60
  if (codeCooldownTimer) clearInterval(codeCooldownTimer)
  codeCooldownTimer = setInterval(() => {
    codeCooldown.value -= 1
    if (codeCooldown.value <= 0 && codeCooldownTimer) {
      clearInterval(codeCooldownTimer)
      codeCooldownTimer = null
    }
  }, 1000)
}

async function onSendCode() {
  if (!isEmailValid.value) {
    sendingError.value = '请输入有效的邮箱地址'
    return
  }
  sending.value = true
  sendingError.value = ''
  try {
    await sendResetPasswordCode({ email: email.value.trim() })
    showToast('如果该邮箱已注册，验证码已发送', 'success')
    startCooldown()
    step.value = 2
    await nextTick()
    codeRefs.value[0]?.focus()
  } catch {
    // toast 由 request 层提示
  } finally {
    sending.value = false
  }
}

function onCodeInput(idx: number, e: Event) {
  const target = e.target as HTMLInputElement
  const v = target.value.replace(/\D/g, '')
  codeDigits.value[idx] = v.slice(-1)
  if (v && idx < 5) {
    codeRefs.value[idx + 1]?.focus()
  }
}

function onCodeKeydown(idx: number, e: KeyboardEvent) {
  if (e.key === 'Backspace' && !codeDigits.value[idx] && idx > 0) {
    codeRefs.value[idx - 1]?.focus()
  }
}

function onCodePaste(e: ClipboardEvent) {
  const text = e.clipboardData?.getData('text') || ''
  const digits = text.replace(/\D/g, '').slice(0, 6).split('')
  if (!digits.length) return
  e.preventDefault()
  for (let i = 0; i < 6; i++) {
    codeDigits.value[i] = digits[i] || ''
  }
  const nextEmpty = codeDigits.value.findIndex(d => !d)
  const focusIdx = nextEmpty === -1 ? 5 : nextEmpty
  codeRefs.value[focusIdx]?.focus()
}

async function onSubmit() {
  submitError.value = ''
  if (!isEmailValid.value) {
    submitError.value = '请输入有效的邮箱地址'
    return
  }
  if (!isCodeComplete.value) {
    submitError.value = '请输入完整的 6 位验证码'
    return
  }
  if (!isPasswordValid.value) {
    submitError.value = '密码至少 8 位且包含数字'
    return
  }
  if (!isPasswordMatch.value) {
    submitError.value = '两次输入的密码不一致'
    return
  }
  submitting.value = true
  try {
    await resetPasswordByCode({
      email: email.value.trim(),
      code: codeDigits.value.join(''),
      new_password: newPassword.value,
      confirm_password: confirmPassword.value,
    })
    showToast('密码重置成功，请使用新密码登录', 'success')
    router.push('/login')
  } catch {
    // toast 由 request 层提示
  } finally {
    submitting.value = false
  }
}

function goBack() {
  step.value = 1
}

onBeforeUnmount(() => {
  if (codeCooldownTimer) {
    clearInterval(codeCooldownTimer)
    codeCooldownTimer = null
  }
})
</script>

<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden bg-gray-50 px-4 py-12">
    <div class="pointer-events-none absolute inset-0">
      <div class="bg-brand-gradient absolute -top-32 left-1/2 h-80 w-[42rem] -translate-x-1/2 rounded-full opacity-15 blur-3xl" />
      <div class="absolute -bottom-24 -left-24 h-72 w-72 rounded-full bg-violet-300 opacity-20 blur-3xl" />
      <div class="absolute -right-24 top-1/3 h-72 w-72 rounded-full bg-indigo-300 opacity-20 blur-3xl" />
    </div>

    <div class="relative w-full max-w-md">
      <div class="mb-8 flex flex-col items-center">
        <div class="bg-brand-gradient mb-4 flex h-14 w-14 items-center justify-center rounded-2xl text-white shadow-lg shadow-indigo-600/30">
          <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold tracking-tight text-gray-900">找回密码</h1>
        <p class="mt-1.5 text-sm text-gray-500">通过注册邮箱验证身份，重置登录密码</p>
      </div>

      <div class="rounded-2xl border border-gray-100 bg-white/90 p-8 shadow-xl shadow-gray-900/5 backdrop-blur">
        <div class="mb-6 flex items-center gap-2">
          <div
            class="flex h-7 w-7 items-center justify-center rounded-full text-xs font-semibold transition-colors"
            :class="step === 1 ? 'bg-brand-gradient text-white shadow-sm shadow-indigo-600/25' : 'bg-emerald-500 text-white'"
          >
            <svg v-if="step > 1" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
            <span v-else>1</span>
          </div>
          <div class="h-px flex-1 bg-gray-200"></div>
          <div
            class="flex h-7 w-7 items-center justify-center rounded-full text-xs font-semibold transition-colors"
            :class="step === 2 ? 'bg-brand-gradient text-white shadow-sm shadow-indigo-600/25' : 'bg-gray-100 text-gray-400'"
          >
            2
          </div>
        </div>

        <div v-if="step === 1" class="space-y-5">
          <div class="rounded-xl border border-indigo-100 bg-indigo-50/40 p-3.5 text-sm leading-relaxed text-gray-600">
            <p>
              <span class="font-medium text-gray-800">说明：</span>请输入注册时使用的邮箱，我们会向该邮箱发送 6 位数字验证码，5 分钟内有效。
            </p>
          </div>

          <div v-if="sendingError" class="flex items-start gap-2 rounded-lg border border-red-100 bg-red-50/70 px-3.5 py-2.5 text-sm text-red-600">
            <svg class="mt-0.5 h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
            </svg>
            <span>{{ sendingError }}</span>
          </div>

          <div>
            <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">注册邮箱</label>
            <div class="flex gap-2">
              <input
                v-model="email"
                type="email"
                maxlength="100"
                placeholder="请输入注册邮箱"
                autocomplete="email"
                class="block w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-indigo-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-100"
              />
              <button
                type="button"
                class="inline-flex shrink-0 cursor-pointer items-center justify-center gap-1.5 rounded-xl px-3.5 py-2.5 text-sm font-medium transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500/40 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50 disabled:active:scale-100"
                :class="isCooldown
                  ? 'border border-gray-200 bg-gray-50 text-gray-400'
                  : 'border border-indigo-200 bg-indigo-50 text-indigo-600 hover:border-indigo-300 hover:bg-indigo-100'"
                :disabled="sending || isCooldown || !isEmailValid"
                @click="onSendCode"
              >
                <svg v-if="sending" class="h-3.5 w-3.5 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
                <svg v-else-if="isCooldown" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <svg v-else class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
                </svg>
                <span>{{ isCooldown ? `${codeCooldown}s` : (sending ? '发送中' : '获取验证码') }}</span>
              </button>
            </div>
          </div>

          <div class="text-center">
            <p class="text-sm text-gray-400">
              想起密码了？
              <router-link to="/login" class="font-semibold text-indigo-600 transition-colors hover:text-indigo-700">返回登录</router-link>
            </p>
          </div>
        </div>

        <div v-else class="space-y-5">
          <div class="rounded-xl border border-violet-100 bg-violet-50/40 p-3.5 text-sm leading-relaxed text-gray-600">
            <p>
              <span class="font-medium text-gray-800">验证码已发送至</span>
              <span class="font-medium text-gray-900">{{ email }}</span>
              <button
                type="button"
                class="ml-1 cursor-pointer text-xs font-medium text-indigo-600 transition-colors hover:text-indigo-700"
                @click="goBack"
              >
                修改
              </button>
            </p>
          </div>

          <div v-if="submitError" class="flex items-start gap-2 rounded-lg border border-red-100 bg-red-50/70 px-3.5 py-2.5 text-sm text-red-600">
            <svg class="mt-0.5 h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
            </svg>
            <span>{{ submitError }}</span>
          </div>

          <div>
            <label class="mb-2 block text-xs font-medium uppercase tracking-wider text-gray-500">邮箱验证码</label>
            <div class="flex gap-2 sm:gap-2.5" @paste="onCodePaste">
              <input
                v-for="(digit, idx) in codeDigits"
                :key="idx"
                :ref="el => { codeRefs[idx] = el as HTMLInputElement | null }"
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
            <div class="mt-2 flex items-center justify-between text-xs">
              <span class="text-gray-400">验证码 6 位数字，5 分钟内有效</span>
              <button
                v-if="!isCooldown"
                type="button"
                class="cursor-pointer font-medium text-indigo-600 transition-colors hover:text-indigo-700 disabled:cursor-not-allowed disabled:text-gray-400"
                :disabled="!isEmailValid || sending"
                @click="onSendCode"
              >
                重新发送
              </button>
              <span v-else class="text-gray-400">{{ codeCooldown }}s 后可重发</span>
            </div>
          </div>

          <div>
            <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">新密码</label>
            <div class="relative">
              <input
                v-model="newPassword"
                :type="showNewPwd ? 'text' : 'password'"
                maxlength="128"
                placeholder="至少 8 位，包含数字"
                autocomplete="new-password"
                class="block w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 pr-11 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-indigo-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-100"
              />
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 cursor-pointer text-xs font-medium text-gray-400 transition-colors hover:text-gray-600"
                @click="showNewPwd = !showNewPwd"
              >
                {{ showNewPwd ? '隐藏' : '显示' }}
              </button>
            </div>
            <div v-if="newPassword" class="mt-2 space-y-1.5">
              <div class="flex items-center gap-2">
                <div class="flex h-1.5 flex-1 gap-1">
                  <div
                    v-for="i in 4"
                    :key="i"
                    class="flex-1 rounded-full transition-colors"
                    :class="i <= passwordStrength.level ? passwordStrength.color : 'bg-gray-100'"
                  />
                </div>
                <span class="text-xs" :class="passwordStrength.level >= 3 ? 'text-emerald-600' : passwordStrength.level >= 2 ? 'text-amber-600' : 'text-red-500'">
                  {{ passwordStrength.label }}
                </span>
              </div>
              <ul class="space-y-0.5 text-xs">
                <li class="flex items-center gap-1.5" :class="passwordRules.length ? 'text-emerald-600' : 'text-gray-400'">
                  <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path v-if="passwordRules.length" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                    <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
                  </svg>
                  至少 8 个字符
                </li>
                <li class="flex items-center gap-1.5" :class="passwordRules.digit ? 'text-emerald-600' : 'text-gray-400'">
                  <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path v-if="passwordRules.digit" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                    <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
                  </svg>
                  至少包含一个数字
                </li>
              </ul>
            </div>
          </div>

          <div>
            <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">确认新密码</label>
            <div class="relative">
              <input
                v-model="confirmPassword"
                :type="showConfirmPwd ? 'text' : 'password'"
                maxlength="128"
                placeholder="再次输入新密码"
                autocomplete="new-password"
                class="block w-full rounded-xl border bg-gray-50 px-4 py-2.5 pr-11 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:bg-white focus:outline-none focus:ring-2"
                :class="confirmPassword && !isPasswordMatch
                  ? 'border-red-200 focus:border-red-400 focus:ring-red-100'
                  : 'border-gray-200 focus:border-indigo-400 focus:ring-indigo-100'"
              />
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 cursor-pointer text-xs font-medium text-gray-400 transition-colors hover:text-gray-600"
                @click="showConfirmPwd = !showConfirmPwd"
              >
                {{ showConfirmPwd ? '隐藏' : '显示' }}
              </button>
            </div>
            <p v-if="confirmPassword && !isPasswordMatch" class="mt-1 text-xs text-red-500">
              两次输入的密码不一致
            </p>
          </div>

          <button
            type="button"
            :disabled="!canSubmit"
            class="w-full rounded-xl px-4 py-2.5 text-sm font-semibold text-white transition-all duration-200"
            :class="canSubmit ? 'bg-brand-gradient shadow-md shadow-indigo-600/25 hover:shadow-lg hover:brightness-110 active:scale-[0.99]' : 'cursor-not-allowed bg-indigo-300'"
            @click="onSubmit"
          >
            <span v-if="submitting" class="inline-flex items-center justify-center gap-2">
              <span class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
              重置中...
            </span>
            <span v-else>确认重置</span>
          </button>

          <div class="text-center">
            <p class="text-sm text-gray-400">
              <button
                type="button"
                class="cursor-pointer font-medium text-gray-500 transition-colors hover:text-indigo-600"
                @click="goBack"
              >
                ← 上一步
              </button>
            </p>
          </div>
        </div>
      </div>

      <p class="mt-6 text-center text-xs text-gray-400">软考学系 · 你的智能备考助手</p>
    </div>
  </div>
</template>
