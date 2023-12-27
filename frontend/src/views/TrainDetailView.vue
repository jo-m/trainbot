<script setup lang="ts">
import { inject, computed } from 'vue'
import { dbKey, getTrain, queryOne } from '@/lib/db'
import { getBlobURL, imgFileName, gifFileName } from '@/lib/paths'
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

const nextId = computed(
  () =>
    queryOne(db, `SELECT id FROM trains_v2 WHERE id > ${id.value} ORDER BY id ASC LIMIT 1`) as
      | number
      | undefined
)
const prevId = computed(
  () =>
    queryOne(db, `SELECT id FROM trains_v2 WHERE id < ${id.value} ORDER BY id DESC LIMIT 1`) as
      | number
      | undefined
)
</script>

<template>
  <!-- App bar -->
  <Teleport to="#app-bar-teleport">
    <v-btn
      :disabled="prevId === undefined"
      variant="text"
      icon="mdi-arrow-left"
      :to="{ name: 'trainDetail', params: { id: prevId } }"
      aria-label="Previous"
    ></v-btn>

    <v-btn
      :disabled="nextId === undefined"
      variant="text"
      icon="mdi-arrow-right"
      :to="{ name: 'trainDetail', params: { id: nextId } }"
      aria-label="Next"
    ></v-btn>
  </Teleport>

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

    <a :href="getBlobURL(imgFileName(train.start_ts))" target="_blank">
      <v-img cover :src="getBlobURL(imgFileName(train.start_ts))"></v-img>
    </a>

    <v-divider class="mx-4 mb-1"></v-divider>
    <v-card-title>GIF</v-card-title>

    <a :href="getBlobURL(gifFileName(train.start_ts))" target="_blank">
      <v-img width="10em" :src="getBlobURL(gifFileName(train.start_ts))"></v-img>
    </a>
  </v-card>
</template>
