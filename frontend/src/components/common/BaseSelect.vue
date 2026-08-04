<script setup lang="ts">
withDefaults(defineProps<{
  label?: string
  placeholder?: string
  disabled?: boolean
}>(), {
  disabled: false,
})

const model = defineModel<string | number | null>({ default: '' })
</script>

<template>
  <div>
    <label v-if="label" class="mb-1.5 block text-sm font-medium text-gray-700">{{ label }}</label>
    <select
      v-model="model"
      :disabled="disabled"
      class="w-full cursor-pointer rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-800 shadow-sm transition-colors hover:border-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 disabled:bg-gray-100"
    >
      <slot />
    </select>
  </div>
</template>
