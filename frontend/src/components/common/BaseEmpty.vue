<script setup lang="ts">
withDefaults(defineProps<{
  title?: string
  description?: string
}>(), {
  title: '暂无数据',
})

defineEmits<{ (e: 'action'): void }>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-14 text-center">
    <div class="mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-gray-100">
      <svg class="h-7 w-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
      </svg>
    </div>
    <p class="text-sm font-medium text-gray-600">{{ title }}</p>
    <p v-if="description" class="mt-1 text-xs text-gray-400">{{ description }}</p>
    <slot name="action" />
  </div>
</template>
