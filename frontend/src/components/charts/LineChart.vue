<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = withDefaults(defineProps<{
  data: { label: string; value: number; hint?: string }[]
  height?: number
  color?: string
  emptyText?: string
  /** 阈值参考线，0 = 不显示；正值时按 value 着色点（>= 阈值为 ok 色，< 为 warn 色） */
  threshold?: number
  thresholdLabel?: string
  /** 是否在每个数据点上方显示数值 */
  showValues?: boolean
  /** 是否使用品牌渐变描边（indigo → violet） */
  gradient?: boolean
  /** 是否使用平滑曲线（贝塞尔） */
  smooth?: boolean
  /** 达标点颜色（默认 emerald-500） */
  okColor?: string
  /** 未达标点颜色（默认 rose-500） */
  warnColor?: string
  /** 是否启用悬停 tooltip */
  interactive?: boolean
  /** 是否强调最后一个数据点（外发光圆环） */
  emphasizeLast?: boolean
  /** 是否在阈值线下方填充警示色带 */
  thresholdBand?: boolean
  /** 数值标签节流：数据点超过该数量时只显示关键点数值（首尾 + 最高 + 最新） */
  valueLabelLimit?: number
}>(), {
  height: 280,
  color: '#4f46e5',
  emptyText: '暂无数据',
  threshold: 0,
  thresholdLabel: '',
  showValues: false,
  gradient: false,
  smooth: false,
  okColor: '#10b981',
  warnColor: '#f43f5e',
  interactive: false,
  emphasizeLast: false,
  thresholdBand: false,
  valueLabelLimit: 8,
})

const WIDTH = 600
const PAD_X = 20
const PAD_TOP = 36
const PAD_BOTTOM = 32

const max = computed(() => {
  if (!props.data.length) return 100
  const dataMax = Math.max(...props.data.map(d => d.value))
  return Math.max(dataMax, props.threshold || 0, 100)
})

function pointColor(value: number) {
  if (props.threshold > 0) {
    return value >= props.threshold ? props.okColor : props.warnColor
  }
  return props.color
}

const points = computed(() => {
  const n = props.data.length
  if (!n) return []
  const span = WIDTH - PAD_X * 2
  const maxVal = max.value
  return props.data.map((d, i) => {
    const x = n === 1 ? WIDTH / 2 : PAD_X + (i / (n - 1)) * span
    const y = PAD_TOP + (1 - d.value / maxVal) * (props.height - PAD_TOP - PAD_BOTTOM)
    return { ...d, x, y, color: pointColor(d.value) }
  })
})

// 需要显示数值标签的索引集合（节流）
const labeledIndexes = computed(() => {
  const n = points.value.length
  if (!n || !props.showValues) return new Set<number>()
  if (n <= props.valueLabelLimit) return new Set(points.value.map((_, i) => i))
  const set = new Set<number>()
  set.add(0)
  set.add(n - 1)
  let best = 0
  points.value.forEach((p, i) => { if (p.value > points.value[best].value) best = i })
  set.add(best)
  return set
})

const bestIndex = computed(() => {
  let best = 0
  points.value.forEach((p, i) => { if (p.value > points.value[best].value) best = i })
  return points.value.length ? best : -1
})

// 折线路径（直线 / 平滑曲线二选一）
const linePath = computed(() => {
  if (!points.value.length) return ''
  if (!props.smooth) {
    return points.value.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x},${p.y}`).join(' ')
  }
  // 三次贝塞尔平滑（中点控制）
  let path = `M${points.value[0].x},${points.value[0].y}`
  for (let i = 1; i < points.value.length; i++) {
    const prev = points.value[i - 1]
    const curr = points.value[i]
    const cpX = (prev.x + curr.x) / 2
    path += ` C${cpX},${prev.y} ${cpX},${curr.y} ${curr.x},${curr.y}`
  }
  return path
})

const areaPath = computed(() => {
  if (!points.value.length) return ''
  const first = points.value[0]
  const last = points.value[points.value.length - 1]
  return `${linePath.value} L${last.x},${props.height - PAD_BOTTOM} L${first.x},${props.height - PAD_BOTTOM} Z`
})

// 阈值参考线 Y
const thresholdY = computed(() => {
  if (!props.threshold) return null
  return PAD_TOP + (1 - props.threshold / max.value) * (props.height - PAD_TOP - PAD_BOTTOM)
})

// 各 ID（避免多实例同色冲突）
const idBase = computed(() => `lc-${Math.random().toString(36).slice(2, 8)}`)

// ============ 交互 tooltip ============
const svgRef = ref<SVGSVGElement | null>(null)
const hoverIndex = ref(-1)

watch(points, () => { hoverIndex.value = -1 })

const hovered = computed(() => {
  if (!props.interactive || hoverIndex.value < 0) return null
  return points.value[hoverIndex.value] ?? null
})

function onMove(e: MouseEvent) {
  if (!props.interactive || !svgRef.value || !points.value.length) return
  const rect = svgRef.value.getBoundingClientRect()
  const ratioX = WIDTH / rect.width
  const ratioY = props.height / rect.height
  const x = (e.clientX - rect.left) * ratioX
  const y = (e.clientY - rect.top) * ratioY
  let nearest = 0
  let best = Infinity
  points.value.forEach((p, i) => {
    const d = Math.hypot(p.x - x, p.y - y)
    if (d < best) { best = d; nearest = i }
  })
  if (best < 40) hoverIndex.value = nearest
}

// tooltip 定位：相对包裹容器百分比；靠近顶部时翻转到点下方，贴近左右边缘时居中钳制
const tooltipStyle = computed(() => {
  const p = hovered.value
  if (!p) return {}
  const left = Math.min(88, Math.max(12, (p.x / WIDTH) * 100))
  return { left: `${left}%`, top: `${(p.y / props.height) * 100}%` }
})

const tooltipFlip = computed(() => {
  const p = hovered.value
  return !!p && p.y < PAD_TOP + 22
})
</script>

<template>
  <div v-if="!points.length" class="flex items-center justify-center text-xs text-gray-400" :style="{ height: height + 'px' }">
    {{ emptyText }}
  </div>

  <div v-else class="relative" :style="{ height: height + 'px' }">
    <svg
      ref="svgRef"
      :viewBox="`0 0 ${WIDTH} ${height}`"
      class="block w-full"
      :style="{ height: height + 'px' }"
      :class="interactive ? 'cursor-pointer' : ''"
      @mousemove="onMove"
      @mouseleave="hoverIndex = -1"
    >
      <defs>
        <linearGradient :id="`${idBase}-area`" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" :stop-color="color" stop-opacity="0.32" />
          <stop offset="100%" :stop-color="color" stop-opacity="0" />
        </linearGradient>
        <linearGradient :id="`${idBase}-line`" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stop-color="#4f46e5" />
          <stop offset="100%" stop-color="#7c3aed" />
        </linearGradient>
      </defs>

      <!-- 阈值下方警示色带 -->
      <rect
        v-if="thresholdBand && thresholdY !== null"
        :x="PAD_X" :y="thresholdY" :width="WIDTH - PAD_X * 2"
        :height="height - PAD_BOTTOM - (thresholdY || 0)"
        :fill="warnColor" opacity="0.05" rx="6"
      />

      <!-- 阈值参考线 -->
      <g v-if="thresholdY !== null">
        <line
          :x1="PAD_X" :y1="thresholdY" :x2="WIDTH - PAD_X" :y2="thresholdY"
          :stroke="okColor" stroke-width="1.2" stroke-dasharray="5 4" opacity="0.65"
        />
        <rect
          :x="WIDTH - PAD_X - 90" :y="thresholdY - 18"
          width="86" height="18" rx="4"
          :fill="okColor" opacity="0.15"
        />
        <text
          :x="WIDTH - PAD_X - 47" :y="thresholdY - 5"
          text-anchor="middle"
          :fill="okColor"
          font-size="11" font-weight="600"
        >{{ thresholdLabel || `及格线 ${threshold}%` }}</text>
      </g>

      <!-- 面积 + 折线 -->
      <path :d="areaPath" :fill="`url(#${idBase}-area)`" />
      <path
        :d="linePath" fill="none"
        :stroke="gradient ? `url(#${idBase}-line)` : color"
        stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round"
      />

      <!-- 数据点 + 标签 -->
      <g v-for="(p, i) in points" :key="i">
        <title>{{ p.label }}：{{ p.value }}%{{ p.hint ? ' · ' + p.hint : '' }}</title>

        <!-- 数值标签（节流后） -->
        <text
          v-if="labeledIndexes.has(i)"
          :x="p.x" :y="p.y - (emphasizeLast && i === points.length - 1 ? 20 : 14)"
          text-anchor="middle"
          :fill="i === bestIndex && i === points.length - 1 ? '#d97706' : p.color"
          font-size="12" font-weight="700"
        >{{ p.value }}%</text>

        <!-- 强调最新点：外发光圆环 + 略大 -->
        <g v-if="emphasizeLast && i === points.length - 1">
          <circle :cx="p.x" :cy="p.y" r="9" fill="#f59e0b" opacity="0.18" />
          <circle :cx="p.x" :cy="p.y" r="7" fill="#fff" stroke="#f59e0b" stroke-width="3" />
          <circle :cx="p.x" :cy="p.y" r="3.4" fill="#f59e0b" />
        </g>
        <!-- 普通点 -->
        <g v-else>
          <circle
            :cx="p.x" :cy="p.y" r="6" fill="#fff" :stroke="p.color" stroke-width="2.5"
            :class="hoverIndex === i ? 'transition-all duration-150' : ''"
          />
          <circle :cx="p.x" :cy="p.y" r="2.8" :fill="p.color" />
        </g>

        <!-- x 轴标签 -->
        <text
          :x="p.x" :y="height - 10"
          text-anchor="middle"
          class="fill-gray-500"
          font-size="11"
        >{{ p.label }}</text>
      </g>
    </svg>

    <!-- 悬停 tooltip -->
    <div
      v-if="hovered"
      class="pointer-events-none absolute z-10 flex -translate-x-1/2 flex-col items-center"
      :class="tooltipFlip ? 'translate-y-3' : '-translate-y-full -mt-3'"
      :style="tooltipStyle"
    >
      <template v-if="tooltipFlip">
        <div class="h-0 w-0 border-x-4 border-b-4 border-x-transparent border-b-gray-800/95" />
        <div class="rounded-lg bg-gray-800/95 px-2.5 py-1.5 text-center text-white shadow-lg backdrop-blur-sm">
          <p class="text-[11px] font-semibold">{{ hovered.label }}</p>
          <p class="mt-0.5 text-sm font-bold tabular-nums" :style="{ color: hovered.color }">{{ hovered.value }}%</p>
          <p v-if="hovered.hint" class="mt-0.5 text-[10px] text-gray-300">{{ hovered.hint }}</p>
        </div>
      </template>
      <template v-else>
        <div class="rounded-lg bg-gray-800/95 px-2.5 py-1.5 text-center text-white shadow-lg backdrop-blur-sm">
          <p class="text-[11px] font-semibold">{{ hovered.label }}</p>
          <p class="mt-0.5 text-sm font-bold tabular-nums" :style="{ color: hovered.color }">{{ hovered.value }}%</p>
          <p v-if="hovered.hint" class="mt-0.5 text-[10px] text-gray-300">{{ hovered.hint }}</p>
        </div>
        <div class="h-0 w-0 border-x-4 border-t-4 border-x-transparent border-t-gray-800/95" />
      </template>
    </div>
  </div>
</template>
