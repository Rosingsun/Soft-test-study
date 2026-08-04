<script setup lang="ts">
withDefaults(defineProps<{
  text?: string
}>(), {
  text: '',
})
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-3 py-20">
    <div class="relative h-10 w-10">
      <div class="absolute inset-0 rounded-full border-4 border-indigo-100" />
      <div class="absolute inset-0 animate-spin rounded-full border-4 border-transparent border-t-indigo-600" />
    </div>
    <p v-if="text" class="text-sm text-gray-400">{{ text }}</p>
  </div>
</template>
