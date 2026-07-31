<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  data: { label: string; value: number; hint?: string }[]
  height?: number
  color?: string
  emptyColor?: string
}>(), {
  height: 180,
  color: 'bg-indigo-500',
  emptyColor: 'bg-gray-100',
})

const max = computed(() => {
  if (!props.data.length) return 1
  return Math.max(...props.data.map(d => d.value), 1)
})

const bars = computed(() => {
  return props.data.map((d, i) => ({
    ...d,
    pct: d.value / max.value,
    showLabel: i % Math.max(1, Math.floor(props.data.length / 12)) === 0,
  }))
})
</script>

<template>
  <div>
    <div class="flex items-end gap-1" :style="{ height: height + 'px' }">
      <div
        v-for="(bar, idx) in bars"
        :key="idx"
        class="group relative flex flex-1 flex-col justify-end"
      >
        <div
          class="w-full rounded-t transition-all duration-500"
          :class="bar.value === 0 ? emptyColor : color"
          :style="{ height: Math.max(bar.value > 0 ? 6 : 3, bar.pct * (height - 20)) + 'px' }"
        />
        <div class="pointer-events-none absolute -top-1 left-1/2 z-10 hidden -translate-x-1/2 whitespace-nowrap rounded-md bg-gray-800 px-2 py-1 text-xs text-white group-hover:block">
          {{ bar.label }}: {{ bar.value }} {{ bar.hint || '' }}
        </div>
      </div>
    </div>
    <div class="mt-2 flex gap-1 text-[10px] text-gray-400">
      <span v-for="(bar, idx) in bars" :key="idx" class="flex-1 truncate text-center" :title="bar.label">
        {{ bar.showLabel ? bar.label : '' }}
      </span>
    </div>
  </div>
</template>
