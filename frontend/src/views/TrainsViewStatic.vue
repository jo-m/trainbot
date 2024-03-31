<script setup lang="ts">
import { ref, inject } from 'vue'
import { dbKey, getTrains, type Train as TrainType } from '@/lib/db'
import type SqlJs from 'sql.js'
import { getBlobURL, imgFileName } from '@/lib/paths'

const db = inject(dbKey) as SqlJs.Database
const trains = ref<TrainType[] | null>(getTrains(db, 20).trains)
</script>

<template>
  <!-- List -->
  <template v-if="trains !== null">
    <div>
      <v-container fluid>
        <v-row v-for="train in trains" v-bind:key="train.id" no-gutters>
          <v-col cols="12" class="pointer">
            <v-row class="pa-0" no-gutters>
              <v-col cols="2">
                <v-sheet class="pa-2">
                  <v-chip
                    density="comfortable"
                    label
                    :prepend-icon="
                      train.speed_px_s > 0 ? 'mdi-arrow-right-bold' : 'mdi-arrow-left-bold'
                    "
                  >
                    {{ Math.abs(Math.round((train.speed_px_s / train.px_per_m) * 3.6)) }} km/h
                  </v-chip>
                  &nbsp;
                  <v-chip density="comfortable" label
                    >{{ Math.round(train.length_px / train.px_per_m) }} m</v-chip
                  >
                </v-sheet>
              </v-col>

              <v-col cols="10">
                <v-sheet
                  class="ma-1 train-preview"
                  :style="`background-image: url(${getBlobURL(
                    imgFileName(train.start_ts)
                  )}); background-position-x: ${train.speed_px_s > 0 ? 'right' : 'left'}`"
                >
                </v-sheet>
              </v-col>
            </v-row>

            <v-divider></v-divider>
          </v-col>
        </v-row>
      </v-container>
    </div>
  </template>
</template>

<style scoped>
div.v-sheet {
  background-color: inherit;
}
div.train-preview {
  background-color: #eee;
  height: 64px;
  background-size: auto 100%;
  background-repeat: no-repeat;
}
</style>
