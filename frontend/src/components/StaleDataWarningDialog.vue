<script setup lang="ts">
import { dbKey, getTrains } from '@/lib/db'
import { ref, inject } from 'vue'
import type SqlJs from 'sql.js'
import { DateTime } from 'luxon'

const showWarningIfOlderThanHours = 8

const showDialog = ref<boolean>(false)
const db = inject(dbKey) as SqlJs.Database

const train = getTrains(db, 1, 0, { orderBy: 'start_ts DESC' }).trains
const date = train.length > 0 ? train[0].start_ts : DateTime.fromMillis(0)
const agoHours = DateTime.now().diff(date, 'hours').hours
</script>

<template>
  <v-btn variant="text" color="warning" icon v-if="agoHours > showWarningIfOlderThanHours">
    <v-icon>mdi-alert</v-icon>

    <v-dialog v-model="showDialog" activator="parent" width="auto">
      <v-card>
        <v-card-title>Stale Data Warning</v-card-title>
        <v-divider></v-divider>
        <v-card-text>
          The last data upload was {{ date.toRelative() }}. This probably means that the Raspberry
          Pi on my balcony is currently offline/broken. Note that this frontend is hosted
          independently. This will be fixed eventually.
        </v-card-text>
        <v-card-actions>
          <v-btn color="primary" variant="flat" block @click="showDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-btn>
</template>
