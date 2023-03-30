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
  <div>
    ts: {{ train.start_ts }}<br />
    speed km/h: {{ Math.round((train.speed_px_s / train.px_per_m) * 3.6) }} +
    {{ Math.round(train.accel_px_s_2 / train.px_per_m) }} m^2/s<br />
    length m: {{ Math.round(train.length_px / train.px_per_m) }}<br />
    <img :src="getURL(train.image_file_path)" height="30" />
  </div>
</template>
