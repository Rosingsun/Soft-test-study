<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  data: { label: string; value: number; hint?: string }[]
  height?: number
  color?: string
  emptyText?: string
}>(), {
  height: 200,
  color: '#4f46e5',
  emptyText: '暂无数据',
})

const WIDTH = 300
const PAD = 24

const max = computed(() => {
  if (!props.data.length) return 100
  return Math.max(...props.data.map(d => d.value), 100)
})

const points = computed(() => {
  const n = props.data.length
  if (!n) return []
  const span = WIDTH - PAD * 2
  return props.data.map((d, i) => {
    const x = n === 1 ? WIDTH / 2 : PAD + (i / (n - 1)) * span
    const y = PAD + (1 - d.value / max.value) * (props.height - PAD * 2)
    return { ...d, x, y }
  })
})

const linePath = computed(() => points.value.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x},${p.y}`).join(' '))

const areaPath = computed(() => {
  if (!points.value.length) return ''
  const first = points.value[0]
  const last = points.value[points.value.length - 1]
  return `${linePath.value} L${last.x},${props.height - PAD} L${first.x},${props.height - PAD} Z`
})
</script>

<template>
  <div v-if="!points.length" class="flex items-center justify-center text-xs text-gray-400" :style="{ height: height + 'px' }">
    {{ emptyText }}
  </div>
  <svg v-else :viewBox="`0 0 ${WIDTH} ${height}`" class="w-full" :style="{ height: height + 'px' }">
    <defs>
      <linearGradient :id="'grad-' + color.replace('#', '')" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" :stop-color="color" stop-opacity="0.2" />
        <stop offset="100%" :stop-color="color" stop-opacity="0" />
      </linearGradient>
    </defs>
    <path :d="areaPath" :fill="`url(#grad-${color.replace('#', '')})`" />
    <path :d="linePath" fill="none" :stroke="color" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
    <g v-for="(p, i) in points" :key="i">
      <circle :cx="p.x" :cy="p.y" r="3" :fill="color" stroke="#fff" stroke-width="1.5" />
      <text
        :x="p.x"
        :y="height - 6"
        text-anchor="middle"
        class="fill-gray-400"
        font-size="9"
      >{{ p.label }}</text>
      <title>{{ p.label }}：{{ p.value }}%{{ p.hint ? ' · ' + p.hint : '' }}</title>
    </g>
  </svg>
</template>
