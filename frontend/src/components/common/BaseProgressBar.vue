<script setup lang="ts">
withDefaults(defineProps<{
  value: number
  color?: 'indigo' | 'emerald' | 'red' | 'yellow' | 'gray'
  size?: 'sm' | 'md'
}>(), {
  color: 'indigo',
  size: 'md',
})

const colorMap = {
  indigo: 'bg-gradient-to-r from-indigo-500 to-violet-500',
  emerald: 'bg-gradient-to-r from-emerald-500 to-teal-400',
  red: 'bg-gradient-to-r from-red-500 to-rose-400',
  yellow: 'bg-gradient-to-r from-amber-400 to-yellow-400',
  gray: 'bg-gray-300',
}
</script>

<template>
  <div class="w-full overflow-hidden rounded-full bg-gray-100">
    <div
      :class="['h-full rounded-full transition-all duration-500', colorMap[color], size === 'sm' ? 'h-1.5' : 'h-2']"
      :style="{ width: Math.min(100, Math.max(0, Number(value) || 0)) + '%' }"
    />
  </div>
</template>
