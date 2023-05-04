<script setup lang="ts">
import TrainGrid from '@/components/TrainGrid.vue'
import CopyFavoritesDialog from '@/components/CopyFavoritesDialog.vue'
import { ref, inject, watch } from 'vue'
import { dbKey, getTrains, type Train as TrainType } from '@/lib/db'
import useFavoritesStore from '@/lib/favorites'
import type SqlJs from 'sql.js'

const favs = useFavoritesStore()

const db = inject(dbKey) as SqlJs.Database

const trains = ref<TrainType[] | null>(null)

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
  <!-- App bar -->
  <Teleport to="#app-bar-teleport">
    <CopyFavoritesDialog />
  </Teleport>

  <template v-if="trains !== null">
    <TrainGrid :trains="trains" />

    <v-card v-if="trains?.length == 0" min-height="50%">
      <v-card-item>
        <v-card-title>Nothing to see here</v-card-title>
      </v-card-item>

      <v-card-text>
        You have no favorites yet. Click on the
        <v-icon icon="mdi-star" color="#FFC107" style="padding-bottom: 2px" /> icon to save trains
        to favorites <v-icon icon="mdi-star-outline" color="#BDBDBD" style="padding-bottom: 2px" />.
      </v-card-text>
    </v-card></template
  >
</template>
