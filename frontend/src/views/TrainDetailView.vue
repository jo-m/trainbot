<script setup lang="ts">
import { inject, computed, defineProps } from 'vue'
import { dbKey, getTrain } from '@/lib/db'
import { getBlobURL } from '@/lib/paths'
import type SqlJs from 'sql.js'
import { useRouter } from 'vue-router'
import RelativeTime from '@/components/RelativeTime.vue'
import FavoriteIcon from '@/components/FavoriteIcon.vue'

const props = defineProps<{
  id: string
}>()
const router = useRouter()
const db = inject(dbKey) as SqlJs.Database

const id = computed(() => {
  return parseInt(props.id)
})
const train = computed(() => {
  const t = getTrain(db, id.value)
  if (t === undefined) {
    router.push({ name: 'notFound' })
  }
  return t
})
</script>

<template>
  <v-card v-if="train !== undefined">
    <v-card-item>
      <v-card-title>Train #{{ train.id }} <FavoriteIcon :id="train.id" /></v-card-title>
    </v-card-item>

    <v-card-text>
      <v-table>
        <thead>
          <tr>
            <th class="text-left">Name</th>
            <th class="text-left">Value</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>ID</td>
            <td>{{ train.id }}</td>
          </tr>
          <tr>
            <td>Start timestamp</td>
            <td>{{ train.start_ts.toSQL() }} (<RelativeTime :ts="train.start_ts" />)</td>
          </tr>
          <tr>
            <td>End timestamp</td>
            <td>{{ train.end_ts.toSQL() }}</td>
          </tr>
          <tr v-if="train.uploaded_at">
            <td>Upload timestamp</td>
            <td>{{ train.uploaded_at.toSQL() }}</td>
          </tr>
          <tr>
            <td>Direction</td>
            <td>{{ train.speed_px_s > 0 ? 'Right' : 'Left' }}</td>
          </tr>
          <tr>
            <td>Length [m]</td>
            <td>{{ Math.round(train.length_px / train.px_per_m) }}</td>
          </tr>
          <tr>
            <td>Speed [km/h]</td>
            <td>{{ Math.abs(Math.round((train.speed_px_s / train.px_per_m) * 3.6)) }}</td>
          </tr>
          <tr>
            <td>Acceleration [m/s^2]</td>
            <td>
              {{
                Math.round(
                  (train.accel_px_s_2 / train.px_per_m) * Math.sign(train.speed_px_s) * 10
                ) / 10
              }}
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card-text>

    <v-divider class="mx-4 mb-1"></v-divider>
    <v-card-title>Image</v-card-title>

    <a :href="getBlobURL(train?.image_file_path)" target="_blank">
      <v-img cover :src="getBlobURL(train?.image_file_path)"></v-img>
    </a>

    <v-divider class="mx-4 mb-1"></v-divider>
    <v-card-title>GIF</v-card-title>

    <a :href="getBlobURL(train?.gif_file_path)" target="_blank">
      <v-img width="10em" :src="getBlobURL(train?.gif_file_path)"></v-img>
    </a>
  </v-card>
</template>
