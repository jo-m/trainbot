<script setup lang="ts">
import type { Filter } from '@/lib/db'
import { DateTime } from 'luxon'

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'updateFilter', newFilter: Filter, replace: boolean): void
  (e: 'close'): void
}>()
</script>

<template>
  <v-dialog
    v-bind:model-value="show"
    fullscreen
    :scrim="false"
    transition="dialog-bottom-transition"
  >
    <v-card>
      <v-toolbar color="primary">
        <v-btn icon="" @click="emit('close')">
          <v-icon>mdi-close</v-icon>
        </v-btn>
        <v-toolbar-title>Filter</v-toolbar-title>
      </v-toolbar>

      <v-list>
        <v-list-item
          title="Reset (show all, most recent first)"
          @click="emit('updateFilter', {}, true)"
          ><template v-slot:prepend> <v-icon icon="mdi-arrow-u-left-top"></v-icon> </template
        ></v-list-item>
        <v-divider></v-divider>

        <v-list-subheader inset>ORDER</v-list-subheader>
        <v-list-item
          title="Longest"
          @click="emit('updateFilter', { orderBy: 'length_px * px_per_m DESC' }, false)"
        >
          <template v-slot:prepend>
            <v-icon icon="mdi-arrow-expand-horizontal"></v-icon>
          </template>
        </v-list-item>
        <v-list-item
          title="Shortest"
          @click="emit('updateFilter', { orderBy: 'length_px * px_per_m ASC' }, false)"
        >
          <template v-slot:prepend>
            <v-icon icon="mdi-arrow-collapse-horizontal"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Fastest"
          @click="emit('updateFilter', { orderBy: 'ABS(speed_px_s * px_per_m) DESC' }, false)"
        >
          <template v-slot:prepend> <v-icon icon="mdi-speedometer"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Slowest"
          @click="emit('updateFilter', { orderBy: 'ABS(speed_px_s * px_per_m) ASC' }, false)"
        >
          <template v-slot:prepend> <v-icon icon="mdi-speedometer-slow"></v-icon> </template
        ></v-list-item>
        <v-divider></v-divider>

        <v-list-subheader inset>FILTER</v-list-subheader>
        <v-list-item
          title="Today"
          @click="
            emit(
              'updateFilter',
              {
                where: { start_ts: `DATE(start_ts) = DATE('${DateTime.now().toSQLDate()}')` }
              },
              false
            )
          "
          ><template v-slot:prepend> <v-icon icon="mdi-calendar-today"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Yesterday"
          @click="
            emit(
              'updateFilter',
              {
                where: {
                  start_ts: `DATE(start_ts) = DATE('${DateTime.now()
                    .minus({ days: 1 })
                    .toSQLDate()}')`
                }
              },
              false
            )
          "
          ><template v-slot:prepend> <v-icon icon="mdi-calendar-arrow-left"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Right"
          @click="
            emit(
              'updateFilter',
              {
                where: {
                  dir: `speed_px_s > 0`
                }
              },
              false
            )
          "
          ><template v-slot:prepend> <v-icon icon="mdi-arrow-right"></v-icon> </template
        ></v-list-item>
        <v-list-item
          title="Left"
          @click="
            emit(
              'updateFilter',
              {
                where: {
                  dir: `speed_px_s < 0`
                }
              },
              false
            )
          "
          ><template v-slot:prepend> <v-icon icon="mdi-arrow-left"></v-icon> </template
        ></v-list-item>
      </v-list>
    </v-card>
  </v-dialog>
</template>
