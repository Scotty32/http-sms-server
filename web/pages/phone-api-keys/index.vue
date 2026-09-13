<template>
  <v-container
    fluid
    class="px-0 pt-0"
    :style="{ height: display.lgAndUp.value ? '100vh' : 'auto' }"
  >
    <div class="w-full h-full">
      <v-app-bar height="60" :density="display.mdAndDown.value ? 'compact' : 'default'">
        <v-btn :icon="mdiArrowLeft" to="/threads" />
        <v-toolbar-title>Phone API Keys</v-toolbar-title>
        <v-progress-linear :active="loading" :indeterminate="loading" absolute location="bottom" />
      </v-app-bar>
      <v-container>
        <v-row>
          <v-col cols="12" md="9" offset-md="1" xl="8" offset-xl="2">
            <div class="d-flex mt-3 mb-4">
              <v-progress-circular
                v-if="loading"
                :size="24"
                :width="2"
                color="primary"
                class="mt-2 mr-2"
                indeterminate
              />
              <h5 class="text-h4">Phone API Keys</h5>
              <v-btn color="primary" class="ml-4 mt-1" @click="showCreateAPIKeyDialog = true">
                <v-icon :icon="mdiPlus" start />
                Create API Key
              </v-btn>
              <v-dialog v-model="showCreateAPIKeyDialog" max-width="600px">
                <v-card>
                  <v-card-title>Create Phone API Key</v-card-title>
                  <v-card-subtitle class="mt-2">
                    After creating the API key you can use it to login to the httpSMS Android app
                  </v-card-subtitle>
                  <v-card-text>
                    <v-text-field
                      v-model="formPhoneApiKeyName"
                      label="Name"
                      placeholder="Enter a name for your API key"
                      variant="outlined"
                      :disabled="loading"
                      :error="errorMessages.has('name')"
                      :error-messages="errorMessages.get('name')"
                      persistent-hint
                      class="mb-n2"
                    />
                  </v-card-text>
                  <v-card-actions class="mt-n4">
                    <v-btn color="primary" :loading="loading" @click="createPhoneApiKey">
                      Create Key
                    </v-btn>
                    <v-spacer />
                    <v-btn variant="text" @click="showCreateAPIKeyDialog = false">Close</v-btn>
                  </v-card-actions>
                </v-card>
              </v-dialog>
              <v-spacer />
              <v-btn
                v-if="display.lgAndUp.value"
                href="https://docs.httpsms.com/features/phone-api-keys"
                color="secondary"
                class="mt-1"
              >
                Documentation
              </v-btn>
            </div>
            <v-table class="mb-4 api-key-table">
              <thead>
                <tr class="text-uppercase text-subtitle-2">
                  <th class="text-left">Name</th>
                  <th class="text-left">Created At</th>
                  <th class="text-left">Phone Numbers</th>
                  <th class="text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="phoneApiKey in phoneApiKeys" :key="phoneApiKey.id">
                  <td>{{ phoneApiKey.name }}</td>
                  <td>{{ formatTimestamp(phoneApiKey.created_at) }}</td>
                  <td>
                    <ul v-if="phoneApiKey.phone_numbers.length" class="ml-n3">
                      <li
                        v-for="phoneNumber in phoneApiKey.phone_numbers"
                        :key="phoneNumber"
                        class="my-3"
                      >
                        <b>{{ formatPhoneNumber(phoneNumber) }}</b>
                        <v-btn
                          class="ml-2 mt-n1"
                          size="small"
                          color="error"
                          @click="showRemovePhoneFromApiKeyDialog(phoneApiKey, phoneNumber)"
                        >
                          Remove
                        </v-btn>
                      </li>
                    </ul>
                    <span v-else class="text-medium-emphasis">-</span>
                  </td>
                  <td>
                    <v-btn
                      size="small"
                      color="primary"
                      :disabled="loading"
                      :loading="loading"
                      @click="showPhoneApiKey(phoneApiKey)"
                    >
                      <v-icon :icon="mdiEye" start /> View
                    </v-btn>
                    <v-btn
                      class="ml-2"
                      size="small"
                      :disabled="loading"
                      color="error"
                      @click="showDeletePhoneApiKeyDialog(phoneApiKey)"
                    >
                      <v-icon :icon="mdiDelete" start /> Delete
                    </v-btn>
                  </td>
                </tr>
              </tbody>
            </v-table>
          </v-col>
        </v-row>
      </v-container>
    </div>
    <!-- View QR Code dialog -->
    <v-dialog v-model="showPhoneApiKeyQrCode" max-width="600">
      <v-card>
        <v-card-title>Phone API Key QR Code</v-card-title>
        <v-card-subtitle class="mt-2">
          Scan this QR code with the
          <a class="text-decoration-none" :href="store.getAppData.appDownloadUrl">httpSMS app</a>
          on your Android phone to login.
        </v-card-subtitle>
        <v-card-text class="text-center">
          <v-text-field
            :model-value="activePhoneApiKey?.api_key"
            readonly
            variant="outlined"
            class="mb-n2"
          />
          <canvas ref="qrCodeCanvas" />
        </v-card-text>
        <v-card-actions>
          <CopyButton
            :value="activePhoneApiKey?.api_key"
            color="primary"
            copy-text="Copy API key"
            notification-text="Phone API Key copied successfully"
          />
          <v-spacer />
          <v-btn variant="text" class="mb-4" @click="showPhoneApiKeyQrCode = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <!-- Delete dialog -->
    <v-dialog v-model="deleteApiKeyDialog" max-width="600">
      <v-card>
        <v-card-title class="text-h5 text-break">
          Are you sure you want to delete the
          <code>{{ activePhoneApiKey?.name }}</code> API Key?
        </v-card-title>
        <v-card-text>
          You will have to logout and login again on the <b>httpSMS</b> Android app on all phones using this API key.
        </v-card-text>
        <v-card-actions class="pb-4">
          <v-btn color="error" :loading="loading" @click="deleteApiKey">
            <v-icon :icon="mdiDelete" start /> Delete API Key
          </v-btn>
          <v-spacer />
          <v-btn variant="text" @click="deleteApiKeyDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <!-- Remove phone dialog -->
    <v-dialog v-model="removePhoneFromApiKeyDialog" max-width="600">
      <v-card>
        <v-card-title class="text-h5 text-break">
          Are you sure you want to remove this phone number from the Phone API Key?
        </v-card-title>
        <v-card-text>
          This will remove <code>{{ formatPhoneNumber(activePhoneNumber) }}</code> from your phone API key.
        </v-card-text>
        <v-card-actions class="pb-4">
          <v-btn color="error" :loading="loading" @click="removePhoneFromPhoneKey">
            <v-icon :icon="mdiDelete" start /> Remove Phone from key
          </v-btn>
          <v-spacer />
          <v-btn variant="text" @click="removePhoneFromApiKeyDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'
import { mdiArrowLeft, mdiPlus, mdiDelete, mdiEye } from '@mdi/js'
import * as QRCode from 'qrcode'
import type { AxiosError } from 'axios'
import { ErrorMessages, getErrorMessages } from '~/utils/errors'
import type { EntitiesPhone, EntitiesPhoneAPIKey, ResponsesUnprocessableEntity } from '~/models/api'
import { formatPhoneNumber, formatTimestamp } from '~/plugins/filters'

definePageMeta({ middleware: ['auth'] })
useHead({ title: 'Phone API keys - httpSMS' })

const store = useAppStore()
const display = useDisplay()

const loading = ref(true)
const formPhoneApiKeyName = ref('')
const showPhoneApiKeyQrCode = ref(false)
const errorMessages = ref(new ErrorMessages())
const phoneApiKeys = ref<EntitiesPhoneAPIKey[]>([])
const activePhoneApiKey = ref<EntitiesPhoneAPIKey | null>(null)
const activePhoneNumber = ref('')
const showCreateAPIKeyDialog = ref(false)
const deleteApiKeyDialog = ref(false)
const removePhoneFromApiKeyDialog = ref(false)
const qrCodeCanvas = ref<HTMLCanvasElement | null>(null)

onMounted(async () => {
  await store.loadUser()
  await store.loadPhones()
  loadPhoneApiKeys()
  loading.value = false
})

function showPhoneApiKey(apiKey: EntitiesPhoneAPIKey) {
  activePhoneApiKey.value = apiKey
  showPhoneApiKeyQrCode.value = true
  nextTick(() => generateQrCode(apiKey.api_key))
}

function showDeletePhoneApiKeyDialog(apiKey: EntitiesPhoneAPIKey) {
  activePhoneApiKey.value = apiKey
  deleteApiKeyDialog.value = true
}

function showRemovePhoneFromApiKeyDialog(apiKey: EntitiesPhoneAPIKey, phoneNumber: string) {
  activePhoneNumber.value = phoneNumber
  activePhoneApiKey.value = apiKey
  removePhoneFromApiKeyDialog.value = true
}

function generateQrCode(text: string) {
  if (qrCodeCanvas.value) {
    QRCode.toCanvas(qrCodeCanvas.value, text, { errorCorrectionLevel: 'H' }, (err: any) => {
      if (err) store.addNotification({ message: 'Failed to generate phone API key QR code', type: 'error' })
    })
  }
}

function loadPhoneApiKeys() {
  loading.value = true
  store
    .indexPhoneApiKeys()
    .then((keys: EntitiesPhoneAPIKey[]) => { phoneApiKeys.value = keys })
    .finally(() => { loading.value = false })
}

function removePhoneFromPhoneKey() {
  loading.value = true
  store
    .deletePhoneFromPhoneApiKey({
      phoneApiKeyId: activePhoneApiKey.value?.id ?? '',
      phoneId: store.getPhones.find(
        (phone: EntitiesPhone) => phone.phone_number === activePhoneNumber.value,
      )?.id ?? '',
    })
    .then(() => {
      removePhoneFromApiKeyDialog.value = false
      loadPhoneApiKeys()
    })
    .finally(() => { loading.value = false })
}

function deleteApiKey() {
  loading.value = true
  store
    .deletePhoneApiKey(activePhoneApiKey.value?.id ?? '')
    .then(() => {
      deleteApiKeyDialog.value = false
      loadPhoneApiKeys()
    })
    .finally(() => { loading.value = false })
}

function createPhoneApiKey() {
  errorMessages.value = new ErrorMessages()
  loading.value = true
  store
    .storePhoneApiKey(formPhoneApiKeyName.value)
    .then(() => {
      formPhoneApiKeyName.value = ''
      showCreateAPIKeyDialog.value = false
      loadPhoneApiKeys()
    })
    .catch((error: AxiosError<ResponsesUnprocessableEntity>) => {
      errorMessages.value = getErrorMessages(error)
      loading.value = false
    })
    .finally(() => { loading.value = false })
}
</script>

<style scoped lang="scss">
.api-key-table {
  tbody tr:hover {
    background-color: transparent !important;
  }
}
</style>
