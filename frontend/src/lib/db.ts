import SqlJs from 'sql.js'
import sqlWasmUrl from 'sql.js/dist/sql-wasm.wasm?url'
import { DateTime } from 'luxon'
import type { InjectionKey } from 'vue'

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
  uploaded_at?: DateTime
}

function convertValue(colname: string, value: any): any {
  if (value !== null && ['start_ts', 'end_ts', 'uploaded_at'].indexOf(colname) != -1) {
    return DateTime.fromSQL(value, { setZone: true })
  }
  return value
}

function convertRow(cols: string[], row: any[]) {
  const ret = Object.fromEntries(
    cols.map((colname, ix) => [colname, convertValue(colname, row[ix])])
  )
  Object.freeze(ret)
  return ret
}

interface Result {
  trains: Train[]
  filteredCount: number
  totalCount: number
}

function convertTrains(result: SqlJs.QueryExecResult[]): Train[] {
  if (result.length === 0) {
    return []
  }

  return result[0].values.map((row) => convertRow(result[0].columns, row)) as Train[]
}

export interface Filter {
  // ORDER BY clause, e.g. 'id DESC'.
  orderBy?: string
  // The values are SQL clauses and are AND'ed together.
  // The keys are just for managing multiple filters.
  where?: { [key: string]: string }
}

export function getTrains(
  db: SqlJs.Database,
  limit: number,
  offset: number = 0,
  filter: Filter = {}
): Result {
  const { orderBy, where } = filter
  const orderByStr = orderBy || 'start_ts DESC'
  const whereStr =
    (where !== undefined && Object.keys(where).length && Object.values(where).join(' AND ')) ||
    '1=1'

  // We don't care about SQL injections, because this all happens client side.
  // Muahahah.
  const query = `
    SELECT *
    FROM trains
    WHERE ${whereStr}
    ORDER BY ${orderByStr}
    LIMIT ${limit} OFFSET ${offset}`

  console.log(query.trim())

  const result = db.exec(query)

  const filteredCount = db.exec(`SELECT COUNT(*) FROM trains WHERE ${whereStr}`)
  const totalCount = db.exec(`SELECT COUNT(*) FROM trains`)

  return {
    trains: convertTrains(result),
    filteredCount: filteredCount[0].values[0][0] as number,
    totalCount: totalCount[0].values[0][0] as number
  }
}

export function getTrain(db: SqlJs.Database, id: number): Train | undefined {
  const query = `
  SELECT *
  FROM trains
  WHERE id = ${id}`

  console.log(query.trim())

  const result = db.exec(query)
  if (result.length === 0) return undefined

  return convertRow(result[0].columns, result[0].values[0]) as Train
}

export const dbKey = Symbol() as InjectionKey<SqlJs.Database>
