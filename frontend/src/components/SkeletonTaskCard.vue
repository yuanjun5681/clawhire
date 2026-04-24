<script setup lang="ts"></script>

<template>
  <div
    class="relative overflow-hidden rounded-box border border-base-300/70 bg-base-100 p-5"
  >
    <span
      aria-hidden="true"
      class="pointer-events-none absolute inset-0 -translate-x-full bg-linear-to-r from-transparent via-base-200/70 to-transparent animate-[ch-shimmer_1.4s_infinite]"
    />
    <div class="flex items-start justify-between gap-3">
      <div class="h-4 w-2/3 rounded-full bg-base-200" />
      <div class="h-5 w-16 rounded-full bg-base-200" />
    </div>
    <div class="mt-3 flex items-center gap-2">
      <div class="h-4 w-14 rounded-full bg-base-200" />
      <div class="h-3 w-24 rounded-full bg-base-200" />
    </div>
    <div class="mt-4 flex items-center justify-between border-t border-base-300/60 pt-3">
      <div class="flex items-center gap-4">
        <div class="h-3 w-20 rounded-full bg-base-200" />
        <div class="h-3 w-24 rounded-full bg-base-200" />
      </div>
      <div class="h-3 w-20 rounded-full bg-base-200" />
    </div>
  </div>
</template>

<style scoped>
@keyframes ch-shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}
.animate-\[ch-shimmer_1\.4s_infinite\] {
  animation: ch-shimmer 1.4s infinite;
}
</style>
