<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const showPassword = ref(false)

const canSubmit = computed(() => username.value.trim().length > 0 && password.value.length >= 1)

async function handleLogin() {
  if (!canSubmit.value || loading.value) return
  try {
    loading.value = true
    error.value = ''
    await auth.login(username.value, password.value)
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

    <div class="relative w-full max-w-sm">
      <!-- 品牌标识 -->
      <div class="mb-8 flex flex-col items-center">
        <div class="bg-brand-gradient mb-4 flex h-14 w-14 items-center justify-center rounded-2xl text-white shadow-lg shadow-indigo-600/30">
          <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold tracking-tight text-gray-900">软考学系</h1>
        <p class="mt-1.5 text-sm text-gray-500">登录你的账号，继续备考之旅</p>
      </div>

      <div class="rounded-2xl border border-gray-100 bg-white/90 p-8 shadow-xl shadow-gray-900/5 backdrop-blur">

      <form @submit.prevent="handleLogin" class="space-y-5">
        <div>
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">用户名 / 邮箱</label>
          <input
            v-model="username"
            type="text"
            maxlength="50"
            required
            placeholder="请输入用户名或邮箱"
            class="block w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-indigo-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-100"
          />
        </div>

        <div>
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500">密码</label>
          <div class="relative">
            <input
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              maxlength="128"
              required
              placeholder="请输入密码"
              class="block w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 pr-11 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-indigo-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-100"
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 cursor-pointer text-xs font-medium text-gray-400 transition-colors hover:text-gray-600"
              @click="showPassword = !showPassword"
            >
              {{ showPassword ? '隐藏' : '显示' }}
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
            登录中...
          </span>
          <span v-else>登 录</span>
        </button>
      </form>

      <div class="mt-6 text-center">
        <p class="text-sm text-gray-400">
          还没有账号？
          <router-link to="/register" class="font-semibold text-indigo-600 transition-colors hover:text-indigo-700">立即注册</router-link>
        </p>
      </div>
      </div>

      <p class="mt-6 text-center text-xs text-gray-400">软考学系 · 你的智能备考助手</p>
    </div>
  </div>
</template>
