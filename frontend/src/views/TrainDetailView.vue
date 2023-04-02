<script setup lang="ts">
import { inject } from 'vue'
import { dbKey, getTrain } from '@/lib/db'
import { getBlobURL } from '@/lib/paths'
import type SqlJs from 'sql.js'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const db = inject(dbKey) as SqlJs.Database
const train = getTrain(db, route.params.id as any as number)

if (train === undefined) {
  router.push({ name: 'notFound' })
}
</script>

<template>
  <div v-if="train !== undefined">
    Timestamp: {{ train.start_ts.toSQL() }}
    <a :href="getBlobURL(train?.image_file_path)" target="_blank">
      <img :src="getBlobURL(train?.image_file_path)" style="width: 100%" />
    </a>

    <a :href="getBlobURL(train?.gif_file_path)" target="_blank">
      <img :src="getBlobURL(train?.gif_file_path)" />
    </a>
  </div>
</template>
