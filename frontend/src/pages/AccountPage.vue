<script setup lang="ts">
import { useRoute } from 'vue-router'
import PagePlaceholder from '@/components/PagePlaceholder.vue'

const route = useRoute()
const accountId = route.params.accountId as string
</script>

<template>
  <PagePlaceholder
    title="账号主页"
    :description="`账号 ID：${accountId}`"
  />
</template>
