<script setup lang="ts">
import { computed } from 'vue'
import { useToastStore, type ToastItem, type ToastKind } from '@/stores/toast'

const toast = useToastStore()

const items = computed(() => toast.items)

const ICONS: Record<ToastKind, string> = {
  success:
    'M22 11.08V12a10 10 0 1 1-5.93-9.14 M22 4L12 14.01l-3-3',
  error:
    'M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm0 5v5 M12 17h.01',
  warning:
    'M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z M12 9v4 M12 17h.01',
  info:
    'M12 2a10 10 0 1 1-10 10A10 10 0 0 1 12 2zm0 10v5 M12 7h.01',
}

function toneOf(k: ToastKind) {
  switch (k) {
    case 'success':
      return {
        ring: 'ring-success/30',
        glow: 'bg-success/15',
        icon: 'text-success bg-success/15',
      }
    case 'error':
      return {
        ring: 'ring-error/30',
        glow: 'bg-error/15',
        icon: 'text-error bg-error/15',
      }
    case 'warning':
      return {
        ring: 'ring-warning/40',
        glow: 'bg-warning/20',
        icon: 'text-[color-mix(in_oklch,var(--color-warning)_40%,var(--color-base-content))] bg-warning/25',
      }
    default:
      return {
        ring: 'ring-primary/25',
        glow: 'bg-primary/15',
        icon: 'text-primary bg-primary/15',
      }
  }
}

function iconPath(item: ToastItem) {
  return ICONS[item.kind] ?? ICONS.info
}
</script>

<template>
  <Teleport to="body">
    <div
      aria-live="polite"
      aria-atomic="true"
      class="pointer-events-none fixed inset-x-0 top-4 z-80 flex flex-col items-center gap-2.5 px-4 sm:inset-x-auto sm:right-5 sm:top-5 sm:items-end"
    >
      <TransitionGroup name="ch-toast" tag="div" class="flex flex-col items-center gap-2.5 sm:items-end">
        <div
          v-for="(item, idx) in items"
          :key="item.id"
          :class="[
            'ch-toast pointer-events-auto relative w-full max-w-sm overflow-hidden rounded-box',
            'glass ring-1',
            toneOf(item.kind).ring,
            'shadow-[0_20px_50px_-18px_color-mix(in_oklch,var(--color-base-content)_30%,transparent)]',
          ]"
          :style="{
            transform: `translateY(${idx * 2}px) scale(${1 - idx * 0.015})`,
            zIndex: 80 - idx,
          }"
        >
          <span
            aria-hidden="true"
            :class="[
              'pointer-events-none absolute -left-10 -top-10 h-32 w-32 rounded-full blur-2xl opacity-70',
              toneOf(item.kind).glow,
            ]"
          />

          <div class="relative flex items-start gap-3 px-4 py-3.5">
            <span
              :class="[
                'grid h-8 w-8 shrink-0 place-items-center rounded-full',
                toneOf(item.kind).icon,
              ]"
            >
              <svg
                class="h-4 w-4"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path :d="iconPath(item)" />
              </svg>
            </span>

            <div class="min-w-0 flex-1 space-y-0.5">
              <p v-if="item.title" class="text-sm font-semibold text-base-content">
                {{ item.title }}
              </p>
              <p class="text-[13px] leading-relaxed text-base-content/80 wrap-break-word">
                {{ item.message }}
              </p>
            </div>

            <button
              type="button"
              class="grid h-7 w-7 shrink-0 place-items-center rounded-full text-base-content/45 transition hover:bg-base-200 hover:text-base-content"
              aria-label="关闭"
              @click="toast.dismiss(item.id)"
            >
              <svg
                class="h-3.5 w-3.5"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>

          <span
            v-if="item.duration > 0"
            aria-hidden="true"
            class="absolute bottom-0 left-0 h-[2px] bg-[linear-gradient(90deg,var(--color-primary),var(--color-accent))]"
            :style="{
              animation: `ch-toast-progress ${item.duration}ms linear forwards`,
              width: '100%',
            }"
          />
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style>
@keyframes ch-toast-progress {
  from { transform: translateX(0); }
  to { transform: translateX(-100%); }
}

.ch-toast-enter-active {
  animation: ch-toast-in 360ms cubic-bezier(0.22, 1, 0.36, 1);
}
.ch-toast-leave-active {
  animation: ch-toast-out 220ms ease-in forwards;
}
.ch-toast-move {
  transition: transform 320ms cubic-bezier(0.22, 1, 0.36, 1);
}
</style>
