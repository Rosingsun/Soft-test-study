import { useToastStore, type ToastType } from '@/stores/toast'

export function showToast(message: string, type: ToastType = 'info') {
  useToastStore().show(message, type)
}
