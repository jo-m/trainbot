import { ref, type Ref, type UnwrapRef, watch } from 'vue'
import { useRoute, type LocationQueryValue } from 'vue-router'
import router from '@/plugins/router'

function parseQueryParam<T>(val: LocationQueryValue | LocationQueryValue[], default_: T): T {
  if (!val || Array.isArray(val)) {
    return default_
  }

  try {
    return JSON.parse(decodeURIComponent(val))
  } catch {
    return default_
  }
}

function encodeQueryParam(val: any): string {
  return encodeURIComponent(JSON.stringify(val))
}

let skipNextRefUpdate = false
let skipNextUrlUpdate = false

export default function useQueryParam<T>(name: string, default_: T): Ref<UnwrapRef<T>> {
  const route = useRoute()
  const refVal = ref<T>(parseQueryParam<T>(route.query[name], default_))

  watch(
    () => route.query[name],
    (val) => {
      if (skipNextUrlUpdate) {
        skipNextUrlUpdate = false
        return
      }

      skipNextRefUpdate = true
      refVal.value = parseQueryParam<T>(val, default_) as UnwrapRef<T>
    }
  )

  watch(refVal, (val) => {
    if (skipNextRefUpdate) {
      skipNextRefUpdate = false
      return
    }
    skipNextUrlUpdate = true
    router.push({ query: { filter: encodeQueryParam(val) } })
  })

  return refVal
}
