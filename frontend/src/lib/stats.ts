import { queryOne } from '@/lib/db'
import type SqlJs from 'sql.js'

export const avgLengthM = (db: SqlJs.Database): number =>
  queryOne(
    db,
    `SELECT
      SUM(ABS(length_px / px_per_m)) / COUNT(*)
    FROM trains_v2;`
  ) as number

export const histCountPerDayOfWeek = (db: SqlJs.Database): number[][] =>
  db.exec(`
    SELECT
      -- have to shift by 1, because SQLite thinks 0 = Sunday
      (CAST(strftime('%w', start_ts) AS INT) + 6) % 7 AS dow,
      COUNT(*)
    FROM trains_v2
    GROUP BY dow
    ORDER BY dow;`)[0].values as number[][]

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
  db.exec(`
  SELECT
    CAST(strftime('%H', start_ts) AS INT) AS hod,
    COUNT(*)
  FROM trains_v2
  GROUP BY hod
  ORDER BY hod;`)[0].values as number[][]

// No magic needed to fill in missing bins, we have values for everything.
export const histSpeedKPH = (db: SqlJs.Database, binSz: number = 10): number[][] =>
  db.exec(`
  WITH speed_rounded AS (
    SELECT
      CAST(ABS(speed_px_s / px_per_m * 3.6)/${binSz} AS INTEGER) * ${binSz} AS speed_rounded
    FROM trains_v2
  )
  SELECT
    speed_rounded,
    COUNT(*)
  FROM speed_rounded
  GROUP BY speed_rounded
  ORDER BY speed_rounded;`)[0].values as number[][]

export const tempDegCPast24hAvg = (db: SqlJs.Database): number[][] => {
  const res = db.exec(`
  SELECT
    CAST(strftime('%H', timestamp) AS INT) AS hod,
    ROUND(AVG(temp_deg_c))
  FROM temperatures
  WHERE timestamp >= DATETIME('now','-24 hours')
  GROUP BY hod
  ORDER BY hod ASC`)

  if (res.length == 0) {
    return []
  }

  return res[0].values as number[][]
}
