<template>
  <v-container fluid :style="{ height: display.lgAndUp.value ? '100vh' : 'auto' }">
    <v-row v-if="display.lgAndUp.value" align="center" justify="center">
      <div>
        <v-img
          class="mx-auto mb-4"
          max-height="400"
          max-width="90%"
          contain
          src="~/assets/img/person-texting.svg"
        />
        <div class="text-center">
          <h3 class="text-h5 mt-4">Select a Message</h3>
          <p class="text-medium-emphasis">
            Don't hesitate to
            <a
              href="https://discord.gg/kGk8HVqeEZ"
              target="_blank"
              class="text-decoration-none"
            >message us on Discord</a>
            if you have any questions
          </p>
        </div>
      </div>
    </v-row>
    <v-row v-else justify="end">
      <v-col class="px-0 py-0">
        <MessageThreadHeader />
        <MessageThread />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'

definePageMeta({ middleware: ['auth'] })
useHead({ title: 'Threads - httpSMS' })

const store = useAppStore()
const display = useDisplay()

onMounted(async () => {
  await store.loadUser()
  await store.loadPhones()
  await store.loadThreads()
})
</script>
