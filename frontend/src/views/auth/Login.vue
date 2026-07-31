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
  <div class="mx-auto mt-16 max-w-sm px-4">
    <div class="rounded-xl border border-gray-200 bg-white p-8">
      <div class="mb-8 text-center">
        <h1 class="text-2xl font-bold tracking-tight text-gray-900">欢迎回来</h1>
        <p class="mt-2 text-sm text-gray-500">登录你的账号继续学习</p>
      </div>

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
          class="w-full rounded-xl px-4 py-2.5 text-sm font-medium text-white transition-colors"
          :class="canSubmit && !loading ? 'bg-indigo-600 hover:bg-indigo-700' : 'cursor-not-allowed bg-indigo-300'"
        >
          <span v-if="loading" class="inline-flex items-center justify-center gap-2">
            <span class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
            登录中...
          </span>
          <span v-else>登录</span>
        </button>
      </form>

      <div class="mt-6 text-center">
        <p class="text-sm text-gray-400">
          还没有账号？
          <router-link to="/register" class="font-medium text-indigo-600 transition-colors hover:text-indigo-700">立即注册</router-link>
        </p>
      </div>
    </div>
  </div>
</template>
