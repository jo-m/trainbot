import { queryOne } from '@/lib/db'
import type SqlJs from 'sql.js'

export const avgSpeedMPS = (db: SqlJs.Database): number =>
  queryOne(
    db,
    `SELECT
        SUM(ABS(speed_px_s / px_per_m)) / COUNT(*)
    FROM trains;`
  ) as number

export const avgLengthM = (db: SqlJs.Database): number =>
  queryOne(
    db,
    `SELECT
        SUM(ABS(length_px / px_per_m)) / COUNT(*)
    FROM trains;`
  ) as number

export const histCountPerDayOfWeek = (db: SqlJs.Database): number[][] =>
  db.exec(
    `SELECT
        -- have to shift by 1, because SQLite thinks 0 = Sunday
        (CAST(strftime('%w', start_ts) AS INT) + 6) % 7 AS dow,
        COUNT(*)
    FROM trains
    GROUP BY dow
    ORDER BY dow;`
  )[0].values as number[][]

export const dayOfWeekLabels: { [key: number]: string } = {
  0: 'Mon',
  1: 'Tue',
  2: 'Wed',
  3: 'Thu',
  4: 'Fri',
  5: 'Sat',
  6: 'Sun'
}

export const histHourOfDay = (db: SqlJs.Database): number[][] =>
  db.exec(
    `SELECT
      CAST(strftime('%H', start_ts) AS INT) AS hod,
      COUNT(*)
  FROM trains
  GROUP BY hod
  ORDER BY hod;`
  )[0].values as number[][]
