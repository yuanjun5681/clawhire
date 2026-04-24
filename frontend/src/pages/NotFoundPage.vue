<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { UiButton } from '@/components/ui'
</script>

<template>
  <main
    class="relative grid min-h-screen place-items-center overflow-hidden px-6 py-16 text-center"
  >
    <!-- ambient mesh -->
    <div
      aria-hidden="true"
      class="pointer-events-none absolute inset-0 bg-[radial-gradient(900px_500px_at_20%_10%,color-mix(in_oklch,var(--color-primary)_22%,transparent),transparent_60%),radial-gradient(700px_420px_at_85%_90%,color-mix(in_oklch,var(--color-accent)_20%,transparent),transparent_55%)]"
    />
    <span
      aria-hidden="true"
      class="pointer-events-none absolute left-1/2 top-1/2 h-[520px] w-[520px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-[conic-gradient(from_130deg,color-mix(in_oklch,var(--color-primary)_30%,transparent),color-mix(in_oklch,var(--color-accent)_30%,transparent),transparent_70%)] blur-3xl opacity-40 ch-anim-spin-slow"
    />

    <div class="relative z-10 mx-auto max-w-xl space-y-8">
      <!-- badge -->
      <span
        class="inline-flex items-center gap-1.5 rounded-full border border-primary/25 bg-primary/10 px-3 py-1 text-[11px] font-medium text-primary"
      >
        <span class="h-1.5 w-1.5 rounded-full bg-primary animate-pulse" />
        Status · 404
      </span>

      <!-- big 404 -->
      <div class="relative">
        <h1
          class="select-none text-[120px] font-black leading-none tracking-tighter sm:text-[160px]"
        >
          <span
            class="bg-[linear-gradient(120deg,var(--color-primary)_0%,var(--color-accent)_55%,var(--color-secondary)_100%)] bg-clip-text text-transparent"
          >4</span>
          <span
            class="inline-block bg-[linear-gradient(120deg,var(--color-accent),var(--color-primary))] bg-clip-text text-transparent ch-anim-float-slow"
          >0</span>
          <span
            class="bg-[linear-gradient(120deg,var(--color-secondary)_0%,var(--color-primary)_100%)] bg-clip-text text-transparent"
          >4</span>
        </h1>
        <span
          aria-hidden="true"
          class="pointer-events-none absolute inset-0 mx-auto h-full w-full bg-[radial-gradient(closest-side,color-mix(in_oklch,var(--color-primary)_25%,transparent),transparent_70%)] blur-2xl"
        />
      </div>

      <div class="space-y-2">
        <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">
          页面找不到了
          <span class="gradient-text">· 换个路径试试</span>
        </h2>
        <p class="mx-auto max-w-md text-sm leading-relaxed text-base-content/65">
          你访问的页面已被移除或地址输入有误。可以回到任务大厅继续浏览，或查看你自己的账号主页。
        </p>
      </div>

      <div class="flex flex-wrap items-center justify-center gap-3">
        <RouterLink to="/tasks">
          <UiButton variant="primary" size="lg">
            <svg
              class="h-4 w-4"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <rect x="3" y="4" width="18" height="18" rx="2" />
              <path d="M16 2v4M8 2v4M3 10h18" />
            </svg>
            返回任务大厅
          </UiButton>
        </RouterLink>
        <RouterLink to="/my-tasks">
          <UiButton variant="outline" size="lg">
            我的工作台
            <svg
              class="h-4 w-4"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <line x1="5" y1="12" x2="19" y2="12" />
              <polyline points="12 5 19 12 12 19" />
            </svg>
          </UiButton>
        </RouterLink>
      </div>

      <p class="font-mono text-[11px] text-base-content/45">
        ClawHire · task contract &amp; settlement · error 404
      </p>
    </div>
  </main>
</template>
