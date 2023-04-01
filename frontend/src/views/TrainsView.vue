<script setup lang="ts">
import Trains from '@/components/Trains.vue'
</script>

<template>
  <main>
    <h1>Trains</h1>
    <Suspense>
      <Trains />
    </Suspense>
  </main>
</template>
