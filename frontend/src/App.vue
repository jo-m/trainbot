<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import logoDayUrl from '@/assets/logo-day.svg'
import logoNightUrl from '@/assets/logo-night.svg'
import { useTheme } from 'vuetify'

const devMode = import.meta.env.MODE === 'development'

const theme = useTheme()
function toggleTheme() {
  theme.global.name.value = theme.global.current.value.dark ? 'light' : 'dark'
}
</script>

<template>
  <v-app>
    <v-app-bar color="primary">
      <router-link
        :to="{ name: 'root' }"
        style="text-decoration: none; color: inherit; margin-inline-start: 16px; padding-top: 6px"
      >
        <img
          width="48"
          :src="theme.global.current.value.dark ? logoNightUrl : logoDayUrl"
          style="margin-left: -16px; margin-top: -2px"
        />
      </router-link>

      <v-spacer></v-spacer>

      <div id="app-bar-teleport"></div>

      <!-- Dark mode toggle - only in development -->
      <v-btn v-if="devMode" variant="text" icon="mdi-theme-light-dark" @click="toggleTheme"></v-btn>

      <!-- Github link -->
      <v-btn variant="text" icon="mdi-github" href="https://github.com/jo-m/trainbot"></v-btn>
    </v-app-bar>

    <v-main>
      <Suspense>
        <router-view />

        <template v-slot:fallback>Loading...</template>
      </Suspense>
    </v-main>
  </v-app>
</template>
