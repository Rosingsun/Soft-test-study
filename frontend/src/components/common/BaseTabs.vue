<script setup lang="ts">
import { ref, watch, useSlots } from 'vue'

const props = withDefaults(defineProps<{
  modelValue?: string | number
  tabs: { label: string; value: string | number; badge?: string | number }[]
  variant?: 'gradient' | 'pill'
  size?: 'sm' | 'md'
}>(), {
  modelValue: '',
  variant: 'gradient',
  size: 'md',
})

const emit = defineEmits<{ (e: 'update:modelValue', v: string | number): void }>()
const active = ref<string | number>(props.modelValue)

watch(() => props.modelValue, v => { if (v !== undefined && v !== '') active.value = v })
watch(active, v => emit('update:modelValue', v))

const slots = useSlots()
function hasIcon(value: string | number) {
  return !!slots[`icon-${value}`]
}
</script>

<template>
  <div
    :class="[
      'inline-flex rounded-2xl bg-gray-100/70 p-1.5 ring-1 ring-inset ring-gray-200/60 backdrop-blur',
      size === 'sm' && 'p-1',
    ]"
    role="tablist"
  >
    <button
      v-for="tab in tabs"
      :key="tab.value"
      type="button"
      role="tab"
      :aria-selected="active === tab.value"
      :class="[
        'group relative flex flex-1 cursor-pointer items-center justify-center gap-1.5 whitespace-nowrap rounded-xl font-semibold transition-all duration-200 active:scale-[0.97]',
        size === 'sm' ? 'px-3 py-1.5 text-xs' : 'px-4 py-2 text-sm',
        active === tab.value
          ? variant === 'gradient'
            ? 'bg-brand-gradient text-white shadow-md shadow-indigo-600/25 ring-1 ring-white/30'
            : 'bg-white text-indigo-600 shadow-sm ring-1 ring-gray-200/60'
          : 'text-gray-500 hover:bg-white/70 hover:text-gray-800',
      ]"
      @click="active = tab.value"
    >
      <span
        v-if="hasIcon(tab.value)"
        :class="[
          'flex h-4 w-4 items-center justify-center transition-transform duration-200',
          !active || active !== tab.value || variant === 'gradient' ? 'group-hover:scale-110' : '',
        ]"
      >
        <slot :name="`icon-${tab.value}`" />
      </span>
      <span>{{ tab.label }}</span>
      <span
        v-if="tab.badge !== undefined"
        :class="[
          'rounded-full px-1.5 py-0.5 text-[10px] font-bold tabular-nums transition-colors',
          active === tab.value
            ? variant === 'gradient'
              ? 'bg-white/25 text-white'
              : 'bg-indigo-50 text-indigo-600'
            : 'bg-gray-200/80 text-gray-600',
        ]"
      >
        {{ tab.badge }}
      </span>
    </button>
  </div>
</template>
