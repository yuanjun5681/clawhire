<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { RouterView } from 'vue-router'
import { ToastContainer } from '@/components/ui'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()
let unbind: (() => void) | null = null

onMounted(() => {
  unbind = theme.bindSystemListener()
})

onBeforeUnmount(() => {
  unbind?.()
})
</script>

<template>
  <RouterView />
  <ToastContainer />
</template>
