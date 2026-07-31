<script setup lang="ts">
withDefaults(defineProps<{
  type?: 'success' | 'warning' | 'danger' | 'info' | 'default'
}>(), {
  type: 'default',
})
</script>

<template>
  <span
    :class="[
      'inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium',
      type === 'success' && 'bg-emerald-50 text-emerald-600',
      type === 'warning' && 'bg-yellow-50 text-yellow-600',
      type === 'danger' && 'bg-red-50 text-red-600',
      type === 'info' && 'bg-indigo-50 text-indigo-600',
      type === 'default' && 'bg-gray-100 text-gray-600',
    ]"
  >
    <slot />
  </span>
</template>
