<script setup lang="ts">
import useFavoritesStore from '@/lib/favorites'

const favs = useFavoritesStore()

defineProps<{
  id: number
}>()
</script>

<template>
  <v-icon
    :icon="favs.isFav(id) ? 'mdi-star' : 'mdi-star-outline'"
    :color="favs.isFav(id) ? '#FFC107' : '#BDBDBD'"
    @click.prevent.stop="favs.toggle(id)"
    style="padding-bottom: 2px"
  />
</template>
