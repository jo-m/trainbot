<script setup lang="ts">
import { ref, computed } from 'vue'
import useFavoritesStore from '@/lib/favorites'
import { encodeQueryParam } from '@/lib/useQueryParam'
import router from '@/plugins/router'
import type { RouteLocationRaw } from 'vue-router'

const showDialog = ref<boolean>(false)

const showSnackbar = ref<boolean>(false)
const snackbarMessage = ref<string>('')

const baseURL = import.meta.env.VITE_BASE_URL.replace(/\/+$/, '')
function resolveFull(to: RouteLocationRaw): string {
  return baseURL + '/' + router.resolve(to).href
}

const favs = useFavoritesStore()
const linkParams = computed(() => {
  const ids = Array.from(favs.favorites).join(',')
  const filter = { where: { favs: `id IN (${ids})` } }
  return { name: 'trainsList', query: { filter: encodeQueryParam(filter) } }
})

const link = computed(() => {
  return resolveFull(linkParams.value)
})

const linkText = resolveFull({ name: 'trainsList', query: { filter: '...' } })

function copyLink() {
  navigator.clipboard.writeText(link.value).then(
    () => {
      snackbarMessage.value = 'Copied to clipboard'
      showSnackbar.value = true
    },
    () => {
      snackbarMessage.value = 'Failed to copy'
      showSnackbar.value = true
    }
  )
}
</script>

<template>
  <v-btn variant="text" icon>
    <v-icon>mdi-share</v-icon>

    <v-dialog v-model="showDialog" activator="parent" width="auto">
      <v-card>
        <v-card-title>Share</v-card-title>
        <v-divider></v-divider>
        <v-card-text>
          Share the link below to share your favorites list:
          <br />
          <span style="font-family: monospace">
            <router-link :to="linkParams">{{ linkText }}</router-link>
          </span>
        </v-card-text>
        <v-card-actions>
          <v-btn color="primary" variant="flat" @click="copyLink">Copy to clipboard</v-btn>
          <v-btn color="primary" variant="flat" @click="showDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-btn>

  <v-snackbar v-model="showSnackbar" :timeout="1000">
    {{ snackbarMessage }}

    <template v-slot:actions>
      <v-btn color="secondary" variant="outlined" @click="showSnackbar = false"> Close </v-btn>
    </template>
  </v-snackbar>
</template>
