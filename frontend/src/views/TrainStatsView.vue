<script setup lang="ts">
import { inject } from 'vue'
import { dbKey, queryOne } from '@/lib/db'
import type SqlJs from 'sql.js'
import {
  avgLengthM,
  histCountPerDayOfWeek,
  dayOfWeekLabels,
  histHourOfDay,
  histSpeedKPH,
  tempDegCPast24hAvg
} from '@/lib/stats'

import VerticalHist from '@/components/VerticalHist.vue'
const db = inject(dbKey) as SqlJs.Database
const widthPx = 200
</script>

<template>
  <v-card>
    <v-card-item>
      <v-card-title>Stats</v-card-title>
    </v-card-item>

    <v-card-text>
      <v-table>
        <tbody>
          <tr>
            <td>Number of trains</td>
            <td>{{ queryOne(db, 'SELECT COUNT(*) FROM trains_v2;') }}</td>
          </tr>
          <tr>
            <td>Going right</td>
            <td>{{ queryOne(db, 'SELECT COUNT(*) FROM trains_v2 WHERE speed_px_s > 0;') }}</td>
          </tr>
          <tr>
            <td>Going left</td>
            <td>{{ queryOne(db, 'SELECT COUNT(*) FROM trains_v2 WHERE speed_px_s < 0;') }}</td>
          </tr>
          <tr>
            <td>Average length</td>
            <td>{{ Math.round(avgLengthM(db) * 10) / 10 }} m</td>
          </tr>
          <tr>
            <td>Hist: Speed [&gt; km/h]</td>
            <td>
              <VerticalHist
                :data="histSpeedKPH(db)"
                :width-px="widthPx"
                color="#0000aa"
              ></VerticalHist>
            </td>
          </tr>
          <tr>
            <td>Hist: Day of week</td>
            <td>
              <VerticalHist
                :data="histCountPerDayOfWeek(db)"
                :labels="dayOfWeekLabels"
                :width-px="widthPx"
                color="#aa0000"
              ></VerticalHist>
            </td>
          </tr>
          <tr>
            <td>Hist: Hour of day</td>
            <td>
              <VerticalHist
                :data="histHourOfDay(db)"
                :width-px="widthPx"
                color="#009900"
              ></VerticalHist>
            </td>
          </tr>
          <tr>
            <td>Avg core temperature 24h [°C]</td>
            <td>
              <VerticalHist
                :data="tempDegCPast24hAvg(db)"
                :width-px="widthPx"
                :labels="
                  Object.fromEntries(
                    Array.from({ length: 24 }, (x, i) => [i, `00${i}:00`.slice(-5)])
                  )
                "
                color="#cc0000"
              ></VerticalHist>
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card-text>
  </v-card>
</template>
