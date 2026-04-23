<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <section class="space-y-4">
    <header class="space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight text-base-content">
        {{ title }}
      </h1>
      <p v-if="description" class="text-sm text-base-content/60">
        {{ description }}
      </p>
    </header>

    <div
      class="rounded-xl border border-dashed border-base-300 bg-base-100 p-10 text-center text-sm text-base-content/50"
    >
      <slot>此页面结构已就位，业务内容将在后续步骤中填充。</slot>
    </div>
  </section>
</template>
