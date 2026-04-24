import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ToastKind = 'success' | 'error' | 'warning' | 'info'

export interface ToastItem {
  id: number
  kind: ToastKind
  title?: string
  message: string
  duration: number
}

export interface ToastInput {
  kind?: ToastKind
  title?: string
  message: string
  duration?: number
}

let seed = 0

export const useToastStore = defineStore('toast', () => {
  const items = ref<ToastItem[]>([])
  const MAX_VISIBLE = 5

  function push(input: ToastInput): number {
    const id = ++seed
    const item: ToastItem = {
      id,
      kind: input.kind ?? 'info',
      title: input.title,
      message: input.message,
      duration: input.duration ?? 3200,
    }
    items.value.push(item)
    if (items.value.length > MAX_VISIBLE) {
      items.value.splice(0, items.value.length - MAX_VISIBLE)
    }
    if (item.duration > 0) {
      window.setTimeout(() => dismiss(id), item.duration)
    }
    return id
  }

  function dismiss(id: number) {
    const idx = items.value.findIndex((t) => t.id === id)
    if (idx >= 0) items.value.splice(idx, 1)
  }

  function success(message: string, title?: string) {
    return push({ kind: 'success', message, title })
  }
  function error(message: string, title?: string) {
    return push({ kind: 'error', message, title, duration: 5000 })
  }
  function info(message: string, title?: string) {
    return push({ kind: 'info', message, title })
  }
  function warning(message: string, title?: string) {
    return push({ kind: 'warning', message, title })
  }

  return { items, push, dismiss, success, error, info, warning }
})
