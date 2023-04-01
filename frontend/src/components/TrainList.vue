<script setup lang="ts">
import type { Train as TrainType } from '@/lib/db'
import Train from '@/components/Train.vue'

defineProps<{
  trains: TrainType[]
}>()
</script>

<template>
  <v-container class="pa-2">
    <v-row v-for="train in trains" v-bind:key="train.id" no-gutters>
      <v-col cols="12">
        <Train :train="train" />
        <v-divider></v-divider>
      </v-col>
    </v-row>
  </v-container>
</template>
