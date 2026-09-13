import { defineStore } from 'pinia'
import type { AxiosError, AxiosResponse } from 'axios'
import { getAxios, setApiKey, setAuthHeader } from '~/plugins/axios'
import type { MessageThread } from '~/models/message-thread'
import type { Message, SearchMessagesRequest } from '~/models/message'
import type { Heartbeat } from '~/models/heartbeat'
import type { User } from '~/models/user'
import type { BillingUsage } from '~/models/billing'
import type {
  EntitiesDiscord,
  EntitiesMessage,
  EntitiesPhone,
  EntitiesPhoneAPIKey,
  EntitiesUser,
  EntitiesWebhook,
  RequestsDiscordStore,
  RequestsDiscordUpdate,
  RequestsUserNotificationUpdate,
  RequestsUserPaymentInvoice,
  RequestsWebhookStore,
  RequestsWebhookUpdate,
  ResponsesDiscordResponse,
  ResponsesDiscordsResponse,
  ResponsesMessagesResponse,
  ResponsesNoContent,
  ResponsesOkString,
  ResponsesPhoneAPIKeyResponse,
  ResponsesPhoneAPIKeysResponse,
  ResponsesUnprocessableEntity,
  ResponsesUserResponse,
  ResponsesUserSubscriptionPaymentsResponse,
  ResponsesWebhookResponse,
  ResponsesWebhooksResponse,
} from '~/models/api'
import { getErrorMessages } from '~/utils/errors'

const defaultNotificationTimeout = 3000
const authTokenStorageKey = 'httpsms_auth_token'

export type NotificationType = 'error' | 'success' | 'info'

export interface Notification {
  message: string
  timeout: number
  active: boolean
  type: NotificationType
}

export interface NotificationRequest {
  message: string
  type: NotificationType
}

export type AuthUser = {
  email: string | null
  displayName: string | null
  id: string
}

export type AppData = {
  url: string
  name: string
  env: string
  appDownloadUrl: string
  documentationUrl: string
  githubUrl: string
}

export type SIM = 'SIM1' | 'SIM2' | 'DEFAULT'

export type SendMessageRequest = {
  from: string
  to: string
  content: string
  sim: SIM
}

export const useAppStore = defineStore('app', () => {
  // ─── State ────────────────────────────────────────────────────────
  const threads = ref<MessageThread[]>([])
  const threadId = ref<string | null>(null)
  const heartbeat = ref<Heartbeat | null>(null)
  const axiosError = ref<AxiosError | null>(null)
  const authStateChanged = ref(false)
  const loadingThreads = ref(true)
  const billingUsage = ref<BillingUsage | null>(null)
  const billingUsageHistory = ref<BillingUsage[]>([])
  const archivedThreads = ref(false)
  const pooling = ref(false)
  const phones = ref<EntitiesPhone[]>([])
  const user = ref<User | null>(null)
  const owner = ref<string | null>(null)
  const authUser = ref<AuthUser | null>(null)
  const notification = ref<Notification>({
    active: false,
    message: '',
    type: 'success',
    timeout: defaultNotificationTimeout,
  })

  // ─── Getters ──────────────────────────────────────────────────────
  const getThreads = computed(() => threads.value)
  const getAuthUser = computed(() => authUser.value)
  const getAxiosError = computed(() => axiosError.value)
  const getUser = computed(() => user.value)
  const getBillingUsage = computed(() => billingUsage.value)
  const getBillingUsageHistory = computed(() => billingUsageHistory.value)
  const getOwner = computed(() => owner.value)
  const getPhones = computed(() => phones.value)
  const getHeartbeat = computed(() => heartbeat.value)
  const getPolling = computed(() => pooling.value)
  const getIsArchived = computed(() => archivedThreads.value)
  const getNotification = computed(() => notification.value)
  const getLoadingThreads = computed(() => loadingThreads.value)
  const isAuthStateChanged = computed(() => authStateChanged.value)
  const isLocal = computed(() => {
    const config = useRuntimeConfig()
    return config.public.appEnv === 'local'
  })
  const getAppData = computed((): AppData => {
    const config = useRuntimeConfig()
    let url = config.public.appUrl as string
    if (url.length > 0 && url[url.length - 1] === '/') {
      url = url.substring(0, url.length - 1)
    }
    return {
      url,
      env: config.public.appEnv as string,
      appDownloadUrl: config.public.appDownloadUrl as string,
      documentationUrl: config.public.appDocumentationUrl as string,
      githubUrl: config.public.appGithubUrl as string,
      name: config.public.appName as string,
    }
  })

  const getActivePhone = computed(() =>
    phones.value.find((x) => x.phone_number === owner.value) ?? null,
  )

  const hasThread = computed(
    () => threadId.value != null && !loadingThreads.value,
  )

  const getThread = computed(() => {
    const thread = threads.value.find((x) => x.id === threadId.value)
    if (thread === undefined) {
      throw new Error(`cannot find thread with id ${threadId.value}`)
    }
    return thread
  })

  function hasThreadId(id: string): boolean {
    return threads.value.find((x) => x.id === id) !== undefined
  }

  // ─── Actions ──────────────────────────────────────────────────────
  async function loadThreads() {
    if (owner.value === null && phones.value.length === 0) {
      loadingThreads.value = false
      return
    }
    const axios = getAxios()
    const response = await axios.get('/v1/message-threads', {
      params: {
        owner: owner.value ?? phones.value[0].phone_number,
        limit: 100,
        is_archived: archivedThreads.value,
      },
    })
    getHeartbeatAction().catch(console.error)
    threads.value = [...response.data.data]
    loadingThreads.value = false
  }

  async function loadBillingUsage() {
    const axios = getAxios()
    const response = await axios.get('/v1/billing/usage')
    billingUsage.value = response.data.data
  }

  async function loadBillingUsageHistory() {
    const axios = getAxios()
    const response = await axios.get('/v1/billing/usage-history')
    billingUsageHistory.value = response.data.data
  }

  function toggleArchive() {
    archivedThreads.value = !archivedThreads.value
  }

  async function loadPhones(force = false) {
    if (phones.value.length > 0 && !force) return
    const axios = getAxios()
    const response = await axios.get('/v1/phones', { params: { limit: 100 } })
    phones.value = response.data.data

    const currentOwner = phones.value.find((x) => x.phone_number === owner.value)
    if (!currentOwner && phones.value.length > 0) {
      owner.value = phones.value[0].phone_number
    }

    if (user.value?.active_phone_id) {
      const phone = response.data.data.find(
        (x: EntitiesPhone) => x.id === user.value?.active_phone_id,
      )
      if (phone) {
        owner.value = phone.phone_number
        loadingThreads.value = true
      }
    }
  }

  async function loadUser() {
    const axios = getAxios()
    const response = await axios.get('/v1/users/me')
    const u: User = response.data.data
    user.value = u
    authUser.value = { id: u.id, email: u.email, displayName: u.email }
    authStateChanged.value = true
    setApiKey(u.api_key)
  }

  async function deletePhone(phoneID: string) {
    const axios = getAxios()
    await axios.delete(`/v1/phones/${phoneID}`)
    await loadPhones(true)
  }

  function resetState() {
    threads.value = []
    billingUsage.value = null
    billingUsageHistory.value = []
    phones.value = []
    user.value = null
    threadId.value = null
    archivedThreads.value = false
    pooling.value = false
    owner.value = null
    setApiKey('')
  }

  async function updatePhone(phone: EntitiesPhone) {
    const axios = getAxios()
    await axios
      .put(`/v1/phones`, {
        fcm_token: phone.fcm_token,
        sim: phone.sim,
        phone_number: phone.phone_number,
        message_expiration_seconds: parseInt(
          phone.message_expiration_seconds.toString(),
        ),
        missed_call_auto_reply: phone.missed_call_auto_reply,
        max_send_attempts: parseInt(phone.max_send_attempts.toString()),
        messages_per_minute: parseInt(phone.messages_per_minute.toString()),
      })
      .catch((error: AxiosError) => handleAxiosError(error))
      .then((response: any) => {
        addNotification({ message: response.data.message, type: 'success' })
      })
    await loadPhones(true)
  }

  function sendBulkMessages(document: File): Promise<ResponsesNoContent> {
    return new Promise<ResponsesNoContent>((resolve, reject) => {
      const formData = new FormData()
      formData.append('document', document)
      const axios = getAxios()
      axios
        .post<ResponsesNoContent>(`/v1/bulk-messages`, formData, {
          headers: { 'content-type': 'multipart/form-data' },
        })
        .then(async (response: AxiosResponse<ResponsesNoContent>) => {
          addNotification({
            message: response.data.message ?? 'Bulk messages sent successfully',
            type: 'success',
          })
          resolve(response.data)
        })
        .catch(async (error: AxiosError<ResponsesUnprocessableEntity>) => {
          addNotification({
            message:
              error.response?.data?.message ?? 'Errors while sending bulk messages',
            type: 'error',
          })
          reject(error)
        })
    })
  }

  function storePhoneApiKey(name: string): Promise<ResponsesPhoneAPIKeyResponse> {
    return new Promise<ResponsesPhoneAPIKeyResponse>((resolve, reject) => {
      const axios = getAxios()
      axios
        .post<ResponsesPhoneAPIKeyResponse>(`/v1/phone-api-keys`, { name })
        .then(async (response: AxiosResponse<ResponsesPhoneAPIKeyResponse>) => {
          addNotification({
            message: response.data.message ?? 'Phone API Key created successfully',
            type: 'success',
          })
          resolve(response.data)
        })
        .catch(async (error: AxiosError<ResponsesUnprocessableEntity>) => {
          addNotification({
            message:
              error.response?.data?.message ?? 'Errors while creating phone API key',
            type: 'error',
          })
          reject(error)
        })
    })
  }

  function indexPhoneApiKeys(): Promise<EntitiesPhoneAPIKey[]> {
    return new Promise<EntitiesPhoneAPIKey[]>((resolve, reject) => {
      const axios = getAxios()
      axios
        .get<ResponsesPhoneAPIKeysResponse>(`/v1/phone-api-keys`, {
          params: { limit: 100 },
        })
        .then((response: AxiosResponse<ResponsesPhoneAPIKeysResponse>) => {
          resolve(response.data.data)
        })
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while fetching phone API keys',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function deletePhoneApiKey(phoneAPIKeyID: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const axios = getAxios()
      axios
        .delete<ResponsesNoContent>(`/v1/phone-api-keys/${phoneAPIKeyID}`)
        .then(async (response: AxiosResponse<ResponsesNoContent>) => {
          addNotification({
            message:
              response.data.message ?? 'The phone API key has been deleted successfully',
            type: 'success',
          })
          resolve()
        })
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while deleting phone API key',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function deletePhoneFromPhoneApiKey(payload: {
    phoneApiKeyId: string
    phoneId: string
  }): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const axios = getAxios()
      axios
        .delete<ResponsesNoContent>(
          `/v1/phone-api-keys/${payload.phoneApiKeyId}/phones/${payload.phoneId}`,
        )
        .then(async (response: AxiosResponse<ResponsesNoContent>) => {
          addNotification({
            message:
              response.data.message ??
              'The phone has been removed from the phone API key successfully',
            type: 'success',
          })
          resolve()
        })
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while deleting phone API key',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function indexSubscriptionPayments(): Promise<ResponsesUserSubscriptionPaymentsResponse> {
    return new Promise<ResponsesUserSubscriptionPaymentsResponse>((resolve, reject) => {
      const axios = getAxios()
      axios
        .get<ResponsesUserSubscriptionPaymentsResponse>(
          `/v1/users/subscription/payments`,
          { params: { limit: 100 } },
        )
        .then(
          (response: AxiosResponse<ResponsesUserSubscriptionPaymentsResponse>) => {
            resolve(response.data)
          },
        )
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while fetching subscription payments.',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function generateSubscriptionPaymentInvoice(payload: {
    subscriptionInvoiceId: string
    request: RequestsUserPaymentInvoice
  }): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const axios = getAxios()
      axios
        .post(
          `/v1/users/subscription/invoices/${payload.subscriptionInvoiceId}`,
          payload.request,
          { responseType: 'blob' },
        )
        .then(async (response: AxiosResponse) => {
          const pdfBlob = new Blob([response.data], {
            type: response.headers['content-type'],
          })
          const url = window.URL.createObjectURL(pdfBlob)
          const tempLink = document.createElement('a')
          tempLink.href = url
          tempLink.setAttribute(
            'download',
            response.headers['content-disposition']
              ?.split('filename=')[1]
              .replaceAll('"', '') || 'Invoice.pdf',
          )
          document.body.appendChild(tempLink)
          tempLink.click()
          document.body.removeChild(tempLink)
          window.URL.revokeObjectURL(url)
          addNotification({
            message: response.data.message ?? 'Your invoice has been generated successfully',
            type: 'success',
          })
          resolve()
        })
        .catch(async (error: AxiosError) => {
          const text = await (error.response as any).data.text()
          if (error.response) error.response.data = JSON.parse(text)
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while generating your invoice',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  async function handleAxiosError(error: AxiosError<ResponsesUnprocessableEntity>) {
    const errorMessage = (error.response?.data as any)?.data[
      Object.keys((error.response?.data as any)?.data)[0]
    ][0]
    addNotification({
      message:
        (errorMessage ? errorMessage.replaceAll('_', ' ') : null) ??
        (error.response?.data as any)?.message,
      type: 'error',
    })
    axiosError.value = error
  }

  function getHeartbeatAction(limit = 1): Promise<Heartbeat[]> {
    return new Promise<Heartbeat[]>((resolve, reject) => {
      const axios = getAxios()
      axios
        .get('/v1/heartbeats', {
          params: { limit, owner: owner.value },
        })
        .then((response: AxiosResponse) => {
          heartbeat.value = response.data.data.length > 0 ? response.data.data[0] : null
          resolve(response.data.data)
        })
        .catch(async (error: AxiosError<ResponsesUnprocessableEntity>) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Errors while fetching heartbeat',
            type: 'error',
          })
          reject(error)
        })
    })
  }

  function setPolling(status: boolean) {
    pooling.value = status
  }

  async function sendMessage(request: SendMessageRequest) {
    const axios = getAxios()
    try {
      const response = await axios.post('/v1/messages/send', request)
      addNotification({ message: response.data.message, type: 'success' })
    } catch (e) {
      addNotification({
        message:
          ((e as AxiosError).response?.data as any)?.message ?? 'Error while sending message',
        type: 'error',
      })
    }
    await loadThreads()
  }

  function deleteMessage(messageId: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const axios = getAxios()
      axios
        .delete<ResponsesNoContent>(`/v1/messages/${messageId}`)
        .then(async () => {
          addNotification({
            message: 'The message has been deleted successfully',
            type: 'success',
          })
          resolve()
        })
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while deleting message',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function searchMessages(payload: SearchMessagesRequest): Promise<EntitiesMessage[]> {
    const token = payload.token
    delete payload.token
    return new Promise<EntitiesMessage[]>((resolve, reject) => {
      const axios = getAxios()
      axios
        .get<ResponsesMessagesResponse>(`/v1/messages/search`, {
          params: payload,
          headers: { token },
        })
        .then((response: AxiosResponse<ResponsesMessagesResponse>) => {
          resolve(response.data.data)
        })
        .catch((error: AxiosError) => reject(error))
    })
  }

  function setThreadId(id: string | null) {
    threadId.value = id
  }

  function addNotification(request: NotificationRequest) {
    notification.value = {
      ...notification.value,
      active: true,
      message: request.message,
      type: request.type,
      timeout: Math.floor(Math.random() * 100) + defaultNotificationTimeout,
    }
  }

  function disableNotification() {
    notification.value.active = false
  }

  function loadThreadMessages(id: string | null): Promise<Message[]> {
    threadId.value = id
    return new Promise<Message[]>((resolve, reject) => {
      const thread = threads.value.find((x) => x.id === id)
      if (!thread) {
        reject(new Error(`Thread ${id} not found`))
        return
      }
      const axios = getAxios()
      axios
        .get('/v1/messages', {
          params: { contact: thread.contact, owner: thread.owner, limit: 50 },
        })
        .then((response: AxiosResponse) => resolve(response.data.data))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Errors while fetching messages',
            type: 'error',
          })
          reject(error)
        })
    })
  }

  async function setAuthUserFromUser(u: User) {
    authStateChanged.value = true
    authUser.value = { id: u.id, email: u.email, displayName: u.email }
    user.value = u
    setApiKey(u.api_key)
  }

  function clearAuthUser() {
    authStateChanged.value = true
    authUser.value = null
    user.value = null
    setApiKey('')
    setAuthToken(null)
  }

  function setAuthToken(token: string | null) {
    if (process.client) {
      if (token) {
        localStorage.setItem(authTokenStorageKey, token)
      } else {
        localStorage.removeItem(authTokenStorageKey)
      }
    }
    setAuthHeader(token)
  }

  async function login(payload: { email: string; password: string }) {
    const axios = getAxios()
    const response = await axios.post('/v1/auth/login', payload)
    const u = response.data.data.user
    setAuthToken(response.data.data.token)
    await setAuthUserFromUser(u)
    await loadPhones()
  }

  async function register(payload: { name: string; email: string; password: string }) {
    const axios = getAxios()
    const response = await axios.post('/v1/auth/register', payload)
    const u = response.data.data.user
    setAuthToken(response.data.data.token)
    await setAuthUserFromUser(u)
    await loadPhones()
  }

  async function logout() {
    const axios = getAxios()
    try {
      await axios.delete('/v1/auth/sessions')
    } catch {
      // ignore errors during logout
    }
    clearAuthUser()
    resetState()
  }

  function restoreAuthToken() {
    if (!process.client) return
    const token = localStorage.getItem(authTokenStorageKey)
    if (token) {
      setAuthHeader(token)
    }
  }

  function initAuth() {
    // no-op: auth is now validated server-side via /me on protected routes
    authStateChanged.value = true
  }

  function clearAxiosError() {
    axiosError.value = null
  }

  async function updateUser(payload: { owner: string; timezone?: string }) {
    owner.value = payload.owner
    loadingThreads.value = true
    const phone = getActivePhone.value
    if (!phone) return
    const axios = getAxios()
    const response = await axios.put('/v1/users/me', {
      active_phone_id: phone.id,
      timezone: payload.timezone ?? user.value?.timezone,
    })
    setApiKey(response.data.data.api_key)
    user.value = response.data.data
  }

  function deleteUserAccount(): Promise<string> {
    return new Promise<string>((resolve, reject) => {
      const axios = getAxios()
      axios
        .delete<ResponsesNoContent>(`/v1/users/me`)
        .then((response: AxiosResponse<ResponsesNoContent>) => resolve(response.data.message))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while deleting your user account',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function updateTimezone(timezone: string): Promise<EntitiesUser> {
    return new Promise<EntitiesUser>((resolve, reject) => {
      const axios = getAxios()
      axios
        .put<ResponsesUserResponse>(`/v1/users/me`, {
          timezone: timezone ?? user.value?.timezone,
        })
        .then((response: AxiosResponse<ResponsesUserResponse>) => resolve(response.data.data))
        .catch((error: AxiosError) => reject(getErrorMessages(error)))
    })
  }

  function updateV2WebhookUrl(webhookUrl: string | null): Promise<EntitiesUser> {
    return new Promise<EntitiesUser>((resolve, reject) => {
      const axios = getAxios()
      axios
        .put<ResponsesUserResponse>(`/v1/users/me`, {
          timezone: user.value?.timezone,
          webhook_url: webhookUrl,
        })
        .then((response: AxiosResponse<ResponsesUserResponse>) => {
          user.value = response.data.data
          resolve(response.data.data)
        })
        .catch((error: AxiosError) => reject(getErrorMessages(error)))
    })
  }

  function testV2WebhookUrl(webhookUrl: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const axios = getAxios()
      axios
        .post(`/v1/users/me/webhook-url/test`, {
          webhook_url: webhookUrl,
        })
        .then(() => resolve())
        .catch((error: AxiosError) => {
          reject(
            (error.response?.data as any)?.message ??
              'Webhook URL test failed',
          )
        })
    })
  }

  function rotateWebhookSecret(): Promise<EntitiesUser> {
    return new Promise<EntitiesUser>((resolve, reject) => {
      const axios = getAxios()
      axios
        .post<ResponsesUserResponse>(`/v1/users/me/webhook-secret/rotate`)
        .then((response: AxiosResponse<ResponsesUserResponse>) => {
          user.value = response.data.data
          resolve(response.data.data)
        })
        .catch((error: AxiosError) => reject(getErrorMessages(error)))
    })
  }

  async function updateThread(payload: { threadId: string; isArchived: boolean }) {
    const axios = getAxios()
    await axios.put(`/v1/message-threads/${payload.threadId}`, {
      is_archived: payload.isArchived,
    })
    archivedThreads.value = payload.isArchived
    await loadThreads()
  }

  function deleteThread(id: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const axios = getAxios()
      axios
        .delete<ResponsesNoContent>(`/v1/message-threads/${id}`)
        .then(async () => {
          threadId.value = null
          addNotification({
            message: 'The message thread has been deleted successfully',
            type: 'success',
          })
          resolve()
        })
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while deleting message thread',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function getSubscriptionUpdateLink(): Promise<string> {
    return new Promise<string>((resolve, reject) => {
      const axios = getAxios()
      axios
        .get<ResponsesOkString>(`/v1/users/subscription-update-url`)
        .then((response: AxiosResponse<ResponsesOkString>) => resolve(response.data.data))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while fetching the update URL',
            type: 'error',
          })
          reject(error)
        })
    })
  }

  function cancelSubscription(): Promise<string> {
    return new Promise<string>((resolve, reject) => {
      const axios = getAxios()
      axios
        .delete<ResponsesNoContent>(`/v1/users/subscription`)
        .then((response: AxiosResponse<ResponsesNoContent>) => resolve(response.data.message))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while cancelling your subscription',
            type: 'error',
          })
          reject(error)
        })
    })
  }

  function createDiscord(payload: RequestsDiscordStore): Promise<EntitiesDiscord> {
    return new Promise<EntitiesDiscord>((resolve, reject) => {
      const axios = getAxios()
      axios
        .post<ResponsesDiscordResponse>(`/v1/discord-integrations`, payload)
        .then((response: AxiosResponse<ResponsesDiscordResponse>) => resolve(response.data.data))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while adding discord integration',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function getDiscordIntegrations(): Promise<EntitiesDiscord[]> {
    return new Promise<EntitiesDiscord[]>((resolve, reject) => {
      const axios = getAxios()
      axios
        .get<ResponsesDiscordsResponse>(`/v1/discord-integrations`, {
          params: { limit: 100 },
        })
        .then((response: AxiosResponse<ResponsesDiscordsResponse>) => resolve(response.data.data))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ??
              'Error while fetching discord integrations',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function updateDiscordIntegration(
    payload: RequestsDiscordUpdate & { id: string },
  ): Promise<EntitiesDiscord> {
    return new Promise<EntitiesDiscord>((resolve, reject) => {
      const axios = getAxios()
      axios
        .put<ResponsesDiscordResponse>(`/v1/discord-integrations/${payload.id}`, payload)
        .then((response: AxiosResponse<ResponsesDiscordResponse>) => resolve(response.data.data))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ??
              'Error while updating discord integration',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function deleteDiscordIntegration(id: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const axios = getAxios()
      axios
        .delete<ResponsesNoContent>(`/v1/discord-integrations/${id}`)
        .then(() => resolve())
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ??
              'Error while deleting discord integration',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function createWebhook(payload: RequestsWebhookStore): Promise<EntitiesWebhook> {
    return new Promise<EntitiesWebhook>((resolve, reject) => {
      const axios = getAxios()
      axios
        .post<ResponsesWebhookResponse>(`/v1/webhooks`, payload)
        .then((response: AxiosResponse<ResponsesWebhookResponse>) => resolve(response.data.data))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while adding webhook',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function getWebhooks(): Promise<EntitiesWebhook[]> {
    return new Promise<EntitiesWebhook[]>((resolve, reject) => {
      const axios = getAxios()
      axios
        .get<ResponsesWebhooksResponse>(`/v1/webhooks`, { params: { limit: 100 } })
        .then((response: AxiosResponse<ResponsesWebhooksResponse>) => resolve(response.data.data))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while fetching webhooks',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function rotateApiKey(userId: string): Promise<EntitiesUser> {
    return new Promise<EntitiesUser>((resolve, reject) => {
      const axios = getAxios()
      axios
        .delete<ResponsesUserResponse>(`/v1/users/${userId}/api-keys`)
        .then((response: AxiosResponse<ResponsesUserResponse>) => {
          user.value = response.data.data
          setApiKey(response.data.data.api_key)
          addNotification({ message: 'API Key rotated successfully', type: 'success' })
          resolve(response.data.data)
        })
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while rotating your API key',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function updateWebhook(
    payload: RequestsWebhookUpdate & { id: string },
  ): Promise<EntitiesWebhook> {
    return new Promise<EntitiesWebhook>((resolve, reject) => {
      const axios = getAxios()
      axios
        .put<ResponsesWebhookResponse>(`/v1/webhooks/${payload.id}`, payload)
        .then((response: AxiosResponse<ResponsesWebhookResponse>) => resolve(response.data.data))
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while updating webhook',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function deleteWebhook(id: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const axios = getAxios()
      axios
        .delete<ResponsesNoContent>(`/v1/webhooks/${id}`)
        .then(() => resolve())
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ?? 'Error while deleting webhook',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  function saveEmailNotifications(
    payload: RequestsUserNotificationUpdate,
  ): Promise<EntitiesUser> {
    return new Promise<EntitiesUser>((resolve, reject) => {
      const axios = getAxios()
      axios
        .put<ResponsesUserResponse>(
          `/v1/users/${user.value?.id}/notifications`,
          payload,
        )
        .then((response: AxiosResponse<ResponsesUserResponse>) => {
          user.value = response.data.data
          resolve(response.data.data)
        })
        .catch(async (error: AxiosError) => {
          addNotification({
            message:
              (error.response?.data as any)?.message ??
              'Error while updating email notification settings',
            type: 'error',
          })
          reject(getErrorMessages(error))
        })
    })
  }

  return {
    // state (expose as readonly refs)
    threads,
    threadId,
    heartbeat,
    axiosError,
    authStateChanged,
    loadingThreads,
    billingUsage,
    billingUsageHistory,
    archivedThreads,
    pooling,
    phones,
    user,
    owner,
    authUser,
    notification,
    // getters
    getThreads,
    getAuthUser,
    getAxiosError,
    getUser,
    getBillingUsage,
    getBillingUsageHistory,
    getOwner,
    getPhones,
    getHeartbeat,
    getPolling,
    getIsArchived,
    getNotification,
    getLoadingThreads,
    isAuthStateChanged,
    isLocal,
    getAppData,
    getActivePhone,
    hasThread,
    getThread,
    hasThreadId,
    // actions
    loadThreads,
    loadBillingUsage,
    loadBillingUsageHistory,
    toggleArchive,
    loadPhones,
    loadUser,
    deletePhone,
    resetState,
    updatePhone,
    sendBulkMessages,
    storePhoneApiKey,
    indexPhoneApiKeys,
    deletePhoneApiKey,
    deletePhoneFromPhoneApiKey,
    indexSubscriptionPayments,
    generateSubscriptionPaymentInvoice,
    handleAxiosError,
    getHeartbeatAction,
    setPolling,
    sendMessage,
    deleteMessage,
    searchMessages,
    setThreadId,
    addNotification,
    disableNotification,
    loadThreadMessages,
    setAuthUser: setAuthUserFromUser,
    clearAuthUser,
    restoreAuthToken,
    login,
    register,
    logout,
    initAuth,
    clearAxiosError,
    updateUser,
    deleteUserAccount,
    updateTimezone,
    updateV2WebhookUrl,
    testV2WebhookUrl,
    rotateWebhookSecret,
    updateThread,
    deleteThread,
    getSubscriptionUpdateLink,
    cancelSubscription,
    createDiscord,
    getDiscordIntegrations,
    updateDiscordIntegration,
    deleteDiscordIntegration,
    createWebhook,
    getWebhooks,
    rotateApiKey,
    updateWebhook,
    deleteWebhook,
    saveEmailNotifications,
  }
})
