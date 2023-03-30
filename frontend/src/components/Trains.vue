<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { loadDB, getTrains, type Train as TrainType } from '@/lib/db'
import Train from '@/components/Train.vue'

const trains = ref<TrainType[]>([])

onMounted(async () => {
  const db = await loadDB()
  trains.value = getTrains(db, 80, 0)
})
</script>

<template>
  <div>
    <Train v-for="train in trains" v-bind:key="train.id" :train="train" />
  </div>
</template>
