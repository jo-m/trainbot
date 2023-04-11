import { computed, type WritableComputedRef } from 'vue'
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

export default function useQueryParam<T>(name: string, default_: T): WritableComputedRef<T> {
  const route = useRoute()

  const value = computed({
    get() {
      return parseQueryParam<T>(route.query[name], default_)
    },
    set(newValue) {
      router.push({ query: { [name]: encodeQueryParam(newValue) } })
    }
  })

  return value
}
