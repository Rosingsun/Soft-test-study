<script setup lang="ts">
withDefaults(defineProps<{
  label?: string
  placeholder?: string
  rows?: number
}>(), {
  rows: 4,
})

const model = defineModel<string | null>({ default: '' })
</script>

<template>
  <div>
    <label v-if="label" class="mb-1.5 block text-sm font-medium text-gray-700">{{ label }}</label>
    <textarea
      v-model="model"
      :rows="rows"
      :placeholder="placeholder"
      class="w-full rounded-lg border border-gray-300 bg-white px-3.5 py-2 text-sm text-gray-800 transition-colors placeholder:text-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
    />
  </div>
</template>
