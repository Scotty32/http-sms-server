<template>
  <div>
    <NuxtLink to="/" class="text-decoration-none d-flex">
      <v-avatar rounded="0" size="33" class="mt-1">
        <v-img contain :src="logoImg"></v-img>
      </v-avatar>
      <h3 class="text-h4 text-high-emphasis ml-1">httpSMS</h3>
    </NuxtLink>
    <p>
      httpSMS is an
      <a
        class="text-decoration-none"
        href="https://github.com/NdoleStudio/httpsms"
        >open source</a
      >
      application that converts your android phone into an SMS gateway so you
      can send and receive SMS messages using a simple HTTP API.
    </p>
    <v-btn :href="store.getAppData.documentationUrl">
      <v-icon start :icon="mdiBookOpenVariant" />
      Documentation
    </v-btn>
  </div>
</template>

<script setup lang="ts">
import { mdiBookOpenVariant } from '@mdi/js'
import logoImg from '@/assets/img/logo.svg'
import { useAppStore } from '~/stores/app'

const store = useAppStore()
</script>
