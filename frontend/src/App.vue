<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <v-app>
    <v-app-bar color="primary">
      <v-toolbar-title>Trainspotter</v-toolbar-title>
      <!-- <router-link :to="{name:'root'}">Trains</router-link> -->

      <v-spacer></v-spacer>

      <div id="app-bar-teleport"></div>
      <v-btn variant="text" icon="mdi-github"></v-btn>
    </v-app-bar>

    <v-main style="min-height: 300px">
      <Suspense>
        <router-view />
      </Suspense>
    </v-main>
  </v-app>
</template>
