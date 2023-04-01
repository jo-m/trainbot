<script setup lang="ts">
import TrainList from '@/components/TrainList.vue'
import { ref, onMounted, onUnmounted } from 'vue'
import { loadDB, getTrains, type Train as TrainType } from '@/lib/db'

const db = await loadDB()

const loadSize = 20
const trains = ref<TrainType[]>([])
const filteredCount = ref<number>(0)
const totalCount = ref<number>(0)
const orderBy = ref<string>('start_ts DESC')
const filters = ref<{ [key: string]: string }>({ _: '1=1' })
const atEnd = ref<boolean>(false)

const scroller = ref<HTMLDivElement | null>(null)

function buildFilter(): string {
  return Object.values(filters.value).join(' AND ')
}

function loadMore() {
  const result = getTrains(db, loadSize, trains.value.length, buildFilter(), orderBy.value)
  if (result.trains.length < loadSize) {
    atEnd.value = true
  }

  trains.value.push(...result.trains)
  filteredCount.value = result.filteredCount
  totalCount.value = result.totalCount
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
  loadMore()

  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <Teleport to="#app-bar-teleport">
    <v-btn variant="text" icon="mdi-filter"></v-btn>
    {{ trains.length }} / {{ filteredCount }} / {{ totalCount }}
  </Teleport>

  <v-card>
    <div ref="scroller">
      <TrainList :trains="trains" />
    </div>
  </v-card>
</template>
