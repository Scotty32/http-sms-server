<template>
  <v-container
    fluid
    class="px-0 pt-0 pb-0"
    :style="{ height: display.lgAndUp.value ? '100vh' : 'auto' }"
  >
    <div class="w-full h-full">
      <v-app-bar height="60" :density="display.mdAndDown.value ? 'compact' : 'default'">
        <v-btn v-if="display.mdAndDown.value" :icon="mdiArrowLeft" to="/threads" />
        <v-toolbar-title>
          <span v-if="store.hasThread">
            {{ formatPhoneNumber(store.getThread.contact) }}
          </span>
        </v-toolbar-title>
        <v-spacer />
        <v-menu>
          <template #activator="{ props }">
            <v-btn :icon="mdiDotsVertical" variant="text" class="mt-2" v-bind="props" />
          </template>
          <v-list class="px-2" nav :density="display.mdAndDown.value ? 'compact' : 'default'">
            <v-list-item
              v-if="store.hasThread && !store.getThread.is_archived"
              :prepend-icon="mdiPackageDown"
              @click.prevent="archiveThread"
            >
              <v-list-item-title>Archive</v-list-item-title>
            </v-list-item>
            <v-list-item
              v-if="store.hasThread && store.getThread.is_archived"
              :prepend-icon="mdiPackageUp"
              @click.prevent="unArchiveThread"
            >
              <v-list-item-title>Unarchive</v-list-item-title>
            </v-list-item>
            <v-list-item
              v-if="store.hasThread"
              :prepend-icon="mdiDelete"
              base-color="error"
              @click.prevent="deleteThread(store.getThread.id)"
            >
              <v-list-item-title>Delete Thread</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </v-app-bar>
      <v-progress-linear v-if="loadingMessages" color="primary" indeterminate />
      <v-container v-if="store.hasThread">
        <div
          ref="messageBody"
          class="messages-body no-scrollbar"
          :class="{ 'pr-7': display.lgAndUp.value }"
        >
          <v-row
            v-for="message in messages"
            :key="message.id"
            :style="{ visibility: messageVisibility }"
          >
            <v-col
              class="d-flex"
              :class="{
                'pr-12': display.mdAndDown.value && !isMT(message),
                'pl-12 pr-8': display.mdAndDown.value && isMT(message),
                'pl-16 ml-16': display.lgAndUp.value && isMT(message),
                'pr-16 mr-16': display.lgAndUp.value && !isMT(message),
              }"
            >
              <v-spacer v-if="isMT(message)" />
              <v-avatar v-if="isMo(message)" :color="store.getThread.color">
                <v-icon :icon="mdiAccount" />
              </v-avatar>
              <v-avatar v-if="isMissedCall(message)" color="#1e1e1e">
                <v-icon size="large" color="red" :icon="mdiCallMissed" />
              </v-avatar>
              <v-menu v-if="isMT(message)">
                <template #activator="{ props }">
                  <v-btn :icon="mdiDotsVertical" variant="text" class="mt-2" v-bind="props" />
                </template>
                <v-list class="px-2" nav density="compact">
                  <v-list-item
                    v-if="canResend(message)"
                    :prepend-icon="mdiRefresh"
                    @click.prevent="resendMessage(message)"
                  >
                    <v-list-item-title>Resend Message</v-list-item-title>
                  </v-list-item>
                  <v-list-item :prepend-icon="mdiContentCopy" @click.prevent="copyMessageId(message)">
                    <v-list-item-title>Copy Message ID</v-list-item-title>
                  </v-list-item>
                  <v-list-item :prepend-icon="mdiDelete" base-color="error" @click.prevent="deleteMessageItem(message)">
                    <v-list-item-title>Delete Message</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
              <div>
                <v-card class="ml-2" :color="isMT(message) ? 'primary' : undefined">
                  <v-card-text class="text-high-emphasis text-break" style="white-space: pre-line">
                    <span v-if="!isMissedCall(message)">{{ message.content }}</span>
                    <span v-else class="text-medium-emphasis">Missed phone call</span>
                  </v-card-text>
                </v-card>
                <div class="d-flex">
                  <p class="ml-2 text-medium-emphasis text-caption mr-2">
                    {{ new Date(message.order_timestamp).toLocaleString() }}
                  </p>
                  <v-spacer />
                  <v-tooltip location="bottom">
                    <template #activator="{ props }">
                      <div v-bind="props">
                        <v-icon v-if="message.status === 'expired'" color="warning" class="mt-n2" :icon="mdiAlert" />
                        <v-progress-circular
                          v-else-if="isPending(message)"
                          indeterminate
                          :size="14"
                          :width="1"
                          class="mt-n2"
                          :color="statusColor(message)"
                        />
                        <v-icon v-else-if="message.status === 'delivered'" color="primary" class="mt-n6" :icon="mdiCheckAll" />
                        <v-icon v-else-if="message.status === 'sent'" class="mt-n6" :icon="mdiCheck" />
                        <v-icon v-else-if="message.status === 'failed'" color="error" class="mt-n2" :icon="mdiAlert" />
                      </div>
                    </template>
                    <span>{{ message.failure_reason ? message.failure_reason : message.status }}</span>
                  </v-tooltip>
                </div>
              </div>
              <v-menu v-if="!isMT(message)">
                <template #activator="{ props }">
                  <v-btn :icon="mdiDotsVertical" variant="text" class="mt-2" v-bind="props" />
                </template>
                <v-list class="px-2" nav density="compact">
                  <v-list-item
                    v-if="canResend(message)"
                    :prepend-icon="mdiRefresh"
                    @click.prevent="resendMessage(message)"
                  >
                    <v-list-item-title>Resend Message</v-list-item-title>
                  </v-list-item>
                  <v-list-item :prepend-icon="mdiContentCopy" @click.prevent="copyMessageId(message)">
                    <v-list-item-title>Copy Message ID</v-list-item-title>
                  </v-list-item>
                  <v-list-item :prepend-icon="mdiDelete" base-color="error" @click.prevent="deleteMessageItem(message)">
                    <v-list-item-title>Delete Message</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </v-col>
          </v-row>
        </div>
        <v-footer absolute color="#121212">
          <v-container class="pb-0">
            <v-form ref="form" class="d-flex" @submit.prevent="sendMessageHandler">
              <v-text-field
                ref="messageInput"
                v-model="formMessage"
                :disabled="submitting || !contactIsPhoneNumber"
                :rows="1"
                variant="filled"
                class="no-scrollbar ml-2"
                :rules="formMessageRules"
                :placeholder="
                  contactIsPhoneNumber
                    ? 'Type your message here'
                    : 'You cannot send messages to ' + contact
                "
                rounded
                @keydown.enter="sendMessageHandler"
              />
              <v-btn
                :disabled="submitting || !contactIsPhoneNumber"
                type="submit"
                color="primary"
                class="text-white ml-2"
                icon
              >
                <v-progress-circular
                  v-if="submitting"
                  indeterminate
                  style="position: absolute"
                  :size="20"
                  :width="3"
                  color="pink"
                />
                <v-icon :icon="mdiSend" />
              </v-btn>
            </v-form>
          </v-container>
        </v-footer>
      </v-container>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'
import { isValidPhoneNumber } from 'libphonenumber-js'
import {
  mdiSend, mdiDotsVertical, mdiArrowLeft, mdiCheckAll, mdiDelete,
  mdiCallMissed, mdiCheck, mdiAlert, mdiPackageUp, mdiPackageDown,
  mdiAccount, mdiRefresh, mdiContentCopy,
} from '@mdi/js'
import Pusher, { type Channel } from 'pusher-js'
import type { Message } from '~/models/message'
import type { NotificationRequest, SendMessageRequest, SIM } from '~/stores/app'
import { formatPhoneNumber } from '~/plugins/filters'

definePageMeta({ middleware: ['auth'] })
useHead({ title: 'Messages - httpSMS' })

const store = useAppStore()
const router = useRouter()
const route = useRoute()
const display = useDisplay()
const config = useRuntimeConfig()

const messageBody = ref<HTMLElement | null>(null)
const form = ref<any>(null)
const hideMessages = ref(true)
const loadingMessages = ref(false)
const messages = ref<Message[]>([])
const formMessage = ref('')
const submitting = ref(false)
let webhookChannel: Channel | null = null

const formMessageRules = [
  (v: string) =>
    v === '' ||
    (v && v.length <= 320) ||
    'Message must be less than 320 characters',
]

const contactIsPhoneNumber = computed(() => {
  if (!store.hasThread) return false
  return (
    isValidPhoneNumber(store.getThread.contact) ||
    !isNaN(Number(store.getThread.contact))
  )
})

const messageVisibility = computed(() => (hideMessages.value ? 'hidden' : 'visible'))
const contact = computed(() => (store.hasThread ? store.getThread.contact : ''))

onMounted(async () => {
  await loadData()

  const pusher = new Pusher(config.public.pusherKey as string, {
    cluster: config.public.pusherCluster as string,
  })
  webhookChannel = pusher.subscribe(store.getAuthUser!.id)
  webhookChannel.bind('message.phone.sent', () => { if (!loadingMessages.value) loadMessages(false) })
  webhookChannel.bind('message.send.failed', () => { if (!loadingMessages.value) loadMessages(false) })
  webhookChannel.bind('message.phone.received', () => { if (!loadingMessages.value) loadMessages(false) })
})

onBeforeUnmount(() => {
  webhookChannel?.unsubscribe()
})

function isPending(message: Message): boolean {
  return ['sending', 'pending', 'scheduled'].includes(message.status)
}

function statusColor(message: Message): string {
  if (message.status === 'sending') return 'warning'
  if (message.status === 'scheduled') return 'teal'
  return 'primary'
}

function canResend(message: Message): boolean {
  return isMT(message) && (message.status === 'expired' || message.status === 'failed')
}

function isMT(message: Message): boolean { return message.type === 'mobile-terminated' }
function isMo(message: Message): boolean { return message.type === 'mobile-originated' }
function isMissedCall(message: Message): boolean { return message.type === 'call/missed' }

function loadMessages(hideMsgs = true) {
  loadingMessages.value = true
  store
    .loadThreadMessages(route.params.id as string)
    .then((msgs: Message[]) => {
      messages.value = [...msgs].reverse()
    })
    .finally(() => {
      setTimeout(() => { loadingMessages.value = false }, 1100)
    })
  hideMessages.value = hideMsgs
  setTimeout(() => scrollToElement(), 950)
}

async function loadData() {
  await store.loadUser()
  await store.loadPhones()
  await store.loadThreads()

  if (!store.hasThreadId(route.params.id as string)) {
    await router.push({ name: 'threads' })
    return
  }
  loadMessages()
}

function scrollToElement() {
  const el = messageBody.value
  if (el) el.scrollTop = el.scrollHeight + 120
  hideMessages.value = false
}

function archiveThread() {
  store.updateThread({ threadId: store.getThread.id, isArchived: true })
}

function unArchiveThread() {
  store.updateThread({ threadId: store.getThread.id, isArchived: false })
}

async function resendMessage(message: Message) {
  await store.sendMessage({ from: message.owner, to: message.contact, content: message.content, sim: 'DEFAULT' })
  loadMessages(false)
}

async function deleteMessageItem(message: Message) {
  await store.deleteMessage(message.id)
  loadMessages(false)
}

async function copyMessageId(message: Message) {
  await navigator.clipboard.writeText(message.id)
  store.addNotification({ message: 'Message ID copied to clipboard', type: 'success' } as NotificationRequest)
}

async function deleteThread(threadID: string) {
  await store.deleteThread(threadID)
  await router.push({ name: 'threads' })
}

async function sendMessageHandler(event: KeyboardEvent) {
  if (event.shiftKey) return
  if (!form.value?.validate()) return

  submitting.value = true
  const request: SendMessageRequest = {
    from: store.getOwner!,
    to: store.getThread.contact,
    content: formMessage.value,
    sim: 'DEFAULT' as SIM,
  }
  await store.sendMessage(request)
  loadMessages(false)
  formMessage.value = ''
  submitting.value = false
}
</script>

<style lang="scss">
.messages-body {
  padding-top: 50px;
  max-height: calc(100vh - 200px);
  position: absolute;
  width: 100%;
  bottom: 120px;
}
@media (min-width: 960px) { .messages-body { max-width: 900px; } }
@media (min-width: 1264px) { .messages-body { max-width: 1185px; } }
@media (min-width: 1904px) { .messages-body { max-width: 1785px; } }

.no-scrollbar,
.no-scrollbar textarea {
  overflow-x: hidden;
  -ms-overflow-style: none;
  overflow-y: scroll;
  &::-webkit-scrollbar { display: none; }
}
</style>
