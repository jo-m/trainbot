<script setup lang="ts">
import TrainList from '@/components/TrainList.vue'
import { ref, onMounted, onUnmounted } from 'vue'
import { loadDB, getTrains, type Train as TrainType } from '@/lib/db'

const db = await loadDB()

const trains = ref<TrainType[]>([])
const atEnd = ref<boolean>(false)

const loadSize = 20

const scroller = ref<HTMLDivElement | null>(null)

function loadMore() {
  const more = getTrains(db, loadSize, trains.value.length).trains
  if (more.length < loadSize) {
    atEnd.value = true
  }
  trains.value.push(...more)
}

function handleScroll() {
  let element = scroller.value
  if (element === null) {
    return
  }
  if (element.getBoundingClientRect().bottom - 10 <= window.innerHeight) {
    loadMore()
  }
}

onMounted(async () => {
  trains.value = getTrains(db, loadSize).trains

  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <Teleport to="#app-bar-teleport">
    <v-btn variant="text" icon="mdi-filter"></v-btn>
  </Teleport>

  <v-card>
    <div ref="scroller">
      <TrainList :trains="trains" />
    </div>
  </v-card>
</template>
