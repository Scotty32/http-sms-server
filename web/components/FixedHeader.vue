<template>
  <v-app-bar elevation="0" color="#121212" height="70" fixed>
    <v-container>
      <v-row>
        <v-col class="w-full d-flex">
          <NuxtLink
            :to="{ name: 'index' }"
            class="text--primary text-h4 text-decoration-none"
            :class="{
              'mt-5': display.mdAndDown.value,
              'mt-4': !display.mdAndDown.value,
            }"
          >
            <v-avatar v-if="display.lgAndUp.value" tile size="30">
              <v-img contain src="~/assets/img/logo.svg"></v-img>
            </v-avatar>
            HTTP SMS
          </NuxtLink>
          <v-spacer></v-spacer>
          <v-btn
            v-if="store.getAuthUser === null"
            exact-path
            color="primary"
            class="mt-5 mb-5"
            :size="display.lgAndUp.value ? 'large' : 'default'"
            :to="{ name: 'login' }"
          >
            Get Started
            <span v-if="display.lgAndUp.value">&nbsp;For Free</span>
          </v-btn>
          <v-btn
            v-else
            exact-path
            color="primary"
            class="mt-5 mb-5"
            :size="display.lgAndUp.value ? 'large' : 'default'"
            :to="{ name: 'threads' }"
          >
            Dashboard
          </v-btn>
        </v-col>
      </v-row>
    </v-container>
  </v-app-bar>
</template>
<script setup lang="ts">
import { useDisplay } from 'vuetify'

const store = useAppStore()
const display = useDisplay()
</script>
