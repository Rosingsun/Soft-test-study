<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { Markmap } from 'markmap-view'
import type { INode, IPureNode } from 'markmap-common'
import type { MindMapNode } from '@/types/material'

const props = withDefaults(defineProps<{
  nodes: MindMapNode[]
  keyword?: string
}>(), {
  keyword: '',
})

const svgRef = ref<SVGSVGElement | null>(null)
let mm: Markmap | null = null

function toPure(n: MindMapNode): IPureNode {
  return {
    content: n.title,
    children: n.children?.length ? n.children.map(toPure) : [],
  }
}

function findFirstMatch(nodes: INode[], kw: string): INode | null {
  for (const n of nodes) {
    if (n.content.toLowerCase().includes(kw)) return n
    if (n.children?.length) {
      const hit = findFirstMatch(n.children, kw)
      if (hit) return hit
    }
  }
  return null
}

onMounted(() => {
  if (!svgRef.value) return
  mm = Markmap.create(svgRef.value, {
    autoFit: true,
    duration: 300,
    initialExpandLevel: -1,
    embedGlobalCSS: true,
    color: (node) => {
      const depth = node.state?.depth ?? 0
      const palette = ['#4f46e5', '#6366f1', '#7c3aed', '#a78bfa', '#c4b5fd', '#8b5cf6']
      return palette[depth % palette.length]
    },
  })
  if (props.nodes.length) {
    void mm.setData({ content: '根节点', children: props.nodes.map(toPure) })
  }
})

watch(
  () => props.nodes,
  (val) => {
    if (!mm) return
    void mm.setData({ content: '根节点', children: val.map(toPure) })
  },
  { deep: true },
)

watch(
  () => props.keyword,
  (val) => {
    if (!mm) return
    const kw = val.trim().toLowerCase()
    if (!kw) {
      void mm.setHighlight(null)
      return
    }
    const root = mm.state.data
    if (!root) return
    const hit = findFirstMatch(root.children ?? [], kw)
    void mm.setHighlight(hit)
  },
)

onBeforeUnmount(() => {
  mm?.destroy()
  mm = null
})
</script>

<template>
  <div class="relative h-[calc(100vh-16rem)] min-h-[32rem] w-full overflow-hidden rounded-lg bg-gradient-to-br from-slate-50 to-indigo-50/40">
    <svg ref="svgRef" class="h-full w-full" />
    <div class="pointer-events-none absolute bottom-3 right-3 rounded-full bg-white/80 px-2.5 py-1 text-[11px] font-medium text-gray-500 shadow-sm ring-1 ring-gray-200/60 backdrop-blur">
      滚轮缩放 · 拖拽平移 · 点击节点展开
    </div>
  </div>
</template>
