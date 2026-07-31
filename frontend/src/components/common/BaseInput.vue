<script setup lang="ts">
withDefaults(defineProps<{
  label?: string
  error?: string
  placeholder?: string
  type?: string
  disabled?: boolean
}>(), {
  type: 'text',
  disabled: false,
})

const model = defineModel<string | number | null>({ default: '' })
</script>

<template>
  <div>
    <label v-if="label" class="mb-1.5 block text-sm font-medium text-gray-700">{{ label }}</label>
    <input
      v-model="model"
      :type="type"
      :placeholder="placeholder"
      :disabled="disabled"
      class="w-full rounded-lg border border-gray-300 bg-white px-3.5 py-2 text-sm text-gray-800 transition-colors placeholder:text-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 disabled:bg-gray-100 disabled:text-gray-400"
    />
    <p v-if="error" class="mt-1 text-xs text-red-500">{{ error }}</p>
  </div>
</template>
