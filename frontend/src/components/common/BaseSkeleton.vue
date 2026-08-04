<script setup lang="ts">
withDefaults(defineProps<{
  /** list: 列表行；grid: 卡片网格；detail: 详情/看板；question: 答题页 */
  variant?: 'list' | 'grid' | 'detail' | 'question'
  /** 重复项数量（list/grid 有效） */
  count?: number
}>(), {
  variant: 'list',
  count: 3,
})
</script>

<template>
  <!-- 列表骨架：徽章 + 两行文本 -->
  <div v-if="variant === 'list'" class="space-y-3">
    <div
      v-for="i in count"
      :key="i"
      class="rounded-xl border border-gray-100 bg-white p-5 shadow-sm"
    >
      <div class="mb-3 flex items-center gap-2">
        <div class="skeleton h-5 w-14 rounded-full" />
        <div class="skeleton h-3.5 w-32" />
      </div>
      <div class="skeleton mb-2 h-4 w-full" />
      <div class="skeleton h-4 w-2/3" />
    </div>
  </div>

  <!-- 网格骨架：卡片标题 + 文本 + 按钮 -->
  <div v-else-if="variant === 'grid'" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
    <div
      v-for="i in count"
      :key="i"
      class="flex flex-col rounded-xl border border-gray-100 bg-white p-5 shadow-sm"
    >
      <div class="mb-3 flex items-center justify-between">
        <div class="skeleton h-5 w-2/5" />
        <div class="skeleton h-5 w-12 rounded-full" />
      </div>
      <div class="skeleton mb-4 h-3.5 w-3/5" />
      <div class="mb-4 grid grid-cols-3 gap-2">
        <div class="skeleton h-12 rounded-xl" />
        <div class="skeleton h-12 rounded-xl" />
        <div class="skeleton h-12 rounded-xl" />
      </div>
      <div class="skeleton mt-auto h-9 rounded-lg" />
    </div>
  </div>

  <!-- 详情骨架：头部横幅 + 统计芯片 + 内容块 -->
  <div v-else-if="variant === 'detail'" class="space-y-6">
    <div class="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm sm:p-7">
      <div class="skeleton mb-3 h-6 w-2/5" />
      <div class="skeleton h-4 w-3/5" />
    </div>
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div v-for="i in 4" :key="i" class="rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
        <div class="skeleton mb-3 h-9 w-9 rounded-lg" />
        <div class="skeleton mb-2 h-6 w-1/2" />
        <div class="skeleton h-3 w-2/3" />
      </div>
    </div>
    <div class="grid gap-6 lg:grid-cols-5">
      <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm lg:col-span-3">
        <div class="skeleton mb-4 h-5 w-1/3" />
        <div class="skeleton h-40 w-full rounded-xl" />
      </div>
      <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm lg:col-span-2">
        <div class="skeleton mb-4 h-5 w-1/3" />
        <div class="space-y-4">
          <div v-for="i in 3" :key="i">
            <div class="skeleton mb-1.5 h-3.5 w-full" />
            <div class="skeleton h-2 w-full rounded-full" />
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- 答题页骨架：题卡 + 选项 + 底部操作栏 -->
  <div v-else class="mx-auto max-w-4xl">
    <div class="mb-4 flex items-center justify-between">
      <div class="skeleton h-8 w-20 rounded-lg" />
      <div class="skeleton h-8 w-32 rounded-lg" />
    </div>
    <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm sm:p-7">
      <div class="mb-5 flex gap-2">
        <div class="skeleton h-5 w-16 rounded-full" />
        <div class="skeleton h-5 w-14 rounded-full" />
      </div>
      <div class="skeleton mb-2 h-4 w-full" />
      <div class="skeleton mb-6 h-4 w-4/5" />
      <div class="space-y-2.5">
        <div v-for="i in 4" :key="i" class="flex items-center gap-3.5 rounded-xl border-2 border-gray-100 px-4 py-3.5">
          <div class="skeleton h-7 w-7 shrink-0 rounded-full" />
          <div class="skeleton h-4 flex-1" />
        </div>
      </div>
    </div>
    <div class="mt-4 flex items-center justify-between rounded-xl border border-gray-100 bg-white p-3 shadow-sm">
      <div class="flex gap-2">
        <div class="skeleton h-9 w-24 rounded-lg" />
        <div class="skeleton h-9 w-24 rounded-lg" />
      </div>
      <div class="skeleton h-9 w-28 rounded-lg" />
    </div>
  </div>
</template>
