<script setup lang="ts">
import type { DateTime, ToRelativeOptions } from 'luxon'
import { getCurrentInstance, onMounted, onUnmounted } from 'vue'

let timer: ReturnType<typeof setInterval> | undefined = undefined

defineProps<{
  ts: DateTime
  opts?: ToRelativeOptions
}>()

onMounted(async () => {
  const instance = getCurrentInstance()
  timer = setInterval(() => instance?.proxy?.$forceUpdate(), 1000)
})

onUnmounted(() => {
  clearInterval(timer)
})
</script>

<template>
  <span>{{ ts.toRelative(opts) }}</span>
</template>
