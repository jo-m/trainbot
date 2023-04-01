<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { loadDB, getTrains, type Train as TrainType } from '@/lib/db'
import Train from '@/components/Train.vue'

const pageSize = 20
const db = await loadDB()

const trains = ref<TrainType[]>([])
const atEnd = ref<boolean>(false)

const scroller = ref<HTMLDivElement | null>(null)

function loadMore() {
  const lastId = trains.value[trains.value.length - 1].id
  const more = getTrains(db, pageSize, lastId)
  if (more.length < pageSize) {
    atEnd.value = true
  }
  trains.value.push(...more)
}

function handleScroll() {
  let element = scroller.value
  if (element === null) {
    return
  }
  if (element.getBoundingClientRect().bottom <= window.innerHeight) {
    loadMore()
  }
}

onMounted(async () => {
  trains.value = getTrains(db, pageSize)

  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <div ref="scroller">
    <Train v-for="train in trains" v-bind:key="train.id" :train="train" />
    {{ atEnd ? 'No more trains to load.' : '' }}
  </div>
</template>
