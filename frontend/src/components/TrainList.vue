<script setup lang="ts">
import type { Train as TrainType } from '@/lib/db'
import Train from '@/components/Train.vue'

defineProps<{
  trains: TrainType[]
  noMoreData: boolean
}>()

defineEmits<{
  (e: 'trainSelected', id: number): void
}>()
</script>

<template>
  <v-container class="pa-2">
    <v-row v-for="train in trains" v-bind:key="train.id" no-gutters>
      <v-col cols="12" @click="$emit('trainSelected', train.id)" class="pointer">
        <Train :train="train" />
        <v-divider></v-divider>
      </v-col>
    </v-row>

    <v-row no-gutters>
      <v-col cols="12">
        <v-row class="pa-0" no-gutters align="center">
          <v-col cols="12">
            <v-sheet class="pa-2" v-if="noMoreData">
              <v-icon icon="mdi-arrow-collapse-down"></v-icon> End of list ({{ trains.length }}
              trains).
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
