<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useToastStore } from '@/stores/toast'

const { message, visible } = storeToRefs(useToastStore())
</script>

<template>
  <Transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="translate-y-2 opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="translate-y-2 opacity-0"
  >
    <div
      v-if="visible"
      class="fixed bottom-6 left-1/2 z-[100] max-w-md -translate-x-1/2 rounded-lg bg-gray-800 px-4 py-3 text-sm text-white shadow-lg"
      role="alert"
    >
      {{ message }}
    </div>
  </Transition>
</template>
