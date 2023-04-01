import SqlJs from 'sql.js'
import sqlWasmUrl from 'sql.js/dist/sql-wasm.wasm?url'
import { DateTime } from 'luxon'

let sqlJs: SqlJs.SqlJsStatic | null = null

async function loadSqlJS(): Promise<SqlJs.SqlJsStatic> {
  if (sqlJs !== null) {
    return sqlJs
  }

  sqlJs = await SqlJs({ locateFile: () => sqlWasmUrl })

  return sqlJs
}

export async function loadDB(): Promise<SqlJs.Database> {
  const url = import.meta.env.VITE_DB_URL
  const dbFile = await fetch(url)
  const dbBuf = await dbFile.arrayBuffer()
  return new (await loadSqlJS()).Database(new Uint8Array(dbBuf))
}

export interface Train {
  id: number
  start_ts: DateTime
  end_ts: DateTime
  n_frames: number
  length_px: number
  speed_px_s: number
  accel_px_s_2: number
  px_per_m: number
  image_file_path: string
  gif_file_path: string
  // TODO: Parse dates
  uploaded_at: DateTime
}

function convertValue(colname: string, value: any): any {
  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
  if (value !== null && ['start_ts', 'end_ts', 'uploaded_at'].indexOf(colname) != -1) {
    return DateTime.fromSQL(value, { setZone: true })
  }
  return value
}

function convertRow(cols: string[], row: any[]) {
  return Object.fromEntries(cols.map((colname, ix) => [colname, convertValue(colname, row[ix])]))
}

export function getTrains(db: SqlJs.Database, limit: number, offset: number): Train[] {
  const result = db.exec(
    `SELECT * FROM trains ORDER BY start_ts DESC LIMIT ${limit} OFFSET ${offset}`
  )[0]

  return result.values.map((row) => convertRow(result.columns, row)) as any
}
