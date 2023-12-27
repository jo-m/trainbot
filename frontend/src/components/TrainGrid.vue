<script setup lang="ts">
import type { Train as TrainType } from '@/lib/db'
import { getBlobURL, gifFileName } from '@/lib/paths'
import RelativeTime from '@/components/RelativeTime.vue'
import FavoriteIcon from '@/components/FavoriteIcon.vue'

defineProps<{
  trains: TrainType[]
}>()
</script>

<template>
  <v-container fluid>
    <v-row dense>
      <v-col v-for="train in trains" v-bind:key="train.id" cols="6" sm="3" md="2" xl="1">
        <router-link
          :to="{ name: 'trainDetail', params: { id: train.id } }"
          style="text-decoration: none; color: inherit"
        >
          <v-card>
            <v-img
              :src="getBlobURL(gifFileName(train.start_ts))"
              class="align-end"
              gradient="to bottom, rgba(0,0,0,.1), rgba(0,0,0,.5)"
              height="200px"
              cover
            >
              <v-card-title class="text-white">
                <RelativeTime :ts="train.start_ts" />
                <FavoriteIcon :id="train.id" />
              </v-card-title>
            </v-img>
          </v-card>
        </router-link>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped>
.pointer {
  cursor: pointer;
}
</style>
