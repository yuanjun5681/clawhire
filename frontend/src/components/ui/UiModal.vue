<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'

const props = defineProps<{
  open: boolean
  title?: string
  description?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  closeOnOverlay?: boolean
}>()

const emit = defineEmits<{ close: [] }>()

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open) emit('close')
}

watch(
  () => props.open,
  (open) => {
    if (typeof document === 'undefined') return
    document.body.style.overflow = open ? 'hidden' : ''
    if (open) document.addEventListener('keydown', onKey)
    else document.removeEventListener('keydown', onKey)
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  if (typeof document !== 'undefined') {
    document.body.style.overflow = ''
    document.removeEventListener('keydown', onKey)
  }
})

function onOverlay() {
  if (props.closeOnOverlay !== false) emit('close')
}

function sizeClass() {
  switch (props.size) {
    case 'sm':
      return 'max-w-md'
    case 'lg':
      return 'max-w-3xl'
    case 'xl':
      return 'max-w-5xl'
    default:
      return 'max-w-2xl'
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      leave-active-class="transition duration-150 ease-in"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-[60] overflow-y-auto"
        role="dialog"
        aria-modal="true"
      >
        <div
          class="fixed inset-0 bg-base-300/40 backdrop-blur-md"
          aria-hidden="true"
          @click="onOverlay"
        />

        <div class="relative flex min-h-full items-start justify-center p-4 sm:p-8">
          <Transition
            enter-active-class="transition duration-300 ease-out"
            enter-from-class="opacity-0 translate-y-3 scale-[0.97]"
            leave-active-class="transition duration-200 ease-in"
            leave-to-class="opacity-0 translate-y-3 scale-[0.97]"
            appear
          >
            <div
              v-if="open"
              :class="[
                'relative w-full overflow-hidden rounded-box surface',
                'shadow-[0_30px_60px_-20px_color-mix(in_oklch,var(--color-primary)_35%,transparent)]',
                sizeClass(),
              ]"
            >
              <header
                v-if="title || $slots.header"
                class="relative flex items-start justify-between gap-4 border-b border-base-300/60 px-6 py-4"
              >
                <div class="min-w-0 flex-1">
                  <slot name="header">
                    <h2
                      class="text-[15px] font-semibold tracking-tight text-base-content"
                    >
                      {{ title }}
                    </h2>
                    <p
                      v-if="description"
                      class="mt-0.5 text-xs text-base-content/60"
                    >
                      {{ description }}
                    </p>
                  </slot>
                </div>
                <button
                  type="button"
                  class="grid h-8 w-8 shrink-0 place-items-center rounded-full text-base-content/55 transition hover:bg-base-200 hover:text-base-content"
                  aria-label="关闭"
                  @click="$emit('close')"
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
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                </button>
              </header>

              <div class="px-6 py-5">
                <slot />
              </div>

              <footer
                v-if="$slots.footer"
                class="flex items-center justify-end gap-2 border-t border-base-300/60 bg-base-200/40 px-6 py-3.5"
              >
                <slot name="footer" />
              </footer>
            </div>
          </Transition>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
