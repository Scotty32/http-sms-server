<template>
  <div class="d-flex mb-6 mt-8">
    <v-avatar class="mb-n2">
      <v-img :src="arnoldImg"></v-img>
    </v-avatar>
    <div class="ml-2">
      <p class="text-subtitle-1 mb-n1">Acho Arnold</p>
      <a
        class="text-decoration-none text-high-emphasis"
        href="https://twitter.com/acho_arnold"
      >
        <v-icon color="#1DA1F2" :icon="mdiTwitter" />
      </a>
      <a
        class="ml-2 text-decoration-none text-high-emphasis"
        href="https://github.com/AchoArnold"
      >
        <v-icon color="#FFFFFF" :icon="mdiGithub" />
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { mdiTwitter, mdiGithub } from '@mdi/js'
import arnoldImg from '@/assets/img/arnold.png'
</script>
