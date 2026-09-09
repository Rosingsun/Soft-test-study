<script setup lang="ts">
import { ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    label?: string
    error?: string
    placeholder?: string
    type?: string
    disabled?: boolean
    /** 当 type="password" 时，显示明文切换（眼睛）开关 */
    toggleable?: boolean
  }>(),
  {
    type: 'text',
    disabled: false,
    toggleable: false,
  },
)

const model = defineModel<string | number | null>({ default: '' })

// 明文切换：showPlain 控制当前是否为明文展示
const showPlain = ref(false)
watch(
  () => props.type,
  () => {
    // 切换 type 时复位为掩码状态，避免遗留明文
    if (!props.toggleable || props.type !== 'password') showPlain.value = false
  },
)
const isPassword = () => props.type === 'password'
const currentType = () => (isPassword() && showPlain.value ? 'text' : props.type)
</script>

<template>
  <div>
    <label v-if="label" class="mb-1.5 block text-sm font-medium text-gray-700">{{ label }}</label>
    <div class="relative">
      <input
        v-model="model"
        :type="currentType()"
        :placeholder="placeholder"
        :disabled="disabled"
        class="w-full rounded-lg border border-gray-300 bg-white px-3.5 py-2 text-sm text-gray-800 shadow-sm transition-colors placeholder:text-gray-400 hover:border-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 disabled:bg-gray-100 disabled:text-gray-400"
        :class="isPassword() && toggleable ? 'pr-10' : ''"
      />
      <button
        v-if="isPassword() && toggleable"
        type="button"
        :disabled="disabled"
        :aria-label="showPlain ? '隐藏密码' : '显示密码'"
        class="absolute right-2.5 top-1/2 flex -translate-y-1/2 cursor-pointer items-center justify-center rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 disabled:cursor-not-allowed disabled:opacity-50"
        @click="showPlain = !showPlain"
      >
        <svg v-if="!showPlain" class="h-4.5 w-4.5" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" />
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
        </svg>
        <svg v-else class="h-4.5 w-4.5" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88" />
        </svg>
      </button>
    </div>
    <p v-if="error" class="mt-1 text-xs text-red-500">{{ error }}</p>
  </div>
</template>
