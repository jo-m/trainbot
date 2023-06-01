<script setup lang="ts">
import { ref, inject, watch } from 'vue'
import { dbKey, getTrains, type Train as TrainType } from '@/lib/db'
import useFavoritesStore from '@/lib/favorites'
import type SqlJs from 'sql.js'

const favs = useFavoritesStore()
const trains = ref<TrainType[] | null>(null)
const db = inject(dbKey) as SqlJs.Database

watch(
  favs.favorites,
  () => {
    const ids = Array.from(favs.favorites).join(',')
    const result = getTrains(db, -1, 0, { where: { favs: `id IN (${ids})` } })
    trains.value = result.trains
  },
  { immediate: true }
)
</script>

<template>
  sqlite3 data/db.sqlite3
  <br />

  DELETE FROM trains
  <br />
  WHERE id IN ({{ Array.from(favs.favorites).join(', ') }});
</template>
