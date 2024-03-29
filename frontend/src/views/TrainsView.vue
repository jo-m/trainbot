<script setup lang="ts">
import TrainList from '@/components/TrainList.vue'
import TrainGrid from '@/components/TrainGrid.vue'
import FilterDialog, { type updateFilterArgs } from '@/components/FilterDialog.vue'
import StaleDataWarningDialog from '@/components/StaleDataWarningDialog.vue'
import FavoritesDialog from '@/components/FavoritesDialog.vue'
import { ref, onMounted, onUnmounted, inject, watch } from 'vue'
import { dbKey, getTrains, type Train as TrainType, type Filter } from '@/lib/db'
import useQueryParam from '@/lib/useQueryParam'
import type SqlJs from 'sql.js'

const db = inject(dbKey) as SqlJs.Database

const tileView = useQueryParam<boolean>('tiles', false)

// How many trains to load at a time.
const pageSize = 60
// Currently loaded data.
const trains = ref<TrainType[] | null>(null)
const filteredCount = ref<number | null>(null)
const totalCount = ref<number | null>(null)
// If we have reached the end of pagination.
const allDataLoaded = ref<boolean>(false)
// Currently active filter.
// When this changes, data must be reset.
const filter = useQueryParam<Filter>('filter', {})

const scroller = ref<HTMLDivElement | null>(null)
const showFilterDialog = ref<boolean>(false)
const filterSnackbarShow = ref<boolean>(true)

function updateFilter(args: updateFilterArgs) {
  const { newFilter, replace } = args

  if (replace) {
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

  showFilterDialog.value = false
}

watch(filter, () => {
  trains.value = null
  filteredCount.value = null
  totalCount.value = null
  allDataLoaded.value = false

  loadNextData()
})

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

watch(filteredCount, () => {
  filterSnackbarShow.value = true
})

onMounted(async () => {
  loadNextData()

  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})

function scrollUp() {
  window.scrollTo(0, 0)
}
</script>

<template>
  <!-- App bar -->
  <Teleport to="#app-bar-teleport">
    <StaleDataWarningDialog />

    <v-btn
      variant="text"
      icon="mdi-chart-bar"
      :to="{ name: 'trainStats' }"
      aria-label="Show the stats page"
    ></v-btn>

    <v-btn
      variant="text"
      :icon="tileView ? 'mdi-view-list' : 'mdi-view-grid'"
      @click="tileView = !tileView"
      aria-label="Toggle between grid and list view"
    ></v-btn>

    <v-btn
      variant="text"
      icon="mdi-filter"
      @click="showFilterDialog = true"
      :active="Object.keys(filter).length > 0"
      aria-label="Show the filter dialog"
    ></v-btn>

    <FavoritesDialog />
  </Teleport>

  <!-- List -->
  <template v-if="trains !== null">
    <div ref="scroller">
      <TrainList v-if="!tileView" :trains="trains" />
      <TrainGrid v-else :trains="trains" />
    </div>
    <v-container class="pa-2">
      <v-row no-gutters>
        <v-col cols="12">
          <v-row class="pa-0" no-gutters align="center">
            <v-col cols="12">
              <v-sheet class="pa-2" v-if="allDataLoaded">
                <v-icon icon="mdi-arrow-collapse-down"></v-icon>
                End of list ({{ trains.length }} trains).
                <v-btn
                  variant="tonal"
                  size="small"
                  @click="updateFilter({ newFilter: {}, replace: true })"
                  >Reset filters.</v-btn
                >
              </v-sheet>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-container>
  </template>

  <!-- Filter -->
  <FilterDialog
    :show="showFilterDialog"
    @updateFilter="updateFilter($event)"
    @close="showFilterDialog = false"
  />

  <!-- Filter Snackbar -->
  <v-snackbar v-model="filterSnackbarShow" :timeout="4000">
    Current filter includes {{ filteredCount === totalCount ? 'all' : `${filteredCount} of` }}
    {{ totalCount }} trains.

    <template v-slot:actions>
      <v-btn color="secondary" variant="outlined" @click="filterSnackbarShow = false">
        Close
      </v-btn>
    </template>
  </v-snackbar>

  <v-layout-item model-value position="bottom" class="text-end" size="88">
    <div class="ma-4">
      <v-btn
        icon="mdi-chevron-up"
        size="large"
        color="primary"
        elevation="8"
        @click="scrollUp()"
        aria-label="Scroll to top"
      />
    </div>
  </v-layout-item>
</template>
