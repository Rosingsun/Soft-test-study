<script setup lang="ts">
withDefaults(defineProps<{
  type?: 'success' | 'warning' | 'danger' | 'info' | 'default'
  dot?: boolean
}>(), {
  type: 'default',
  dot: false,
})

const dotColorMap: Record<string, string> = {
  success: 'bg-emerald-500',
  warning: 'bg-amber-500',
  danger: 'bg-red-500',
  info: 'bg-indigo-500',
  default: 'bg-gray-400',
}
</script>

<template>
  <span
    :class="[
      'inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset',
      type === 'success' && 'bg-emerald-50 text-emerald-700 ring-emerald-600/20',
      type === 'warning' && 'bg-amber-50 text-amber-700 ring-amber-600/20',
      type === 'danger' && 'bg-red-50 text-red-700 ring-red-600/20',
      type === 'info' && 'bg-indigo-50 text-indigo-700 ring-indigo-600/20',
      type === 'default' && 'bg-gray-100 text-gray-600 ring-gray-500/10',
    ]"
  >
    <span v-if="dot" class="h-1.5 w-1.5 rounded-full" :class="dotColorMap[type]" />
    <slot />
  </span>
</template>
