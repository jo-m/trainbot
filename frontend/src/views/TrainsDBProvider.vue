<script setup lang="ts">
import { RouterView } from 'vue-router'
import { provide } from 'vue'
import { loadDB, dbKey } from '@/lib/db'

provide(dbKey, await loadDB())
</script>

<template>
  <router-view></router-view>
</template>
