<template>
  <v-container
    fluid
    class="px-0 pt-0 fill-height-lg"
  >
    <div class="w-full h-full">
      <v-app-bar height="60" fixed :density="display.mdAndDown.value ? 'compact' : 'default'">
        <v-btn icon to="/threads">
          <v-icon :icon="mdiArrowLeft"></v-icon>
        </v-btn>
        <v-toolbar-title>
          <div class="py-16">Settings</div>
        </v-toolbar-title>
      </v-app-bar>
      <v-container class="mt-16">
        <v-row>
          <v-col cols="12" md="9" offset-md="1" xl="8" offset-xl="2">
            <div v-if="store.getAuthUser" class="text-center">
              <v-avatar size="100" color="indigo" class="mx-auto">
                <v-icon dark size="70" :icon="mdiAccountCircle"></v-icon>
              </v-avatar>
              <h4 class="text-medium-emphasis">
                {{ store.getAuthUser.email }}
              </h4>
              <v-autocomplete
                v-if="store.getUser"
                density="compact"
                variant="outlined"
                :model-value="store.getUser.timezone"
                class="mx-auto mt-2"
                style="max-width: 250px"
                label="Timezone"
                :items="timezones"
                @update:model-value="updateTimezoneHandler"
              ></v-autocomplete>
            </div>
            <h5 class="text-h4 mb-3 mt-3">API Key</h5>
            <p class="text-medium-emphasis">
              Use your API Key in the <code>x-api-key</code> HTTP Header when
              sending requests to
              <code>http://sms-dev.om-ci.org</code> endpoints.
            </p>
            <div v-if="apiKey === ''" class="mb-n9 pl-3 pt-5">
              <v-progress-circular
                :size="20"
                :width="2"
                color="primary"
                indeterminate
              ></v-progress-circular>
            </div>
            <v-text-field
              v-else
              :append-icon="apiKeyShow ? mdiEye : mdiEyeOff"
              :type="apiKeyShow ? 'text' : 'password'"
              :model-value="apiKey"
              readonly
              name="api-key"
              variant="outlined"
              class="mb-n2"
              @click:append="apiKeyShow = !apiKeyShow"
            ></v-text-field>
            <div class="d-flex flex-wrap">
              <copy-button
                :value="apiKey"
                color="primary"
                copy-text="Copy API Key"
                notification-text="API Key copied successfully"
              ></copy-button>
              <v-btn
                v-if="display.mdAndUp.value"
                color="primary"
                class="ml-4"
                @click="showQrCodeDialog = true"
              >
                <v-icon start :icon="mdiQrcode"></v-icon>
                Show QR Code
              </v-btn>
              <v-dialog
                v-model="showQrCodeDialog"
                overlay-opacity="0.9"
                max-width="400px"
              >
                <v-card>
                  <v-card-title class="justify-center"
                    >API Key QR Code</v-card-title
                  >
                  <v-card-subtitle class="mt-2 text-center"
                    >Scan this QR code with the
                    <a :href="store.getAppData.appDownloadUrl"
                      >httpSMS app</a
                    >
                    on your Android phone to login.</v-card-subtitle
                  >
                  <v-card-text class="text-center">
                    <canvas ref="qrCodeCanvasRef"></canvas>
                  </v-card-text>
                  <v-card-actions>
                    <v-btn
                      color="primary"
                      block
                      class="mb-4"
                      @click="showQrCodeDialog = false"
                      >Close</v-btn
                    >
                  </v-card-actions>
                </v-card>
              </v-dialog>
              <v-btn
                v-if="display.lgAndUp.value"
                class="ml-4"
                :href="store.getAppData.documentationUrl"
                >Documentation</v-btn
              >
              <v-spacer></v-spacer>
              <v-dialog
                v-model="showRotateApiKey"
                overlay-opacity="0.9"
                max-width="550"
              >
                <template #activator="{ props: activatorProps }">
                  <v-btn
                    :size="display.mdAndDown.value ? 'small' : 'default'"
                    :variant="display.lgAndUp.value ? 'text' : 'elevated'"
                    color="warning"
                    v-bind="activatorProps"
                  >
                    <v-icon start :icon="mdiRefresh"></v-icon>
                    Rotate API Key
                  </v-btn>
                </template>
                <v-card>
                  <v-card-title class="text-h5 text-break">
                    Are you sure you want to rotate your API Key?
                  </v-card-title>
                  <v-card-text>
                    You will have to logout and login again on the
                    <b>httpSMS</b> Android app with your new API key after you
                    rotate it.
                  </v-card-text>
                  <v-card-actions class="pb-4">
                    <v-btn
                      color="primary"
                      :loading="rotatingApiKey"
                      @click="rotateApiKeyHandler"
                    >
                      <v-icon start :icon="mdiRefresh"></v-icon>
                      Yes Rotate Key
                    </v-btn>
                    <v-spacer></v-spacer>
                    <v-btn variant="text" @click="showRotateApiKey = false">
                      Close
                    </v-btn>
                  </v-card-actions>
                </v-card>
              </v-dialog>
            </div>
            <h5 class="text-h4 mb-3 mt-12">V2 API Delivery Webhook</h5>
            <p class="text-medium-emphasis">
              URL called after each SMS delivery or failure when using the
              <code>POST /v2/send</code> endpoint. Leave empty to disable.
            </p>
            <v-text-field
              v-model="v2WebhookUrl"
              variant="outlined"
              density="compact"
              clearable
              persistent-placeholder
              persistent-hint
              label="Delivery Webhook URL"
              placeholder="https://example.com/sms-delivery"
              hint="A POST request with the delivery status will be sent to this URL."
            ></v-text-field>
            <div class="mb-6">
              <loading-button
                color="primary"
                :loading="testingV2WebhookUrl || savingV2WebhookUrl"
                @click="testAndSaveV2WebhookUrl"
              >
                Save Webhook URL
              </loading-button>
            </div>
            <p class="text-medium-emphasis">
              Signing secret sent in the <code>x-webhook-signature</code>
              header of every delivery webhook, as
              <code>sha256=&lt;HMAC-SHA256(body, secret)&gt;</code>. Use it to
              verify that a webhook request actually came from httpSMS.
            </p>
            <div v-if="webhookSecret" class="d-flex flex-wrap align-center">
              <v-text-field
                :append-icon="webhookSecretShow ? mdiEye : mdiEyeOff"
                :type="webhookSecretShow ? 'text' : 'password'"
                :model-value="webhookSecret"
                readonly
                variant="outlined"
                density="compact"
                class="flex-grow-1"
                style="min-width: 250px"
                @click:append="webhookSecretShow = !webhookSecretShow"
              ></v-text-field>
              <copy-button
                :value="webhookSecret"
                color="primary"
                class="ml-4 mb-6"
                copy-text="Copy Secret"
                notification-text="Webhook secret copied successfully"
              ></copy-button>
            </div>
            <div class="mb-6">
              <loading-button
                :color="webhookSecret ? 'warning' : 'primary'"
                :loading="rotatingWebhookSecret"
                @click="rotateWebhookSecretHandler"
              >
                <v-icon start :icon="mdiRefresh"></v-icon>
                {{ webhookSecret ? 'Rotate Secret' : 'Generate Secret' }}
              </loading-button>
            </div>
            <h5 id="webhook-settings" class="text-h4 mb-3 mt-12">Webhooks</h5>
            <p class="text-medium-emphasis">
              Webhooks allow us to send events to your server for example when
              the android phone receives an SMS message we can forward the
              message to your server.
            </p>
            <div v-if="loadingWebhooks">
              <v-progress-circular
                :size="60"
                :width="2"
                color="primary"
                class="mb-4"
                indeterminate
              ></v-progress-circular>
            </div>
            <v-table v-else-if="webhooks.length" class="mb-4">
              <thead>
                <tr class="text-uppercase subtitle-2">
                  <th v-if="display.xlAndUp.value" class="text-left">
                    ID
                  </th>
                  <th class="text-left text-break">Callback URL</th>
                  <th v-if="display.lgAndUp.value" class="text-center">
                    Events
                  </th>
                  <th class="text-center">Action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="webhook in webhooks" :key="webhook.id">
                  <td v-if="display.xlAndUp.value" class="text-left">
                    {{ webhook.id }}
                  </td>
                  <td class="text-break">{{ webhook.url }}</td>
                  <td v-if="display.lgAndUp.value" class="text-center">
                    <v-chip
                      v-for="event in webhook.events"
                      :key="event"
                      size="small"
                      >{{ event }}</v-chip
                    >
                  </td>
                  <td class="text-center">
                    <v-btn
                      :icon="display.mdAndDown.value"
                      size="small"
                      color="info"
                      :disabled="updatingWebhook"
                      @click.prevent="onWebhookEdit(webhook.id)"
                    >
                      <v-icon size="small" :icon="mdiSquareEditOutline"></v-icon>
                      <span v-if="!display.mdAndDown.value">
                        Edit
                      </span>
                    </v-btn>
                  </td>
                </tr>
              </tbody>
            </v-table>
            <div class="d-flex">
              <v-btn color="primary" @click="onWebhookCreate">
                <v-icon start :icon="mdiLinkVariant"></v-icon>
                Add webhook
              </v-btn>
              <v-btn
                v-if="display.lgAndUp.value"
                class="ml-4"
                href="https://docs.httpsms.com/webhooks/introduction"
                >Documentation</v-btn
              >
            </div>
            <h5 id="discord-settings" class="text-h4 mb-3 mt-12">
              Discord Integration
            </h5>
            <p class="text-medium-emphasis">
              Send and receive SMS messages without leaving your discord server
              with the httpSMS discord app using the
              <code>/httpsms</code> command.
            </p>
            <div v-if="loadingDiscordIntegrations">
              <v-progress-circular
                :size="60"
                :width="2"
                color="primary"
                class="mb-4"
                indeterminate
              ></v-progress-circular>
            </div>
            <v-table v-else-if="discords.length" class="mb-4">
              <thead>
                <tr class="text-uppercase subtitle-2">
                  <th class="text-left">Name</th>
                  <th class="text-left">Server ID</th>
                  <th class="text-left">Channel ID</th>
                  <th class="text-center">Action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="discord in discords" :key="discord.id">
                  <td class="text-left">
                    {{ discord.name }}
                  </td>
                  <td class="text-left">
                    {{ discord.server_id }}
                  </td>
                  <td class="text-left">
                    {{ discord.incoming_channel_id }}
                  </td>
                  <td class="text-center">
                    <v-btn
                      :icon="display.mdAndDown.value"
                      size="small"
                      color="info"
                      :disabled="updatingDiscord"
                      @click.prevent="onDiscordEdit(discord.id)"
                    >
                      <v-icon size="small" :icon="mdiSquareEditOutline"></v-icon>
                      <span v-if="!display.mdAndDown.value">
                        Edit
                      </span>
                    </v-btn>
                  </td>
                </tr>
              </tbody>
            </v-table>
            <v-btn color="#5865f2" @click="onDiscordCreate">
              <v-img
                contain
                height="24"
                width="24"
                class="mr-2"
                src="~/assets/img/discord-logo.svg"
              ></v-img>
              Add Discord
            </v-btn>
            <h5 id="phones" class="text-h4 mb-3 mt-12">Phones</h5>
            <p class="text-medium-emphasis">
              List of mobile phones which are registered for sending and
              receiving SMS messages.
            </p>
            <v-table>
              <thead>
                <tr class="text-uppercase subtitle-2">
                  <th v-if="display.xlAndUp.value" class="text-left">
                    ID
                  </th>
                  <th class="text-left">Phone Number</th>
                  <th v-if="display.lgAndUp.value" class="text-center">
                    Retries
                  </th>
                  <th class="text-center">Rate</th>
                  <th class="text-center">Updated At</th>
                  <th class="text-center">Action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="phone in store.getPhones" :key="phone.id">
                  <td v-if="display.xlAndUp.value" class="text-left">
                    {{ phone.id }}
                  </td>
                  <td>{{ formatPhoneNumber(phone.phone_number) }}</td>
                  <td v-if="display.lgAndUp.value">
                    <div class="d-flex justify-center">
                      {{
                        phone.max_send_attempts ? phone.max_send_attempts : 1
                      }}
                    </div>
                  </td>
                  <td class="text-center">
                    <span v-if="phone.messages_per_minute"
                      >{{ phone.messages_per_minute }}/min</span
                    >
                    <span v-else>Unlimited</span>
                  </td>
                  <td class="text-center">
                    {{ formatTimestamp(phone.updated_at) }}
                  </td>
                  <td class="text-center">
                    <v-btn
                      :icon="display.mdAndDown.value"
                      color="info"
                      :disabled="updatingPhone"
                      @click.prevent="showEditPhone(phone.id)"
                    >
                      <v-icon size="small" :icon="mdiSquareEditOutline"></v-icon>
                      <span v-if="!display.mdAndDown.value">
                        Edit
                      </span>
                    </v-btn>
                  </td>
                </tr>
              </tbody>
            </v-table>
            <h5 id="email-notifications" class="text-h4 mb-3 mt-12">
              Email Notifications
            </h5>
            <p class="text-medium-emphasis">
              Manage the email notifications which you receive from httpSMS.
              Feel free to turn on/off individual notifications anytime so you
              don't get overloaded with emails
            </p>
            <v-switch
              v-model="notificationSettings.heartbeat_enabled"
              label="Heartbeat emails"
              :disabled="updatingEmailNotifications"
              hint="This switch controls email notifications we send when we don't receive a heartbeat from your phone for 1 hour."
              persistent-hint
            ></v-switch>
            <v-switch
              v-model="notificationSettings.webhook_enabled"
              label="Webhook and discord emails"
              :disabled="updatingEmailNotifications"
              hint="This switch controls email notifications we send when we can't forward events to your discord server or to your webhook."
              persistent-hint
            ></v-switch>
            <v-switch
              v-model="notificationSettings.message_status_enabled"
              label="Message status emails"
              :disabled="updatingEmailNotifications"
              hint="This switch controls email notifications we send when we your message is failed or expired."
              persistent-hint
            ></v-switch>
            <v-switch
              v-model="notificationSettings.newsletter_enabled"
              label="Newsletter emails"
              :disabled="updatingEmailNotifications"
              hint="This switch controls newsletter emails about new features, updates, and promotions."
              persistent-hint
            ></v-switch>
            <v-btn
              color="primary"
              :loading="updatingEmailNotifications"
              class="mt-4"
              @click="saveEmailNotificationsHandler"
            >
              <v-icon start :icon="mdiContentSave"></v-icon>
              Save Notification Settings
            </v-btn>
            <h5 id="email-notifications" class="text-h4 text-error mb-3 mt-12">
              Delete Account
            </h5>
            <p v-if="hasActiveSubscription" class="text-medium-emphasis">
              You cannot delete your account because you have an active
              subscription on httpSMS.
              <router-link class="text-decoration-none" to="/billing"
                >Cancel your subscription</router-link
              >
              before deleting your account.
            </p>
            <p v-else class="text-medium-emphasis">
              You can delete all your data on httpSMS by clicking the button
              below. This action is <b>irreversible</b> and all your data will
              be permanently deleted from the httpSMS database instantly and it
              cannot be recovered.
            </p>
            <v-btn
              color="error"
              :loading="deletingAccount"
              class="mt-4"
              :disabled="hasActiveSubscription"
              @click="showDeleteAccountDialog = true"
            >
              <v-icon start :icon="mdiDelete"></v-icon>
              Delete your Account
            </v-btn>
            <v-dialog
              v-model="showDeleteAccountDialog"
              overlay-opacity="0.9"
              max-width="600px"
            >
              <v-card>
                <v-card-title class="justify-center text-center"
                  >Delete your httpSMS account</v-card-title
                >
                <v-card-text class="mt-2 text-center">
                  Are you sure you want to delete your account? This action is
                  <b>irreversible</b> and all your data will be permanently
                  deleted from the httpSMS database instantly.
                </v-card-text>
                <v-card-actions>
                  <v-btn
                    color="error"
                    variant="text"
                    :loading="deletingAccount"
                    @click="deleteUserAccountHandler"
                  >
                    <v-icon v-if="display.lgAndUp.value" start :icon="mdiDelete"></v-icon>
                    Delete My Account
                  </v-btn>
                  <v-spacer></v-spacer>
                  <v-btn
                    color="primary"
                    @click="showDeleteAccountDialog = false"
                  >
                    <span v-if="display.lgAndUp.value"
                      >Keep My account</span
                    >
                    <span v-else>Close</span>
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-dialog>
          </v-col>
        </v-row>
      </v-container>
    </div>
    <v-dialog v-model="showPhoneEdit" overlay-opacity="0.9" max-width="700px">
      <v-card>
        <v-card-title>Edit Phone</v-card-title>
        <v-card-text v-if="activePhone" class="mt-6">
          <v-container>
            <v-row>
              <v-col>
                <v-text-field
                  variant="outlined"
                  density="compact"
                  disabled
                  label="ID"
                  :model-value="activePhone.id"
                >
                </v-text-field>
                <v-text-field
                  variant="outlined"
                  disabled
                  density="compact"
                  label="Phone Number"
                  :model-value="activePhone.phone_number"
                >
                </v-text-field>
                <v-text-field
                  variant="outlined"
                  disabled
                  density="compact"
                  label="SIM"
                  :model-value="activePhone.sim"
                >
                </v-text-field>
                <v-textarea
                  variant="outlined"
                  disabled
                  density="compact"
                  label="FCM Token"
                  :model-value="activePhone.fcm_token"
                >
                </v-textarea>
                <v-text-field
                  v-model="activePhone.message_expiration_seconds"
                  variant="outlined"
                  type="number"
                  density="compact"
                  label="Message Expiration (seconds)"
                >
                </v-text-field>
                <v-text-field
                  v-model="activePhone.messages_per_minute"
                  variant="outlined"
                  type="number"
                  density="compact"
                  label="Messages Per Minute"
                >
                </v-text-field>
                <v-text-field
                  v-model="activePhone.max_send_attempts"
                  variant="outlined"
                  type="number"
                  density="compact"
                  placeholder="How many retries when sending an SMS"
                  label="Max Send Attempts"
                >
                </v-text-field>
                <v-textarea
                  v-model="activePhone.missed_call_auto_reply"
                  variant="outlined"
                  density="compact"
                  label="Missed Call AutoReply"
                  persistent-placeholder
                  persistent-hint
                  placeholder="We are currently closed at the moment, please send us a text message from  09:00 to 17:00"
                  hint="Here you can configure an automated SMS message which is sent to the caller when this phone has a missed call"
                >
                </v-textarea>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions class="mt-n8">
          <v-btn size="small" color="info" @click="updatePhoneHandler">
            <v-icon v-if="display.lgAndUp.value" size="small" :icon="mdiContentSave"></v-icon>
            Update
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn size="small" color="error" variant="text" @click="deletePhoneHandler(activePhone!.id)">
            <v-icon v-if="display.lgAndUp.value" size="small" :icon="mdiDelete"></v-icon>
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="showWebhookEdit" overlay-opacity="0.9" max-width="600px">
      <v-card>
        <v-card-title>
          <span v-if="!activeWebhook.id">Add a new&nbsp;</span>
          <span v-else>Edit&nbsp;</span>
          webhook
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col class="pt-8">
              <v-text-field
                v-if="activeWebhook.id"
                variant="outlined"
                density="compact"
                disabled
                label="ID"
                :model-value="activeWebhook.id"
              >
              </v-text-field>
              <v-text-field
                v-model="activeWebhook.url"
                variant="outlined"
                density="compact"
                label="Callback URL"
                persistent-placeholder
                persistent-hint
                :error="errorMessages.has('url')"
                :error-messages="errorMessages.get('url')"
                hint="A POST request will be sent to this URL every time an event is triggered in httpSMS."
                placeholder="https://example.com/webhook"
              >
              </v-text-field>
              <v-text-field
                v-model="activeWebhook.signing_key"
                variant="outlined"
                density="compact"
                class="mt-6"
                persistent-placeholder
                persistent-hint
                label="Signing Key (optional)"
                placeholder="******************"
                :error="errorMessages.has('signing_key')"
                :error-messages="errorMessages.get('signing_key')"
                hint="The signing key is used to verify the webhook is sent from httpSMS."
              >
              </v-text-field>
              <v-select
                v-model="activeWebhook.events"
                :items="events"
                label="Events"
                multiple
                variant="outlined"
                persistent-placeholder
                class="mt-6"
                density="compact"
                :error="errorMessages.has('events')"
                :error-messages="errorMessages.get('events')"
                hint="Select multiple httpSMS events to watch for"
                persistent-hint
              ></v-select>
              <v-select
                v-model="activeWebhook.phone_numbers"
                :items="phoneNumbers"
                label="Phone Numbers"
                multiple
                variant="outlined"
                persistent-placeholder
                class="mt-6"
                density="compact"
                :error="errorMessages.has('phone_numbers')"
                :error-messages="errorMessages.get('phone_numbers')"
                hint="Select multiple phone numbers to watch for events"
                persistent-hint
              ></v-select>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="mt-n4 pb-4">
          <loading-button
            v-if="!activeWebhook.id"
            :icon="mdiContentSave"
            :loading="updatingWebhook"
            @click="createWebhookHandler"
          >
            Save Webhook
          </loading-button>
          <loading-button
            v-else
            size="small"
            color="info"
            :loading="updatingWebhook"
            @click="updateWebhookHandler"
          >
            <v-icon v-if="display.lgAndUp.value" size="small" :icon="mdiContentSave"></v-icon>
            Update Webhook
          </loading-button>
          <v-spacer></v-spacer>
          <v-btn
            v-if="activeWebhook.id"
            :disabled="updatingWebhook"
            size="small"
            color="error"
            variant="text"
            @click="deleteWebhookHandler(activeWebhook.id!)"
          >
            <v-icon v-if="display.lgAndUp.value" size="small" :icon="mdiDelete"></v-icon>
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="showDiscordEdit" overlay-opacity="0.9" max-width="700px">
      <v-card>
        <v-card-title>
          <span v-if="!activeDiscord.id">Add a new&nbsp;</span>
          <span v-else>Edit&nbsp;</span>
          discord integration
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col class="pt-8">
              <p class="mt-n4 subtitle-1">
                Click the button below to add the httpSMS bot to your discord
                server. You need to do this so we can have permission to send
                and receive messages on your discord server.
              </p>
              <v-btn
                color="#5865f2"
                class="mb-6"
                target="_blank"
                href="https://discord.com/api/oauth2/authorize?client_id=1095780203256627291&permissions=2147485760&scope=bot%20applications.commands"
              >
                <v-icon start :icon="mdiConnection"></v-icon>
                Add Discord Bot
              </v-btn>
              <v-text-field
                v-if="activeDiscord.id"
                variant="outlined"
                density="compact"
                disabled
                label="ID"
                :model-value="activeDiscord.id"
              >
              </v-text-field>
              <v-text-field
                v-model="activeDiscord.name"
                variant="outlined"
                density="compact"
                label="Name"
                persistent-placeholder
                persistent-hint
                :error="errorMessages.has('name')"
                :error-messages="errorMessages.get('name')"
                hint="The name of the discord integration"
                placeholder="e.g Game Server"
              >
              </v-text-field>
              <v-text-field
                v-model="activeDiscord.server_id"
                variant="outlined"
                density="compact"
                class="mt-6"
                persistent-placeholder
                persistent-hint
                label="Discord Server ID"
                placeholder="e.g 1095778291488653372"
                :error="errorMessages.has('server_id')"
                :error-messages="errorMessages.get('server_id')"
                hint="You can get this by right clicking on your server and clicking Copy Server ID."
              >
              </v-text-field>
              <v-text-field
                v-model="activeDiscord.incoming_channel_id"
                variant="outlined"
                density="compact"
                class="mt-6"
                persistent-placeholder
                persistent-hint
                label="Discord Incoming Channel ID"
                placeholder="e.g 1095778291488653372"
                :error="errorMessages.has('incoming_channel_id')"
                :error-messages="errorMessages.get('incoming_channel_id')"
                hint="You can get this by right clicking on your discord channel and clicking Copy Chanel ID."
              >
              </v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="mt-n4 pb-4 pl-6">
          <loading-button
            v-if="!activeDiscord.id"
            :icon="mdiContentSave"
            :loading="updatingDiscord"
            @click="createDiscordHandler"
          >
            Save Discord Integration
          </loading-button>
          <loading-button
            v-else
            color="info"
            :loading="updatingDiscord"
            @click="updateDiscordHandler"
          >
            <v-icon v-if="display.lgAndUp.value" size="small" :icon="mdiContentSave"></v-icon>
            Update Discord Integration
          </loading-button>
          <v-spacer></v-spacer>
          <v-btn
            v-if="activeDiscord.id"
            :disabled="updatingDiscord"
            color="error"
            variant="text"
            @click="deleteDiscordHandler(activeDiscord.id!)"
          >
            <v-icon v-if="display.lgAndUp.value" size="small" :icon="mdiDelete"></v-icon>
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useDisplay } from 'vuetify'
import {
  mdiArrowLeft,
  mdiAccountCircle,
  mdiShieldCheck,
  mdiDelete,
  mdiContentSave,
  mdiConnection,
  mdiEye,
  mdiRefresh,
  mdiLinkVariant,
  mdiEyeOff,
  mdiSquareEditOutline,
  mdiQrcode,
} from '@mdi/js'
import QRCode from 'qrcode'
import { useAppStore } from '~/stores/app'
import { ErrorMessages } from '~/utils/errors'
import { formatPhoneNumber, formatTimestamp } from '~/plugins/filters'
import type { EntitiesPhone, EntitiesDiscord, EntitiesWebhook } from '~/models/api'

definePageMeta({ middleware: ['auth'] })

useHead({ title: 'Settings - httpSMS' })

const store = useAppStore()
const router = useRouter()
const route = useRoute()
const display = useDisplay()

// Refs
const qrCodeCanvasRef = ref<HTMLCanvasElement>()

// Data
const errorMessages = ref(new ErrorMessages())
const apiKeyShow = ref(false)
const showPhoneEdit = ref(false)
const showDiscordEdit = ref(false)
const showRotateApiKey = ref(false)
const rotatingApiKey = ref(false)
const showQrCodeDialog = ref(false)
const deletingAccount = ref(false)
const showDeleteAccountDialog = ref(false)
const activeWebhook = ref<{
  id: string | null
  url: string
  signing_key: string
  phone_numbers: string[]
  events: string[]
}>({
  id: null,
  url: '',
  signing_key: '',
  phone_numbers: [],
  events: ['message.phone.received'],
})
const activeDiscord = ref<{
  id: string | null
  name: string
  server_id: string
  missed_call_auto_reply?: string
  incoming_channel_id: string
}>({
  id: null,
  name: '',
  server_id: '',
  missed_call_auto_reply: '',
  incoming_channel_id: '',
})
const v2WebhookUrl = ref('')
const savingV2WebhookUrl = ref(false)
const testingV2WebhookUrl = ref(false)
const webhookSecret = ref('')
const webhookSecretShow = ref(false)
const rotatingWebhookSecret = ref(false)
const updatingEmailNotifications = ref(false)
const notificationSettings = ref({
  webhook_enabled: true,
  message_status_enabled: true,
  newsletter_enabled: true,
  heartbeat_enabled: true,
})
const updatingWebhook = ref(false)
const loadingWebhooks = ref(false)
const discords = ref<EntitiesDiscord[]>([])
const webhooks = ref<EntitiesWebhook[]>([])
const showWebhookEdit = ref(false)
const activePhone = ref<EntitiesPhone | null>(null)
const updatingPhone = ref(false)
const updatingDiscord = ref(false)
const loadingDiscordIntegrations = ref(false)
const events = [
  'message.phone.received',
  'message.phone.sent',
  'message.phone.delivered',
  'message.send.failed',
  'message.send.expired',
  'message.call.missed',
  'phone.heartbeat.offline',
  'phone.heartbeat.online',
]

// Computed
const apiKey = computed(() => {
  if (store.getUser === null) {
    return ''
  }
  return store.getUser.api_key
})

const hasActiveSubscription = computed(() => {
  if (store.getUser === null) {
    return true
  }
  return store.getUser.subscription_renews_at != null
})

const timezones = computed(() => {
  return Intl.supportedValuesOf('timeZone')
})

const phoneNumbers = computed(() => {
  return store.getPhones.map((phone) => phone.phone_number)
})

// Watch
watch(showQrCodeDialog, (newVal) => {
  if (newVal && apiKey.value) {
    nextTick(() => {
      generateQrCode(apiKey.value)
    })
  }
})

// Lifecycle
onMounted(async () => {
  await Promise.all([
    store.clearAxiosError(),
    store.loadUser(),
    store.loadPhones(),
  ])
  loadWebhooks()
  loadDiscordIntegrations()
  updateEmailNotificationsFromStore()
  v2WebhookUrl.value = store.getUser?.webhook_url ?? ''
  webhookSecret.value = store.getUser?.webhook_secret ?? ''
  if (route.hash) {
    const el = document.querySelector(route.hash)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth' })
    }
  }
})

// Methods
function generateQrCode(text: string) {
  const canvas = qrCodeCanvasRef.value
  if (canvas) {
    QRCode.toCanvas(canvas, text, { errorCorrectionLevel: 'H' }, (err: any) => {
      if (err) {
        store.addNotification({
          message: 'Failed to generate API key QR code',
          type: 'error',
        })
      }
    })
  }
}

function updateEmailNotificationsFromStore() {
  notificationSettings.value = {
    webhook_enabled: store.getUser!.notification_webhook_enabled,
    message_status_enabled: store.getUser!.notification_message_status_enabled,
    heartbeat_enabled: store.getUser!.notification_heartbeat_enabled,
    newsletter_enabled: store.getUser!.notification_newsletter_enabled,
  }
}

function showEditPhone(phoneId: string) {
  const phone = store.getPhones.find((x) => x.id === phoneId)
  if (!phone) {
    return
  }
  activePhone.value = { ...phone }
  showPhoneEdit.value = true
  resetErrors()
}

function onWebhookEdit(webhookId: string) {
  const webhook = webhooks.value.find((x) => x.id === webhookId)
  if (!webhook) {
    return
  }
  activeWebhook.value = {
    id: webhook.id,
    url: webhook.url,
    phone_numbers: webhook.phone_numbers.filter(
      (x) => phoneNumbers.value.find((y) => y === x) !== undefined,
    ),
    signing_key: webhook.signing_key,
    events: webhook.events,
  }
  showWebhookEdit.value = true
  resetErrors()
}

function onDiscordEdit(discordId: string) {
  const discord = discords.value.find((x) => x.id === discordId)
  if (!discord) {
    return
  }
  activeDiscord.value = {
    id: discord.id,
    name: discord.name,
    server_id: discord.server_id,
    incoming_channel_id: discord.incoming_channel_id,
  }
  showDiscordEdit.value = true
  resetErrors()
}

function onWebhookCreate() {
  activeWebhook.value = {
    id: null,
    url: '',
    signing_key: '',
    phone_numbers: store.getPhones.map((phone) => phone.phone_number),
    events: [
      'message.phone.received',
      'message.phone.sent',
      'message.phone.delivered',
      'message.send.failed',
      'message.send.expired',
    ],
  }
  showWebhookEdit.value = true
  resetErrors()
}

function onDiscordCreate() {
  activeDiscord.value = {
    id: null,
    name: '',
    server_id: '',
    incoming_channel_id: '',
    missed_call_auto_reply: '',
  }
  showDiscordEdit.value = true
  resetErrors()
}

async function updatePhoneHandler() {
  updatingPhone.value = true
  await store.clearAxiosError()
  store.updatePhone(activePhone.value!).finally(() => {
    if (!store.getAxiosError) {
      updatingPhone.value = false
      showPhoneEdit.value = false
      activePhone.value = null
    }
  })
}

function resetErrors() {
  errorMessages.value = new ErrorMessages()
}

function createDiscordHandler() {
  resetErrors()
  updatingDiscord.value = true
  store
    .createDiscord(activeDiscord.value as any)
    .then(() => {
      store.addNotification({
        message: 'Discord integration created successfully',
        type: 'success',
      })
      showDiscordEdit.value = false
      loadDiscordIntegrations()
    })
    .catch((errors) => {
      errorMessages.value = errors
    })
    .finally(() => {
      updatingDiscord.value = false
    })
}

function saveEmailNotificationsHandler() {
  updatingEmailNotifications.value = true
  store
    .saveEmailNotifications(notificationSettings.value)
    .then(() => {
      store.addNotification({
        message: 'Email notifications saved successfully',
        type: 'success',
      })
      updateEmailNotificationsFromStore()
    })
    .finally(() => {
      updatingEmailNotifications.value = false
    })
}

function updateDiscordHandler() {
  resetErrors()
  updatingDiscord.value = true
  store
    .updateDiscordIntegration(activeDiscord.value as any)
    .then(() => {
      store.addNotification({
        message: 'Discord integration updated successfully',
        type: 'success',
      })
      showDiscordEdit.value = false
      loadDiscordIntegrations()
    })
    .catch((errors) => {
      errorMessages.value = errors
    })
    .finally(() => {
      updatingDiscord.value = false
    })
}

function deleteDiscordHandler(discordId: string) {
  updatingDiscord.value = true
  store
    .deleteDiscordIntegration(discordId)
    .then(() => {
      store.addNotification({
        message: 'Discord integration deleted successfully',
        type: 'success',
      })
      showDiscordEdit.value = false
      loadDiscordIntegrations()
    })
    .finally(() => {
      updatingDiscord.value = false
    })
}

function createWebhookHandler() {
  resetErrors()
  updatingWebhook.value = true
  store
    .createWebhook(activeWebhook.value as any)
    .then(() => {
      store.addNotification({
        message: 'Webhook created successfully',
        type: 'success',
      })
      showWebhookEdit.value = false
      loadWebhooks()
    })
    .catch((errors) => {
      errorMessages.value = errors
    })
    .finally(() => {
      updatingWebhook.value = false
    })
}

function saveV2WebhookUrl() {
  savingV2WebhookUrl.value = true
  store
    .updateV2WebhookUrl(v2WebhookUrl.value || null)
    .then(() => {
      store.addNotification({
        message: 'Delivery webhook URL saved successfully',
        type: 'success',
      })
    })
    .catch(() => {
      store.addNotification({
        message: 'Failed to save delivery webhook URL',
        type: 'error',
      })
    })
    .finally(() => {
      savingV2WebhookUrl.value = false
    })
}

function testAndSaveV2WebhookUrl() {
  if (!v2WebhookUrl.value) {
    saveV2WebhookUrl()
    return
  }

  testingV2WebhookUrl.value = true
  store
    .testV2WebhookUrl(v2WebhookUrl.value)
    .then(() => {
      saveV2WebhookUrl()
    })
    .catch((message: string) => {
      store.addNotification({
        message: message
          ? `${message} It was not saved.`
          : 'Webhook URL test failed. It was not saved.',
        type: 'error',
      })
    })
    .finally(() => {
      testingV2WebhookUrl.value = false
    })
}

function rotateWebhookSecretHandler() {
  rotatingWebhookSecret.value = true
  store
    .rotateWebhookSecret()
    .then((user) => {
      webhookSecret.value = user.webhook_secret ?? ''
      webhookSecretShow.value = true
      store.addNotification({
        message: 'Webhook secret rotated successfully',
        type: 'success',
      })
    })
    .catch(() => {
      store.addNotification({
        message: 'Failed to rotate webhook secret',
        type: 'error',
      })
    })
    .finally(() => {
      rotatingWebhookSecret.value = false
    })
}

function updateTimezoneHandler(timezone: string) {
  resetErrors()
  store
    .updateTimezone(timezone)
    .then(() => {
      store.addNotification({
        message: 'Timezone updated successfully',
        type: 'success',
      })
    })
    .catch(() => {
      store.addNotification({
        message: 'Failed to update timezone',
        type: 'error',
      })
    })
}

function updateWebhookHandler() {
  resetErrors()
  updatingWebhook.value = true
  store
    .updateWebhook(activeWebhook.value as any)
    .then(() => {
      store.addNotification({
        message: 'Webhook updated successfully',
        type: 'success',
      })
      showWebhookEdit.value = false
      loadWebhooks()
    })
    .catch((errors) => {
      errorMessages.value = errors
    })
    .finally(() => {
      updatingWebhook.value = false
    })
}

function rotateApiKeyHandler() {
  rotatingApiKey.value = true
  store
    .rotateApiKey(store.getUser!.id)
    .finally(() => {
      rotatingApiKey.value = false
      showRotateApiKey.value = false
    })
}

function deleteWebhookHandler(webhookId: string) {
  updatingWebhook.value = true
  store
    .deleteWebhook(webhookId)
    .then(() => {
      store.addNotification({
        message: 'Webhook deleted successfully',
        type: 'success',
      })
      showWebhookEdit.value = false
      loadWebhooks()
    })
    .finally(() => {
      updatingWebhook.value = false
    })
}

function loadWebhooks() {
  loadingWebhooks.value = true
  store
    .getWebhooks()
    .then((result) => {
      webhooks.value = result
    })
    .finally(() => {
      loadingWebhooks.value = false
    })
}

function loadDiscordIntegrations() {
  loadingDiscordIntegrations.value = true
  store
    .getDiscordIntegrations()
    .then((result) => {
      discords.value = result
    })
    .finally(() => {
      loadingDiscordIntegrations.value = false
    })
}

function deleteUserAccountHandler() {
  deletingAccount.value = true
  store
    .deleteUserAccount()
    .then((message) => {
      store.addNotification({
        message: message ?? 'Your account has been deleted successfully',
        type: 'success',
      })
      store.logout().then(() => {
        store.addNotification({
          type: 'info',
          message: 'You have successfully logged out',
        })
        router.push({ name: 'index' })
      })
    })
    .finally(() => {
      deletingAccount.value = false
      showDeleteAccountDialog.value = false
    })
}

function deletePhoneHandler(phoneId: string) {
  updatingPhone.value = true
  store.deletePhone(phoneId).finally(() => {
    updatingPhone.value = false
    showPhoneEdit.value = false
    activePhone.value = null
  })
}
</script>
