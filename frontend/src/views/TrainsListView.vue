<script setup lang="ts">
import TrainList from '@/components/TrainList.vue'
import { ref, onMounted, onUnmounted, inject } from 'vue'
import { dbKey, getTrains, type Train as TrainType, type Filter } from '@/lib/db'
import type SqlJs from 'sql.js'
import { DateTime } from 'luxon'

const db = inject(dbKey) as SqlJs.Database

// How many trains to load at a time.
const pageSize = 20
// Currently loaded data.
const trains = ref<TrainType[] | null>(null)
const filteredCount = ref<number | null>(null)
const totalCount = ref<number | null>(null)
// If we have reached the end of pagination.
const allDataLoaded = ref<boolean>(false)
// Currently active filter.
// When this changes, data must be reset.
const filter = ref<Filter>({})

const scroller = ref<HTMLDivElement | null>(null)

const showFilterDialog = ref<boolean>(false)

function updateFilter(newFilter: Filter, reset: boolean = false) {
  if (reset) {
    filter.value = newFilter
  } else {
    const copy = filter.value !== undefined ? JSON.parse(JSON.stringify(filter.value)) : {}
    if (newFilter.orderBy !== undefined) {
      copy.orderBy = newFilter.orderBy
    }
    if (copy.where === undefined) {
      copy.where = {}
    }
    if (newFilter.where !== undefined) {
      Object.assign(copy.where, newFilter.where)
    }

    filter.value = copy
  }

  trains.value = null
  filteredCount.value = null
  totalCount.value = null
  allDataLoaded.value = false

  showFilterDialog.value = false

  loadNextData()
}

function loadNextData() {
  if (allDataLoaded.value) return

  const currentLen = trains.value?.length || 0
  const result = getTrains(db, pageSize, currentLen, filter.value)
  if (result.trains.length < pageSize) {
    allDataLoaded.value = true
  }

  if (trains.value === null) {
    trains.value = result.trains
  } else {
    trains.value.push(...result.trains)
  }
  filteredCount.value = result.filteredCount
  totalCount.value = result.totalCount
}

function handleScroll() {
  let element = scroller.value
  if (element === null) {
    return
  }
  if (element.getBoundingClientRect().bottom - 5 <= window.innerHeight) {
    loadNextData()
  }
}

onMounted(async () => {
  loadNextData()

  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <!-- App bar -->
  <Teleport to="#app-bar-teleport">
    <v-btn
      variant="text"
      icon="mdi-filter"
      @click="showFilterDialog = true"
      :active="Object.keys(filter).length > 0"
    ></v-btn>
  </Teleport>

  <!-- List -->
  <div ref="scroller" v-if="trains !== null">
    <TrainList :trains="trains" :allDataLoaded="allDataLoaded" />
  </div>

  <!-- Filter -->
  <v-dialog
    v-model="showFilterDialog"
    fullscreen
    :scrim="false"
    transition="dialog-bottom-transition"
  >
    <v-card>
      <v-toolbar color="primary">
        <v-btn icon="" @click="showFilterDialog = false">
          <v-icon>mdi-close</v-icon>
        </v-btn>
        <v-toolbar-title>Filter</v-toolbar-title>
      </v-toolbar>

      <v-list>
        <v-list-item title="Reset (show all, most recent first)" @click="updateFilter({}, true)"
          ><template v-slot:prepend> <v-icon icon="mdi-arrow-u-left-top"></v-icon> </template
        ></v-list-item>
        <v-divider></v-divider>

        <v-list-subheader inset>ORDER</v-list-subheader>
        <v-list-item
          title="Longest"
          @click="updateFilter({ orderBy: 'length_px * px_per_m DESC' })"
        >
          <template v-slot:prepend>
            <v-icon icon="mdi-arrow-expand-horizontal"></v-icon>
          </template>
        </v-list-item>
        <v-list-item
          title="Shortest"
          @click="updateFilter({ orderBy: 'length_px * px_per_m ASC' })"
        >
          <template v-slot:prepend>
            <v-icon icon="mdi-arrow-collapse-horizontal"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Fastest"
          @click="updateFilter({ orderBy: 'ABS(speed_px_s * px_per_m) DESC' })"
        >
          <template v-slot:prepend> <v-icon icon="mdi-speedometer"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Slowest"
          @click="updateFilter({ orderBy: 'ABS(speed_px_s * px_per_m) ASC' })"
        >
          <template v-slot:prepend> <v-icon icon="mdi-speedometer-slow"></v-icon> </template
        ></v-list-item>
        <v-divider></v-divider>

        <v-list-subheader inset>FILTER</v-list-subheader>
        <v-list-item
          title="Today"
          @click="
            updateFilter({
              where: { start_ts: `DATE(start_ts) = DATE('${DateTime.now().toSQLDate()}')` }
            })
          "
          ><template v-slot:prepend> <v-icon icon="mdi-calendar-today"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Yesterday"
          @click="
            updateFilter({
              where: {
                start_ts: `DATE(start_ts) = DATE('${DateTime.now()
                  .minus({ days: 1 })
                  .toSQLDate()}')`
              }
            })
          "
          ><template v-slot:prepend> <v-icon icon="mdi-calendar-arrow-left"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Right"
          @click="
            updateFilter({
              where: {
                dir: `speed_px_s > 0`
              }
            })
          "
          ><template v-slot:prepend> <v-icon icon="mdi-arrow-right"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Left"
          @click="
            updateFilter({
              where: {
                dir: `speed_px_s < 0`
              }
            })
          "
          ><template v-slot:prepend> <v-icon icon="mdi-arrow-left"></v-icon> </template
        ></v-list-item>
      </v-list>
    </v-card>
  </v-dialog>
</template>
