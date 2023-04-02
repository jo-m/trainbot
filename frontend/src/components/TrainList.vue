<script setup lang="ts">
import type { Train as TrainType } from '@/lib/db'
import TrainListItem from '@/components/TrainListItem.vue'

defineProps<{
  trains: TrainType[]
  allDataLoaded: boolean
}>()
</script>

<template>
  <v-container class="pa-2">
    <v-row v-for="train in trains" v-bind:key="train.id" no-gutters>
      <v-col cols="12" class="pointer">
        <router-link
          :to="{ name: 'trainDetail', params: { id: train.id } }"
          style="text-decoration: none; color: inherit"
        >
          <TrainListItem :train="train" />
        </router-link>

        <v-divider></v-divider>
      </v-col>
    </v-row>

    <v-row no-gutters>
      <v-col cols="12">
        <v-row class="pa-0" no-gutters align="center">
          <v-col cols="12">
            <v-sheet class="pa-2" v-if="allDataLoaded">
              <v-icon icon="mdi-arrow-collapse-down"></v-icon>
              End of list ({{ trains.length }} trains).
            </v-sheet>
            <v-sheet class="pa-2" v-else>
              <v-progress-circular indeterminate></v-progress-circular> Loading...
            </v-sheet>
          </v-col>
        </v-row>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped>
.pointer {
  cursor: pointer;
}
</style>
