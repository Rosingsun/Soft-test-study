<script setup lang="ts">
import { ref, computed, watch } from 'vue'
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

const expanded = ref(props.initialExpand === 'all' ? true : props.depth < 1)

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

watch(() => props.initialExpand, (val) => {
  if (val === 'all') expanded.value = true
  else if (val === 'none') expanded.value = false
})
</script>

<template>
  <section>
    <component
      :is="depth === 0 ? 'h1' : depth === 1 ? 'h2' : depth === 2 ? 'h3' : 'h4'"
      :class="[
        depth === 0 ? 'text-xl font-bold tracking-tight text-gray-900 border-b border-gray-200 pb-2 mb-3 mt-2'
          : depth === 1 ? 'text-base font-semibold text-indigo-700 mt-4 mb-1.5'
          : depth === 2 ? 'text-sm font-semibold text-gray-800 mt-2.5 mb-1'
          : 'text-sm text-gray-700 mt-1.5 mb-0.5',
        hasChildren ? 'cursor-pointer select-none hover:text-indigo-600 transition-colors' : '',
      ]"
      @click="toggle"
    >
      <svg
        v-if="hasChildren"
        :class="[
          'inline-block h-3.5 w-3.5 mr-1.5 align-[-2px] text-gray-400 transition-transform duration-200',
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
        class="inline-block h-1.5 w-1.5 mr-2 align-middle rounded-full bg-indigo-300"
      />
      <span
        v-for="(seg, i) in segments"
        :key="i"
        :class="seg.hit ? 'rounded bg-amber-100 px-0.5 text-amber-700' : ''"
      >{{ seg.text }}</span>
      <span
        v-if="hasChildren && !isOpen"
        class="ml-1.5 align-middle text-xs font-normal text-gray-400"
      >{{ node.children!.length }}</span>
    </component>

    <div
      v-if="hasChildren && isOpen"
      :class="[
        depth === 0 ? 'ml-4 space-y-0.5 border-l-2 border-indigo-100 pl-5'
          : depth === 1 ? 'ml-3 space-y-0.5 border-l border-gray-100 pl-4'
          : 'ml-2 space-y-0.5 border-l border-gray-50 pl-3',
      ]"
    >
      <ChapterDocNode
        v-for="(child, i) in node.children"
        :key="i"
        :node="child"
        :depth="depth + 1"
        :initial-expand="initialExpand"
        :keyword="keyword"
      />
    </div>
  </section>
</template>
