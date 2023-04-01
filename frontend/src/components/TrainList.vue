<script setup lang="ts">
import type { Train as TrainType } from '@/lib/db'
import Train from '@/components/Train.vue'

defineProps<{
  trains: TrainType[]
}>()
</script>

<template>
  <div>
    <Train v-for="train in trains" v-bind:key="train.id" :train="train" />
  </div>
</template>
