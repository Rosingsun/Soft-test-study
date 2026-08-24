<script setup lang="ts">
import { ref, computed } from 'vue'
import type { MindMapNode } from '@/types/material'

const props = withDefaults(defineProps<{
  node: MindMapNode
  depth?: number
  initialExpand?: 'all' | 'none' | null
  keyword?: string
}>(), {
  depth: 0,
  initialExpand: null,
  keyword: '',
})

const expanded = ref(
  props.initialExpand === 'all' ? true : props.depth < 2,
)

const hasChildren = computed(() => !!props.node.children?.length)
const isOpen = computed(() => expanded.value)

function toggle() {
  if (hasChildren.value) expanded.value = !expanded.value
}

const segments = computed(() => {
  const kw = props.keyword.trim().toLowerCase()
  const title = props.node.title
  if (!kw) return [{ text: title, hit: false }]
  const lower = title.toLowerCase()
  const result: { text: string; hit: boolean }[] = []
  let index = 0
  while (index < title.length) {
    const found = lower.indexOf(kw, index)
    if (found === -1) {
      result.push({ text: title.slice(index), hit: false })
      break
    }
    if (found > index) result.push({ text: title.slice(index, found), hit: false })
    result.push({ text: title.slice(found, found + kw.length), hit: true })
    index = found + kw.length
  }
  return result
})
</script>

<template>
  <li>
    <div
      :class="[
        'flex items-start gap-1.5 rounded-lg px-2 py-1.5 transition-colors duration-200',
        hasChildren ? 'cursor-pointer hover:bg-indigo-50/50' : '',
      ]"
      @click="toggle"
    >
      <svg
        v-if="hasChildren"
        :class="[
          'mt-1 h-3.5 w-3.5 shrink-0 text-gray-400 transition-transform duration-200',
          isOpen ? 'rotate-90' : '',
        ]"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
      </svg>
      <span
        v-else
        class="mt-[9px] h-1 w-1 shrink-0 rounded-full"
        :class="depth <= 1 ? 'bg-indigo-400' : 'bg-gray-300'"
      />
      <span
        :class="[
          'min-w-0 flex-1 break-words leading-6',
          depth === 0 ? 'text-[15px] font-semibold text-gray-900'
            : depth === 1 ? 'text-sm font-medium text-gray-800'
            : 'text-sm text-gray-700',
        ]"
      >
        <span
          v-for="(seg, i) in segments"
          :key="i"
          :class="seg.hit ? 'rounded bg-amber-100 px-0.5 text-amber-700' : ''"
        >{{ seg.text }}</span>
        <span
          v-if="hasChildren && !isOpen"
          class="ml-1.5 align-middle text-xs font-normal text-gray-400"
        >{{ node.children!.length }}</span>
      </span>
    </div>

    <ul
      v-if="hasChildren && isOpen"
      class="ml-[15px] space-y-0.5 border-l border-gray-100 pl-3"
    >
      <MindMapTree
        v-for="(child, i) in node.children"
        :key="i"
        :node="child"
        :depth="depth + 1"
        :keyword="keyword"
      />
    </ul>
  </li>
</template>
