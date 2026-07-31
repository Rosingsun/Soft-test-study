import { useToastStore } from '@/stores/toast'

export function showToast(message: string) {
  useToastStore().show(message)
}
