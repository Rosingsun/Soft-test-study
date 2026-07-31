<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  percentage: number
  size?: number
  stroke?: number
  label?: string
  color?: string
  trackColor?: string
}>(), {
  size: 120,
  stroke: 12,
  color: '#4f46e5',
  trackColor: '#e5e7eb',
})

const radius = computed(() => (props.size - props.stroke) / 2)
const circumference = computed(() => 2 * Math.PI * radius.value)
const offset = computed(() => {
  const p = Math.min(100, Math.max(0, props.percentage))
  return circumference.value * (1 - p / 100)
})

const colorClass = computed(() => {
  const p = props.percentage
  if (p >= 80) return 'text-emerald-500'
  if (p >= 60) return 'text-indigo-600'
  if (p >= 40) return 'text-yellow-500'
  return 'text-red-500'
})
</script>

<template>
  <div class="relative inline-flex items-center justify-center">
    <svg :width="size" :height="size" :viewBox="`0 0 ${size} ${size}`">
      <circle
        :cx="size / 2" :cy="size / 2" :r="radius"
        :stroke="trackColor" stroke-width="stroke" fill="none"
      />
      <circle
        :cx="size / 2" :cy="size / 2" :r="radius"
        :stroke="color" stroke-width="stroke" fill="none"
        stroke-linecap="round"
        :stroke-dasharray="circumference"
        :stroke-dashoffset="offset"
        :transform="`rotate(-90 ${size / 2} ${size / 2})`"
        class="transition-all duration-700"
      />
    </svg>
    <div class="absolute inset-0 flex flex-col items-center justify-center">
      <span :class="['text-xl font-bold', colorClass]">{{ percentage.toFixed(0) }}%</span>
      <span v-if="label" class="text-[10px] text-gray-400">{{ label }}</span>
    </div>
  </div>
</template>
