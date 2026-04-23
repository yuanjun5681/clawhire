<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main
    class="grid min-h-screen place-items-center bg-base-200 px-6 text-center"
  >
    <div class="space-y-3">
      <p class="text-xs uppercase tracking-wider text-base-content/50">404</p>
      <h1 class="text-3xl font-semibold text-base-content">页面不存在</h1>
      <p class="text-sm text-base-content/60">
        你访问的页面已被移除或地址错误。
      </p>
      <RouterLink
        to="/tasks"
        class="inline-flex items-center rounded-md bg-primary px-4 py-2 text-sm text-primary-content hover:bg-primary/90"
      >
        返回任务大厅
      </RouterLink>
    </div>
  </main>
</template>
