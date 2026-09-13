<template>
  <v-container fluid class="pa-0" :style="{ height: display.lgAndUp.value ? '100vh' : 'auto' }">
    <div class="w-full h-full">
      <v-app-bar height="60" :density="display.mdAndDown.value ? 'compact' : 'default'" location="top">
        <v-btn :icon="mdiArrowLeft" to="/threads" />
        <v-toolbar-title>
          New Message
          <v-icon size="x-small" class="mx-2" color="primary" :icon="mdiCircle" />
          {{ store.getOwner ? formatPhoneNumber(store.getOwner) : '' }}
        </v-toolbar-title>
      </v-app-bar>
      <v-container class="mt-16">
        <v-row>
          <v-col cols="12" md="8" offset-md="2" xl="6" offset-xl="3">
            <v-form @submit.prevent="sendMessage">
              <v-text-field
                v-model="formPhoneNumber"
                :disabled="sending"
                :error="errors.has('to')"
                :error-messages="errors.get('to')"
                variant="outlined"
                placeholder="Recipient phone number e.g +18005550199"
                label="Phone Number"
              />
              <v-textarea
                v-model="formContent"
                :error="errors.has('content')"
                :error-messages="errors.get('content')"
                :disabled="sending"
                variant="outlined"
                placeholder="Enter your message here"
                label="Content"
              />
              <v-btn
                type="submit"
                color="primary"
                :disabled="sending"
                :block="display.mdAndDown.value"
              >
                <v-icon :icon="mdiSend" />
                Send Message
              </v-btn>
            </v-form>
          </v-col>
        </v-row>
      </v-container>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'
import { mdiArrowLeft, mdiSend, mdiCircle } from '@mdi/js'
import { getAxios } from '~/plugins/axios'
import { ErrorMessages } from '~/utils/errors'
import { formatPhoneNumber } from '~/plugins/filters'

definePageMeta({ middleware: ['auth'] })
useHead({ title: 'New Message - Http SMS' })

const store = useAppStore()
const router = useRouter()
const display = useDisplay()

const sending = ref(false)
const formPhoneNumber = ref('')
const formContent = ref('')
const errors = ref(new ErrorMessages())

async function sendMessage() {
  errors.value = new ErrorMessages()
  sending.value = true
  const axios = getAxios()
  axios
    .post('/v1/messages/send', {
      to: formPhoneNumber.value,
      from: store.getOwner,
      content: formContent.value,
      sim: 'DEFAULT',
    })
    .then(() => {
      store.addNotification({ message: 'Message Sent!', type: 'success' })
      router.push({ name: 'threads' })
    })
    .catch((axiosError: any) => {
      const newErrors = new ErrorMessages()
      const response = axiosError.response
      if (response.data.data.content) newErrors.addMany('content', response.data.data.content)
      if (response.data.data.to) {
        newErrors.addMany('to', response.data.data.to.map((x: string) =>
          x.replace('to field', 'phone number field'),
        ))
      }
      if (response.data.data.from) {
        store.addNotification({ message: response.data.data.from[0], type: 'error' })
      }
      errors.value = newErrors
    })
    .finally(() => { sending.value = false })
}
</script>
