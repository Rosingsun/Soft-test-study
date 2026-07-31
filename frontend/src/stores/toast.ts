import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useToastStore = defineStore('toast', () => {
  const message = ref('')
  const visible = ref(false)
  let timer: number | null = null

  function show(msg: string) {
    message.value = msg
    visible.value = true
    if (timer !== null) {
      clearTimeout(timer)
    }
    timer = window.setTimeout(() => {
      visible.value = false
    }, 3000)
  }

  return { message, visible, show }
})
