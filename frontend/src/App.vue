<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <v-app>
    <v-app-bar color="primary">
      <router-link
        :to="{ name: 'root' }"
        style="text-decoration: none; color: inherit; margin-inline-start: 16px"
      >
        <v-toolbar-title>Trainspotter</v-toolbar-title>
      </router-link>

      <v-spacer></v-spacer>

      <div id="app-bar-teleport"></div>

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
