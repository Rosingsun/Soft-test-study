<script setup lang="ts">
import { ref } from 'vue'
import BaseButton from '@/components/common/BaseButton.vue'

withDefaults(defineProps<{
  title: string
  description?: string
  confirmText?: string
  danger?: boolean
}>(), {
  confirmText: '确认',
  danger: false,
})

const emit = defineEmits<{ (e: 'confirm'): void; (e: 'cancel'): void }>()
const visible = ref(false)

function open() {
  visible.value = true
}

function confirm() {
  visible.value = false
  emit('confirm')
}

function cancel() {
  visible.value = false
  emit('cancel')
}

defineExpose({ open })
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="visible" class="fixed inset-0 z-[110] flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-gray-900/50 backdrop-blur-sm" @click="cancel" />
        <div class="relative w-full max-w-sm rounded-xl bg-white p-6 shadow-2xl">
          <div class="mb-4 flex items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
              :class="danger ? 'bg-red-50' : 'bg-indigo-50'"
            >
              <svg
                class="h-5 w-5"
                :class="danger ? 'text-red-500' : 'text-indigo-600'"
                fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <div>
              <h3 class="text-base font-semibold text-gray-900">{{ title }}</h3>
              <p v-if="description" class="mt-1 text-sm text-gray-500">{{ description }}</p>
            </div>
          </div>
          <div class="flex justify-end gap-2">
            <BaseButton type="secondary" size="sm" @click="cancel">取消</BaseButton>
            <BaseButton :type="danger ? 'danger' : 'primary'" size="sm" @click="confirm">
              {{ confirmText }}
            </BaseButton>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.15s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
