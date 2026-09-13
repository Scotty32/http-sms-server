<template>
  <v-sheet
    class="pa-4 d-flex"
    :elevation="display.lgAndUp.value ? 0 : 2"
    :color="display.lgAndUp.value ? 'grey-darken-4' : 'black'"
  >
    <div :class="{ 'px-2': display.mdAndDown.value }">
      <v-toolbar-title>
        <div class="d-flex pt-2" style="width: 245px">
          <v-select
            variant="outlined"
            density="compact"
            :disabled="owners.length === 0"
            placeholder="Phone Numbers"
            :class="{ 'mb-n6': !store.getOwner }"
            :items="owners"
            :model-value="store.getOwner"
            @update:model-value="onOwnerChanged"
          />
          <div style="width: 50px">
            <v-progress-circular
              v-if="store.getPolling"
              indeterminate
              :size="20"
              :width="1"
              class="mt-3 ml-2"
              color="success"
            />
          </div>
        </div>
      </v-toolbar-title>
      <div v-if="store.getOwner" class="d-flex mt-n4">
        <p class="text-medium-emphasis mb-n1">
          {{ formatPhoneCountry(store.getOwner) }}
        </p>
        <v-tooltip
          v-if="store.getHeartbeat"
          location="right"
          :open-on-focus="false"
          :open-on-hover="true"
        >
          <template #activator="{ props }">
            <v-btn
              size="x-small"
              v-bind="props"
              :to="{
                name: 'heartbeats-id',
                params: { id: store.getOwner },
              }"
              color="success"
              class="ml-2 mt-1 mb-n1"
              icon
            >
              <v-icon
                v-if="store.getHeartbeat.charging"
                size="small"
                :icon="mdiBatteryCharging"
              />
              <v-icon v-else size="x-small" :icon="mdiCircle" />
            </v-btn>
          </template>
          <div>
            <h4>Last Heartbeat</h4>
            {{ formatHumanizeTime(store.getHeartbeat.timestamp) }} ago
          </div>
        </v-tooltip>
      </div>
    </div>
    <v-spacer />
    <v-menu>
      <template #activator="{ props }">
        <v-btn :icon="mdiDotsVertical" variant="text" class="mt-2" v-bind="props" />
      </template>
      <v-list class="px-2" nav :density="display.mdAndDown.value ? 'compact' : 'default'">
        <v-list-item
          :prepend-icon="store.getIsArchived ? mdiPackageUp : mdiPackageDown"
          @click.prevent="toggleArchive"
        >
          <v-list-item-title>
            {{ store.getIsArchived ? 'Unarchived' : 'Archived' }}
          </v-list-item-title>
        </v-list-item>
        <v-list-item
          v-if="store.getOwner"
          :prepend-icon="mdiPlus"
          :to="{ name: 'messages' }"
          exact
        >
          <v-list-item-title>New Message</v-list-item-title>
        </v-list-item>
        <v-list-item
          v-if="store.getOwner"
          :prepend-icon="mdiCommentTextMultipleOutline"
          :to="{ name: 'bulk-messages' }"
          exact
        >
          <v-list-item-title>Bulk Messages</v-list-item-title>
        </v-list-item>
        <v-list-item
          v-if="store.getOwner"
          :prepend-icon="mdiMagnify"
          :to="{ name: 'search-messages' }"
          exact
        >
          <v-list-item-title>Search Messages</v-list-item-title>
        </v-list-item>
        <v-list-item :prepend-icon="mdiAccountCog" :to="{ name: 'settings' }" exact>
          <v-list-item-title>Settings</v-list-item-title>
        </v-list-item>
        <v-list-item :prepend-icon="mdiCellphoneKey" :to="{ name: 'phone-api-keys' }" exact>
          <v-list-item-title>Phone API Keys</v-list-item-title>
        </v-list-item>
        <v-list-item
          v-if="store.getOwner"
          :prepend-icon="mdiDownload"
          :href="store.getAppData.appDownloadUrl"
          @click="store.addNotification({ type: 'info', message: 'Downloading the httpSMS Android App' })"
        >
          <v-list-item-title>Install App</v-list-item-title>
        </v-list-item>
        <v-list-item :prepend-icon="mdiFinance" :to="{ name: 'billing' }" exact>
          <v-list-item-title>Usage &amp; Billing</v-list-item-title>
        </v-list-item>
        <v-list-item :prepend-icon="mdiLogout" @click.prevent="logout">
          <v-list-item-title>Logout</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </v-sheet>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'
import {
  mdiPlus,
  mdiAccountCog,
  mdiLogout,
  mdiCellphoneKey,
  mdiDownload,
  mdiFinance,
  mdiBatteryChargingHigh,
  mdiPackageUp,
  mdiPackageDown,
  mdiDotsVertical,
  mdiMagnify,
  mdiCommentTextMultipleOutline,
  mdiCircle,
} from '@mdi/js'
import type { EntitiesPhone } from '~/models/api'
import { formatPhoneNumber, formatPhoneCountry, formatHumanizeTime } from '~/plugins/filters'

const store = useAppStore()
const display = useDisplay()
const router = useRouter()
const route = useRoute()

const mdiBatteryCharging = mdiBatteryChargingHigh

const owners = computed(() =>
  store.getPhones.map((phone: EntitiesPhone) => ({
    title: formatPhoneNumber(phone.phone_number),
    value: phone.phone_number,
  })),
)

async function onOwnerChanged(newOwner: string) {
  await store.updateUser({ owner: newOwner })
  if (route.name !== 'threads') {
    store.setThreadId(null)
    await router.push({ name: 'threads' })
    return
  }
  await store.loadThreads()
}

async function toggleArchive() {
  store.toggleArchive()
  if (route.name !== 'threads') {
    store.setThreadId(null)
    await router.push({ name: 'threads' })
    return
  }
  await store.loadThreads()
}

function logout() {
  store.logout().then(() => {
    store.addNotification({ type: 'info', message: 'You have successfully logged out' })
    router.push({ name: 'index' })
  })
}
</script>
