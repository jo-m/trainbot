<script setup lang="ts">
import { inject } from 'vue'
import { dbKey, queryOne } from '@/lib/db'
import type SqlJs from 'sql.js'

const db = inject(dbKey) as SqlJs.Database

const avgSpeedMPS = queryOne(
  db,
  `SELECT
      SUM(ABS(speed_px_s / px_per_m)) / COUNT(*)
  FROM trains;`
) as number

const avgLengthM = queryOne(
  db,
  `SELECT
      SUM(ABS(length_px / px_per_m)) / COUNT(*)
  FROM trains;`
) as number

const histDoW = db.exec(
  `SELECT
      -- have to shift by 1, because SQLite thinks 0 = Sunday
      (CAST(strftime('%w', start_ts) AS INT) + 6) % 7 AS dow,
      COUNT(*)
  FROM trains
  GROUP BY dow
  ORDER BY dow;`
)[0].values as number[][]
const histDoWMax = histDoW.reduce((prev, cur) => Math.max(prev, cur[1]), 0)

const histHoD = db.exec(
  `SELECT
      CAST(strftime('%H', start_ts) AS INT) AS hod,
      COUNT(*)
  FROM trains
  GROUP BY hod
  ORDER BY hod;`
)[0].values as number[][]
const histHoDMax = histHoD.reduce((prev, cur) => Math.max(prev, cur[1]), 0)

const dayOfWeek: { [key: number]: string } = {
  0: 'Mon',
  1: 'Tue',
  2: 'Wed',
  3: 'Thu',
  4: 'Fri',
  5: 'Sat',
  6: 'Sun'
}
</script>

<template>
  <v-card>
    <v-card-item>
      <v-card-title>Stats</v-card-title>
    </v-card-item>

    <v-card-text>
      <v-table>
        <thead>
          <tr>
            <th class="text-left">Name</th>
            <th class="text-left">Value</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Number of trains</td>
            <td>{{ queryOne(db, 'SELECT COUNT(*) FROM trains;') }}</td>
          </tr>
          <tr>
            <td>Going right</td>
            <td>{{ queryOne(db, 'SELECT COUNT(*) FROM trains WHERE speed_px_s > 0;') }}</td>
          </tr>
          <tr>
            <td>Going left</td>
            <td>{{ queryOne(db, 'SELECT COUNT(*) FROM trains WHERE speed_px_s < 0;') }}</td>
          </tr>
          <tr>
            <td>Average speed</td>
            <td>{{ Math.round(avgSpeedMPS * 3.6 * 10) / 10 }} km/h</td>
          </tr>
          <tr>
            <td>Average length</td>
            <td>{{ Math.round(avgLengthM * 10) / 10 }} m</td>
          </tr>
          <tr>
            <td>Hist: Day of week</td>
            <td>
              <div
                v-for="[day, count] in histDoW"
                :key="day"
                :style="{
                  color: 'white',
                  backgroundColor: '#aa0000',
                  width: `${(count / histDoWMax) * 150}px`,
                  margin: '4px',
                  fontFamily: 'monospace'
                }"
              >
                &nbsp;{{ dayOfWeek[day] }}: {{ count }}
              </div>
            </td>
          </tr>
          <tr>
            <td>Hist: Hour of day</td>
            <td>
              <div
                v-for="[hour, count] in histHoD"
                :key="hour"
                :style="{
                  color: 'white',
                  backgroundColor: '#009900',
                  width: `${(count / histHoDMax) * 150}px`,
                  margin: '4px',
                  fontFamily: 'monospace',
                  overflow: 'clip',
                  whiteSpace: 'nowrap'
                }"
              >
                &nbsp;{{ hour }}: {{ count }}
              </div>
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card-text>
  </v-card>
</template>
