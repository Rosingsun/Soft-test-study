<script setup lang="ts">
import { ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  modelValue?: string | number
  tabs: { label: string; value: string | number }[]
}>(), {
  modelValue: '',
})

const emit = defineEmits<{ (e: 'update:modelValue', v: string | number): void }>()
const active = ref<string | number>(props.modelValue)

watch(() => props.modelValue, v => { if (v !== undefined && v !== '') active.value = v })
watch(active, v => emit('update:modelValue', v))
</script>

<template>
  <div class="flex gap-1 rounded-lg bg-gray-100 p-1">
    <button
      v-for="tab in tabs"
      :key="tab.value"
      class="flex-1 cursor-pointer rounded-md px-3 py-1.5 text-sm font-medium transition-all"
      :class="active === tab.value ? 'bg-white text-indigo-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'"
      @click="active = tab.value"
    >
      {{ tab.label }}
    </button>
  </div>
</template>
