<script setup lang="ts">
import type { Train } from '@/lib/db'
import { getBlobURL } from '@/lib/paths'
import { DateTime } from 'luxon'
import { useDisplay } from 'vuetify'
import RelativeTime from '@/components/RelativeTime.vue'

defineProps<{
  train: Train
}>()

const { mdAndUp } = useDisplay()
</script>

<template>
  <v-row class="pa-0" no-gutters align="center">
    <v-col cols="5" sm="3" md="2" lg="2">
      <v-tooltip :text="train.start_ts.toSQL()" location="top">
        <template v-slot:activator="{ props }">
          <v-sheet v-bind="props" class="pa-2">
            <RelativeTime :ts="train.start_ts" />
            <div class="text-caption" v-if="mdAndUp">
              {{ train.start_ts.toLocaleString(DateTime.DATETIME_SHORT) }}
            </div>
          </v-sheet>
        </template>
      </v-tooltip>
    </v-col>

    <v-col cols="7" sm="9" md="3" lg="2">
      <v-sheet class="pa-2">
        <v-chip
          density="comfortable"
          label
          :prepend-icon="train.speed_px_s > 0 ? 'mdi-arrow-right-bold' : 'mdi-arrow-left-bold'"
          >{{ Math.abs(Math.round((train.speed_px_s / train.px_per_m) * 3.6)) }} km/h</v-chip
        >
        &nbsp;
        <v-chip density="comfortable" label
          >{{ Math.round(train.length_px / train.px_per_m) }} m</v-chip
        >
      </v-sheet>
    </v-col>

    <v-col cols="12" sm="12" md="6" lg="8">
      <v-sheet
        class="ma-1 train-preview"
        :style="`background-image: url(${getBlobURL(
          train.image_file_path
        )}); background-position-x: ${train.speed_px_s > 0 ? 'right' : 'left'}`"
      >
      </v-sheet>
    </v-col>
  </v-row>
</template>

<style scoped>
div.train-preview {
  background-color: #eee;
  height: 4em;
  background-size: auto 100%;
  background-repeat: no-repeat;
}
div.train-preview.v-theme--dark {
  background-color: #222;
}
</style>
