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
      'rounded-xl border border-gray-100 bg-white shadow-sm',
      padding && 'p-5',
      hover && 'card-lift cursor-pointer',
    ]"
  >
    <slot />
  </div>
</template>
