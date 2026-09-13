<template>
  <v-container
    fluid
    class="px-0 pt-0"
    :class="{ 'fill-height': display.lgAndUp.value }"
  >
    <div class="w-100 h-100">
      <v-app-bar height="60" :density="display.mdAndDown.value ? 'compact' : 'default'">
        <v-btn icon to="/threads">
          <v-icon :icon="mdiArrowLeft" />
        </v-btn>
        <v-toolbar-title>
          <div class="py-16">Search Messages</div>
        </v-toolbar-title>
        <v-progress-linear
          :active="loading"
          :indeterminate="loading"
          absolute
          location="bottom"
        ></v-progress-linear>
      </v-app-bar>
      <v-container>
        <v-row>
          <v-col cols="12">
            <h5 class="text-h4 mb-3 mt-3">Search Messages</h5>
            <p>
              On this page, you can search all your messages by phone number,
              message type, and message status and even using the content of the
              SMS message. You will also be able to bulk delete messages and
              even export your messages in a CSV file.
            </p>
            <v-alert v-if="errorTitle" variant="tonal" prominent type="warning">
              <h6 class="text-subtitle-1 font-weight-bold">{{ errorTitle }}</h6>
            </v-alert>
          </v-col>
        </v-row>
        <v-card>
          <v-card-text class="pt-4 pb-0">
            <v-row>
              <v-col cols="4">
                <v-select
                  v-model="formOwners"
                  :error="errorMessages.has('owners')"
                  :error-messages="errorMessages.get('owners')"
                  :items="phoneNumberSelectItems"
                  multiple
                  density="compact"
                  label="Phone Numbers"
                  variant="outlined"
                ></v-select>
              </v-col>
              <v-col cols="4">
                <v-select
                  v-model="formTypes"
                  :error="errorMessages.has('types')"
                  :error-messages="errorMessages.get('types')"
                  :items="messageTypeSelectItems"
                  density="compact"
                  multiple
                  label="Message Types"
                  variant="outlined"
                ></v-select>
              </v-col>
              <v-col cols="4">
                <v-select
                  v-model="formStatuses"
                  :error="errorMessages.has('statuses')"
                  :error-messages="errorMessages.get('statuses')"
                  :items="messageStatusSelectItems"
                  density="compact"
                  multiple
                  label="Message Status"
                  variant="outlined"
                ></v-select>
              </v-col>
            </v-row>
            <v-row class="mt-n3">
              <v-col cols="8">
                <v-text-field
                  v-model="formQuery"
                  :error="errorMessages.has('query')"
                  :error-messages="errorMessages.get('query')"
                  label="Search Query"
                  variant="outlined"
                  density="compact"
                  clearable
                ></v-text-field>
              </v-col>
              <v-col cols="4">
                <div id="cloudflare-turnstile" class="d-none"></div>
                <v-btn
                  :loading="loading"
                  :disabled="loading"
                  color="primary"
                  class="py-5"
                  @click="fetchMessages(true)"
                >
                  <v-icon start :icon="mdiMagnify" />
                  Search Messages
                </v-btn>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
        <v-row>
          <v-col cols="12" class="mt-16 mb-n2 d-flex">
            <h2 class="text-h4">Search Results</h2>
            <v-dialog
              v-model="showDeleteDialog"
              :scrim-opacity="0.9"
              max-width="550"
            >
              <template #activator="{ props: dialogProps }">
                <v-btn
                  :loading="loading"
                  :disabled="loading || selectedMessages.length < 1"
                  size="small"
                  class="ml-2 mt-2"
                  color="error"
                  v-bind="dialogProps"
                >
                  <v-icon start :icon="mdiDelete" />
                  Delete messages
                </v-btn>
              </template>
              <v-card>
                <v-card-title class="text-h5 text-break">
                  Are you sure you want to delete the
                  {{ selectedMessages.length }} selected messages?
                </v-card-title>
                <v-card-text>
                  The messages will be deleted permanently from the httpSMS
                  server and cannot be recovered.
                </v-card-text>
                <v-card-actions class="pb-4">
                  <v-btn
                    color="primary"
                    :loading="loading"
                    @click="deleteMessages"
                  >
                    <v-icon start :icon="mdiDelete" />
                    Yes Delete Messages
                  </v-btn>
                  <v-spacer></v-spacer>
                  <v-btn variant="text" @click="showDeleteDialog = false"> Close </v-btn>
                </v-card-actions>
              </v-card>
            </v-dialog>
            <v-spacer></v-spacer>
            <v-btn
              :loading="loading"
              :disabled="loading || selectedMessages.length < 1"
              size="small"
              color="primary"
              class="mt-2"
              @click="exportMessages"
            >
              <v-icon start :icon="mdiExport" />
              Export to CSV
            </v-btn>
          </v-col>
          <v-col cols="12">
            <v-data-table
              v-model="selectedMessages"
              item-value="id"
              :headers="headers"
              :items="messages"
              v-model:items-per-page="itemsPerPage"
              v-model:sort-by="sortBy"
              v-model:page="page"
              :items-length="totalMessages"
              :loading="loading"
              show-select
              loading-text="Loading... Please wait"
              no-data-text="You don't have any messages yet"
              class="elevation-1"
            >
              <template #[`item.created_at`]="{ item }">
                {{ formatTimestamp(item.created_at) }}
              </template>
              <template #[`item.owner`]="{ item }">
                {{ item.owner }}
              </template>
              <template #[`item.contact`]="{ item }">
                {{ item.contact }}
              </template>
              <template #[`item.type`]="{ item }">
                <span v-if="item.type === 'call/missed'">
                  <v-icon size="small" color="error" :icon="mdiCallMissed" />
                  missed call
                </span>
                <span v-if="item.type === 'mobile-originated'">
                  <v-icon size="small" :icon="mdiCallReceived" />
                  inbound
                </span>
                <span v-if="item.type === 'mobile-terminated'">
                  <v-icon size="small" color="secondary" :icon="mdiCallMade" />
                  outbound
                </span>
              </template>
              <template #[`item.status`]="{ item }">
                <v-chip
                  v-if="item.status === 'expired'"
                  color="warning"
                  size="small"
                  variant="outlined"
                >
                  <v-icon size="small" start :icon="mdiAlert" />
                  Expired
                </v-chip>

                <v-chip
                  v-else-if="item.status === 'delivered'"
                  color="primary"
                  size="small"
                  variant="outlined"
                >
                  <v-icon size="small" start :icon="mdiCheckAll" />
                  Delivered
                </v-chip>

                <v-chip
                  v-else-if="item.status === 'received'"
                  color="success"
                  size="small"
                  variant="outlined"
                >
                  <v-icon size="small" start :icon="mdiCheckAll" />
                  Received
                </v-chip>

                <v-chip v-else-if="item.status === 'sent'" size="small" variant="outlined">
                  <v-icon size="small" start :icon="mdiCheck" />
                  Sent
                </v-chip>

                <v-chip
                  v-else-if="item.status === 'failed'"
                  color="error"
                  size="small"
                  variant="outlined"
                >
                  <v-icon size="small" start :icon="mdiAlert" />
                  Failed
                </v-chip>

                <v-chip v-else size="small" color="cyan" variant="outlined">
                  <v-icon size="small" start :icon="mdiProgressCheck" />
                  {{ formatCapitalize(item.status) }}
                </v-chip>
              </template>
              <template #[`item.content`]="{ item }">
                <pre
                  style="
                    white-space: pre-wrap;
                    max-width: 300px;
                    word-break: break-all;
                  "
                  >{{ item.content }}</pre
                >
              </template>
            </v-data-table>
          </v-col>
        </v-row>
      </v-container>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  mdiDelete,
  mdiMagnify,
  mdiArrowLeft,
  mdiCheckAll,
  mdiCheck,
  mdiCallMissed,
  mdiCallReceived,
  mdiCallMade,
  mdiExport,
  mdiProgressCheck,
  mdiAlert,
} from '@mdi/js'
import type { AxiosError } from 'axios'
import { useDisplay } from 'vuetify'
import { ErrorMessages, getErrorMessages } from '~/utils/errors'
import capitalize from '~/utils/capitalize'
import { formatTimestamp, formatCapitalize, formatPhoneNumber } from '~/plugins/filters'
import type {
  EntitiesMessage,
  EntitiesPhone,
  ResponsesUnprocessableEntity,
} from '~/models/api'
import type { SearchMessagesRequest } from '~/models/message'
import { useAppStore } from '~/stores/app'

definePageMeta({ middleware: ['auth'] })
useHead({ title: 'Search your Messages - httpSMS' })

interface Turnstile {
  ready(callback: () => void): void
  render(
    container: string | HTMLElement,
    params?: {
      sitekey: string
      action: string
      callback?: (token: string) => void
      'error-callback'?: ((error: string) => void) | undefined
    },
  ): string | null | undefined
}

const store = useAppStore()
const display = useDisplay()
const config = useRuntimeConfig()

const loading = ref(true)
const errorTitle = ref('')
const showDeleteDialog = ref(false)
const selectedMessages = ref<EntitiesMessage[]>([])
const errorMessages = ref(new ErrorMessages())
const formOwners = ref<string[]>([])
const formTypes = ref<string[]>([])
const formStatuses = ref<string[]>([])
const formQuery = ref('')
const messages = ref<EntitiesMessage[]>([])
const totalMessages = ref(-1)
const itemsPerPage = ref(100)
const page = ref(1)
const sortBy = ref([{ key: 'created_at', order: 'desc' as const }])

const headers = [
  { title: 'Created At', key: 'created_at' },
  { title: 'Owner', key: 'owner' },
  { title: 'Contact', key: 'contact' },
  { title: 'Message Type', key: 'type' },
  { title: 'Status', key: 'status' },
  { title: 'Message Content', key: 'content', sortable: false },
]

const phoneNumberSelectItems = computed(() => {
  return store.getPhones.map((phone: EntitiesPhone) => ({
    title: formatPhoneNumber(phone.phone_number),
    value: phone.phone_number,
  }))
})

const messageTypeSelectItems = [
  { title: 'Outbound', value: 'mobile-terminated' },
  { title: 'Inbound', value: 'mobile-originated' },
  { title: 'Missed Calls', value: 'call/missed' },
]

const messageStatusSelectItems = [
  { value: 'pending', title: 'Pending' },
  { value: 'sent', title: 'Sent' },
  { value: 'delivered', title: 'Delivered' },
  { value: 'failed', title: 'Failed' },
  { value: 'expired', title: 'Expired' },
  { value: 'received', title: 'Received' },
]

watch([sortBy, itemsPerPage, page], () => {
  fetchMessages()
}, { deep: true })

onMounted(async () => {
  await store.loadUser()
  await store.loadPhones()
  loading.value = false
})

function getCaptcha(): Promise<string> {
  return new Promise<string>((resolve, reject) => {
    const turnstile = (window as any).turnstile as Turnstile
    turnstile.ready(() => {
      turnstile.render('#cloudflare-turnstile', {
        sitekey: config.public.cloudflareTurnstileSiteKey as string,
        callback: (token) => {
          resolve(token)
        },
        action: 'search_messages',
        'error-callback': (error: string) => {
          reject(error)
        },
      })
    })
  })
}

function exportMessages() {
  let csvContent = 'data:text/csv;charset=utf-8,'
  csvContent +=
    'Message ID,Created At,Owner,Contact,Message Type,Status,Message Content\n'
  selectedMessages.value.forEach((message) => {
    csvContent += `${message.id},${new Date(
      message.created_at,
    ).toLocaleString()},${message.owner},${message.contact},${
      message.type
    },${message.status},${sanitizeContent(message.content)}\n`
  })

  const encodedUri = encodeURI(csvContent)
  const link = document.createElement('a')
  link.setAttribute('href', encodedUri)
  link.setAttribute(
    'download',
    `httpsms-${new Date().toJSON().slice(0, 10)}.csv`,
  )
  document.body.appendChild(link)
  link.click()

  store.addNotification({
    message: 'The selected messages have been exported successfully',
    type: 'success',
  })
}

function sanitizeContent(content: string): string {
  content = content.replaceAll('"', '""')
  return content.includes(',') ? '"' + content + '"' : content
}

function deleteMessages() {
  loading.value = true
  Promise.all(
    selectedMessages.value.map((message) =>
      store.deleteMessage(message.id),
    ),
  )
    .then(() => {
      store.addNotification({
        message: 'The selected messages have been deleted successfully',
        type: 'success',
      })
      selectedMessages.value = []
    })
    .catch(() => {
      store.addNotification({
        message: 'Error while deleting the selected messages',
        type: 'error',
      })
    })
    .finally(() => {
      loading.value = false
      showDeleteDialog.value = false
      fetchMessages()
    })
}

function fetchMessages(reset = false) {
  loading.value = true
  errorMessages.value = new ErrorMessages()
  errorTitle.value = ''

  if (reset) {
    page.value = 1
  }

  const currentSort = sortBy.value[0]

  getCaptcha()
    .then((token: string) => {
      store
        .searchMessages({
          token,
          owners: formOwners.value,
          types: formTypes.value,
          statuses: formStatuses.value,
          query: formQuery.value,
          sort_by: currentSort?.key ?? 'created_at',
          sort_descending: currentSort?.order === 'desc',
          skip: (page.value - 1) * itemsPerPage.value,
          limit: itemsPerPage.value,
        } as SearchMessagesRequest)
        .then((result: EntitiesMessage[]) => {
          messages.value = result
          totalMessages.value =
            (page.value - 1) * itemsPerPage.value + result.length
          if (result.length === itemsPerPage.value) {
            totalMessages.value = totalMessages.value + 1
          }
        })
        .catch((error: AxiosError<ResponsesUnprocessableEntity>) => {
          errorTitle.value = capitalize(
            error.response?.data?.message ??
              'Error while searching messages',
          )
          errorMessages.value = getErrorMessages(error)
        })
        .finally(() => {
          loading.value = false
        })
    })
    .catch((error: string) => {
      errorTitle.value = error
      loading.value = false
    })
}
</script>
