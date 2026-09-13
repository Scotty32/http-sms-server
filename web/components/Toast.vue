<template>
  <v-snackbar
    :model-value="store.getNotification.active"
    :color="store.getNotification.type"
    :timeout="store.getNotification.timeout"
    @update:model-value="store.disableNotification()"
  >
    <v-icon
      v-if="store.getNotification.type === 'success'"
      :icon="mdiCheck"
      :color="store.getNotification.type"
    />
    <v-icon
      v-if="store.getNotification.type === 'info'"
      :icon="mdiInformation"
      :color="store.getNotification.type"
    />
    {{ store.getNotification.message }}
    <template #actions>
      <v-btn
        v-if="display.lgAndUp.value"
        :color="store.getNotification.type"
        variant="text"
        @click="store.disableNotification()"
      >
        <span class="font-weight-bold">Close</span>
      </v-btn>
    </template>
  </v-snackbar>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'
import { mdiCheck, mdiInformation } from '@mdi/js'

const store = useAppStore()
const display = useDisplay()
</script>
