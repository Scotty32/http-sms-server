<template>
  <div>
    <v-progress-linear
      v-if="store.getLoadingThreads"
      color="primary"
      indeterminate
    />
    <div
      v-if="!store.getLoadingThreads && store.getIsArchived"
      class="bg-warning py-1 text-center text-uppercase text-subtitle-1"
    >
      Archived Messages
    </div>
    <v-sheet
      v-if="!store.getLoadingThreads && threads.length === 0 && !store.getIsArchived"
      class="text-center mt-8 mx-3"
      :color="display.mdAndDown.value ? '#121212' : '#363636'"
    >
      <div v-if="display.mdAndDown.value">
        <v-img
          class="mx-auto mb-4"
          max-width="80%"
          contain
          src="~/assets/img/person-texting.svg"
        />
        <p v-if="store.getOwner" class="text-medium-emphasis">
          Start sending messages
        </p>
      </div>
      <v-btn
        v-if="store.getOwner && store.getPhones.length !== 0"
        color="primary"
        :to="{ name: 'messages' }"
      >
        <v-icon :icon="mdiPlus" />
        New Message
      </v-btn>
    </v-sheet>
    <div
      v-if="store.getPhones.length === 0 && store.getLoadingThreads === false"
      class="px-4 text-center"
    >
      <p>
        Install the mobile app on your android phone to start sending messages.
        You can also
        <a href="https://discord.gg/kGk8HVqeEZ" target="_blank" class="text-decoration-none">
          message us on Discord
        </a>
        to help set things up.
      </p>
      <v-btn
        color="primary"
        :href="store.getAppData.appDownloadUrl"
        @click="store.addNotification({ type: 'info', message: 'Downloading the httpSMS Android App' })"
      >
        <v-icon :icon="mdiDownload" />
        Install App
      </v-btn>
    </div>
    <v-list lines="two" class="px-0 py-0">
      <v-list-item
        v-for="thread in threads"
        :key="thread.id"
        :to="'/threads/' + thread.id"
        class="py-1"
        :class="{
          'px-6': display.mdAndDown.value,
          'px-3': display.lgAndUp.value,
        }"
      >
        <template #prepend>
          <v-avatar :color="thread.color">
            <v-icon dark :icon="mdiAccount" />
          </v-avatar>
        </template>
        <v-list-item-title>{{ formatPhoneNumber(thread.contact) }}</v-list-item-title>
        <v-list-item-subtitle>{{ thread.last_message_content }}</v-list-item-subtitle>
        <template #append>
          <div class="d-flex flex-column align-center">
            <span class="text-caption text-medium-emphasis">{{ threadDate(thread.order_timestamp) }}</span>
            <v-icon v-if="thread.status === 'expired'" color="warning" size="small" :icon="mdiAlert" />
            <v-icon v-else-if="thread.status === 'delivered'" color="primary" size="small" :icon="mdiCheckAll" />
            <v-icon v-else-if="thread.status === 'received'" color="success" size="small" :icon="mdiCheckAll" />
            <v-icon v-else-if="thread.status === 'sent'" size="small" :icon="mdiCheck" />
            <v-icon v-else-if="thread.status === 'failed'" color="error" size="small" :icon="mdiAlert" />
          </div>
        </template>
      </v-list-item>
    </v-list>
  </div>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'
import { mdiPlus, mdiDownload, mdiCheckAll, mdiCheck, mdiAlert, mdiAccount } from '@mdi/js'
import { formatPhoneNumber } from '~/plugins/filters'

const store = useAppStore()
const display = useDisplay()

const threads = computed(() => store.getThreads)

function threadDate(date: string): string {
  return new Date(date).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
  })
}
</script>
