<script setup lang="ts">
withDefaults(defineProps<{
  type?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'success'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  disabled?: boolean
  block?: boolean
}>(), {
  type: 'primary',
  size: 'md',
  loading: false,
  disabled: false,
  block: false,
})
</script>

<template>
  <button
    :disabled="disabled || loading"
    :class="[
      'inline-flex cursor-pointer items-center justify-center gap-1.5 rounded-lg font-medium transition-all focus:outline-none focus:ring-2 focus:ring-indigo-500/40 disabled:cursor-not-allowed disabled:opacity-50',
      block ? 'w-full' : '',
      size === 'sm' ? 'px-2.5 py-1.5 text-xs' : size === 'lg' ? 'px-5 py-2.5 text-base' : 'px-4 py-2 text-sm',
      type === 'primary' && 'bg-indigo-600 text-white hover:bg-indigo-700 shadow-sm',
      type === 'secondary' && 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50',
      type === 'danger' && 'bg-red-500 text-white hover:bg-red-600 shadow-sm',
      type === 'success' && 'bg-emerald-500 text-white hover:bg-emerald-600 shadow-sm',
      type === 'ghost' && 'text-gray-600 hover:bg-gray-100',
    ]"
  >
    <svg v-if="loading" class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
    <slot />
  </button>
</template>
