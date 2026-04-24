<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { useThemeStore, type ThemeMode } from '@/stores/theme'

const theme = useThemeStore()

let unbind: (() => void) | null = null

onMounted(() => {
  unbind = theme.bindSystemListener()
})

onBeforeUnmount(() => {
  unbind?.()
})

const OPTIONS: Array<{ value: ThemeMode; label: string; icon: string }> = [
  {
    value: 'light',
    label: '浅色',
    icon: 'M12 3v1.5M12 19.5V21M4.22 4.22l1.06 1.06M18.72 18.72l1.06 1.06M3 12h1.5M19.5 12H21M4.22 19.78l1.06-1.06M18.72 5.28l1.06-1.06M12 8a4 4 0 1 1 0 8 4 4 0 0 1 0-8z',
  },
  {
    value: 'auto',
    label: '跟随',
    icon: 'M12 3a9 9 0 1 0 9 9 9 9 0 0 0 -9 -9zm0 18V3',
  },
  {
    value: 'dark',
    label: '深色',
    icon: 'M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z',
  },
]
</script>

<template>
  <div
    role="radiogroup"
    aria-label="主题切换"
    class="relative inline-flex items-center rounded-full border border-base-300/70 bg-base-200/60 p-0.5 backdrop-blur"
  >
    <span
      aria-hidden="true"
      :class="[
        'absolute top-0.5 bottom-0.5 w-[calc(100%/3)] rounded-full transition-[transform] duration-300 ease-out',
        'bg-[linear-gradient(120deg,var(--color-primary),var(--color-accent))] shadow-[0_4px_14px_-4px_color-mix(in_oklch,var(--color-primary)_60%,transparent)]',
        theme.mode === 'light' && 'translate-x-0',
        theme.mode === 'auto' && 'translate-x-[100%]',
        theme.mode === 'dark' && 'translate-x-[200%]',
      ]"
    />
    <button
      v-for="opt in OPTIONS"
      :key="opt.value"
      type="button"
      role="radio"
      :aria-checked="theme.mode === opt.value"
      :aria-label="opt.label"
      :class="[
        'relative z-10 grid h-7 w-8 place-items-center rounded-full transition-colors',
        theme.mode === opt.value
          ? 'text-primary-content'
          : 'text-base-content/55 hover:text-base-content',
      ]"
      @click="theme.setMode(opt.value)"
    >
      <svg
        viewBox="0 0 24 24"
        class="h-3.5 w-3.5"
        fill="none"
        stroke="currentColor"
        stroke-width="1.8"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path :d="opt.icon" />
      </svg>
    </button>
  </div>
</template>
