<script setup lang="ts">
import type { Train as TrainType } from '@/lib/db'
import TrainListItem from '@/components/TrainListItem.vue'

defineProps<{
  trains: TrainType[]
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
  </v-container>
</template>

<style scoped>
.pointer {
  cursor: pointer;
}
</style>
