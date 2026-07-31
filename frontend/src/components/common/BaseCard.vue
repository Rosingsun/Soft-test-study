<script setup lang="ts">
withDefaults(defineProps<{
  padding?: boolean
  hover?: boolean
}>(), {
  padding: true,
  hover: false,
})
</script>

<template>
  <div
    :class="[
      'rounded-lg border border-gray-200 bg-white shadow-sm',
      padding && 'p-5',
      hover && 'transition-all hover:-translate-y-0.5 hover:shadow-md',
    ]"
  >
    <slot />
  </div>
</template>
