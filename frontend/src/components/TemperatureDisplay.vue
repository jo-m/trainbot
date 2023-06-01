<script setup lang="ts">
import { inject } from 'vue'
import { dbKey } from '@/lib/db'
import type SqlJs from 'sql.js'
import { DateTime } from 'luxon'

const db = inject(dbKey) as SqlJs.Database

const tempRes = db.exec('SELECT timestamp, temp_deg_c FROM temperatures ORDER BY id DESC LIMIT 1')
const tempDegC = tempRes[0].values[0][1]
const tempTS = DateTime.fromSQL(tempRes[0].values[0][0] as string)
</script>

<template>
  <span
    :title="`Raspberry Pi Core Temperature, at ${tempTS.toLocaleString(DateTime.DATETIME_FULL)}`"
  >
    {{ Math.round(tempDegC as number) }} °C
  </span>
</template>
