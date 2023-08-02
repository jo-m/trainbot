<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  data: number[][]
  labels?: { [key: number]: string }
  widthPx: number
  color: string
}>()

const maxVal = computed(() => {
  return props.data.reduce((prev, cur) => Math.max(prev, cur[1]), 0)
})
</script>

<template>
  <div
    v-for="[key, val] in data"
    :key="key"
    :style="{
      color: 'white',
      backgroundColor: color,
      width: `${(val / maxVal) * widthPx}px`,
      margin: '4px',
      fontFamily: 'monospace',
      overflow: 'clip',
      whiteSpace: 'nowrap'
    }"
  >
    &nbsp;{{ labels ? labels[key] : key }}: {{ val }}
  </div>
</template>
