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
          <div class="py-16">Bulk Messages</div>
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
            <h5 class="text-h4 mb-3 mt-3">Bulk Messages</h5>
            <p>
              Fill in our bulk SMS
              <a
                class="text-decoration-none"
                download
                href="/templates/httpsms-bulk.csv"
                >CSV template</a
              >
              or our
              <a
                class="text-decoration-none"
                download
                href="/templates/httpsms-bulk.xlsx"
                >Excel template</a
              >
              and upload it here to send your SMS messages to multiple
              recipients at once.
            </p>
            <v-alert v-if="errorTitle" variant="tonal" prominent type="warning">
              <h6 class="text-subtitle-1 font-weight-bold">{{ errorTitle }}</h6>
              <ul class="text-body-2">
                <li
                  v-for="message in errorMessages.get('document')"
                  :key="message"
                >
                  {{ message }}
                </li>
              </ul>
            </v-alert>
            <v-form @submit.prevent="sendBulkMessages">
              <v-file-input
                v-model="formFile"
                label="File"
                :prepend-icon="undefined"
                accept=".csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                :error-messages="errorMessages.get('document')"
                persistent-placeholder
                placeholder="Click here to upload your bulk SMS file."
                :append-inner-icon="mdiMicrosoftExcel"
                variant="outlined"
              ></v-file-input>
              <div class="d-flex">
                <v-btn
                  color="primary"
                  type="submit"
                  :loading="loading"
                  :disabled="loading"
                  size="large"
                >
                  <v-icon start :icon="mdiSendCheck" />
                  Send Bulk Messages
                </v-btn>
                <v-spacer></v-spacer>
                <v-btn
                  v-if="display.mdAndUp.value"
                  variant="plain"
                  color="info"
                  href="mailto:arnold@httpsms.com?subject=I'm having trouble with the bulk messages"
                >
                  I Need Help
                </v-btn>
              </div>
            </v-form>
          </v-col>
        </v-row>
      </v-container>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { mdiArrowLeft, mdiMicrosoftExcel, mdiSendCheck } from '@mdi/js'
import { useDisplay } from 'vuetify'
import type { AxiosError } from 'axios'
import { ErrorMessages, getErrorMessages } from '~/utils/errors'
import capitalize from '~/utils/capitalize'
import type { ResponsesUnprocessableEntity } from '~/models/api'
import { useAppStore } from '~/stores/app'

definePageMeta({ middleware: ['auth'] })
useHead({ title: 'Send Bulk Messages - httpSMS' })

const store = useAppStore()
const router = useRouter()
const display = useDisplay()

const formFile = ref<File[] | null>(null)
const loading = ref(true)
const errorTitle = ref('')
const errorMessages = ref(new ErrorMessages())

onMounted(async () => {
  await store.loadUser()
  loading.value = false
})

function sendBulkMessages() {
  loading.value = true
  errorMessages.value = new ErrorMessages()
  errorTitle.value = ''

  store
    .sendBulkMessages(formFile.value)
    .then(() => {
      setTimeout(() => {
        loading.value = false
        router.push({ name: 'threads' })
      }, 2000)
    })
    .catch((error: AxiosError<ResponsesUnprocessableEntity>) => {
      errorTitle.value = capitalize(
        error.response?.data?.message ?? 'Error while sending bulk messages',
      )
      errorMessages.value = getErrorMessages(error)
      loading.value = false
    })
}
</script>
