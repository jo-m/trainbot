<script setup lang="ts">
import type { Train } from '@/lib/db'

defineProps<{
  train: Train
}>()

const blobsBaseURL = import.meta.env.VITE_BLOBS_URL
function getURL(blobName: string): string {
  return blobsBaseURL.trimRight('/') + '/' + blobName
}
</script>

<template>
  <v-card>
    {{ train.id }}
    {{ train.start_ts.toRelative() }} ({{ train.start_ts.toString() }})<br />
    speed: {{ Math.abs(Math.round((train.speed_px_s / train.px_per_m) * 3.6)) }}km/h<br />
    length: {{ Math.round(train.length_px / train.px_per_m) }}m<br />
    direction: {{ train.speed_px_s > 0 ? 'right' : 'left' }}
    <div
      class="train-preview"
      :style="`background-image: url(${getURL(train.image_file_path)})`"
    ></div>
    <hr />
  </v-card>
</template>

<style scoped>
div.train-preview {
  height: 4em;
  background-size: auto 100%;
  background-repeat: no-repeat;
}
</style>
