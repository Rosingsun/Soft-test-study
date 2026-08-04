import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ToastType = 'success' | 'error' | 'info'

export const useToastStore = defineStore('toast', () => {
  const message = ref('')
  const type = ref<ToastType>('info')
  const visible = ref(false)
  let timer: number | null = null

  function show(msg: string, t: ToastType = 'info') {
    message.value = msg
    type.value = t
    visible.value = true
    if (timer !== null) {
      clearTimeout(timer)
    }
    timer = window.setTimeout(() => {
      visible.value = false
    }, 3000)
  }

  return { message, type, visible, show }
})
