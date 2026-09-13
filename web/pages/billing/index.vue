<template>
  <v-container
    fluid
    class="px-0 pt-0 fill-height"
  >
    <div class="w-100 h-100">
      <v-app-bar height="60" :density="display.mdAndDown.value ? 'compact' : 'default'">
        <v-btn icon to="/threads">
          <v-icon :icon="mdiArrowLeft" />
        </v-btn>
        <v-toolbar-title>
          <div class="py-16">Account Usage</div>
        </v-toolbar-title>
        <v-progress-linear
          :active="loading"
          :indeterminate="loading"
          absolute
          bottom
        ></v-progress-linear>
      </v-app-bar>
      <v-container>
        <v-row>
          <v-col cols="12" md="9" offset-md="1" xl="8" offset-xl="2">
            <h5 class="text-h4 mb-3 mt-3">Current Plan</h5>
            <v-row v-if="store.getUser">
              <v-col md="6" xl="4">
                <v-alert density="compact" variant="tonal" prominent color="info">
                  <v-row align="center">
                    <v-col cols="12">
                      <h1
                        class="subtitle-1 font-weight-bold text-uppercase mt-3"
                      >
                        <span v-if="isOnFreePlan">{{ plan.name }}</span>
                        <span v-else-if="subscriptionIsCancelled"
                          ><span class="text-warning">{{ plan.name }}</span> →
                          Free</span
                        >
                        <span v-else>{{ plan.name }}</span>
                      </h1>
                      <p
                        v-if="
                          !isOnFreePlan &&
                          !isOnLifetimePlan &&
                          !subscriptionIsCancelled
                        "
                        class="text-medium-emphasis"
                      >
                        Your next bill is for <b>${{ plan.price }}</b> on
                        <b>{{
                          new Date(
                            store.getUser.subscription_renews_at,
                          ).toLocaleDateString()
                        }}</b>
                      </p>
                      <p v-if="isOnLifetimePlan" class="text-medium-emphasis">
                        You are on the life time plan which costs
                        <b>${{ plan.price }}</b>
                      </p>
                      <p
                        v-else-if="subscriptionIsCancelled"
                        class="text-medium-emphasis"
                      >
                        You will be downgraded to the <b>FREE</b> plan on
                        <b>{{
                          new Date(
                            store.getUser.subscription_ends_at,
                          ).toLocaleDateString()
                        }}</b>
                      </p>
                      <p v-else class="text-medium-emphasis">
                        {{ totalMessages }}/{{ plan.messagesPerMonth }} messages
                      </p>
                    </v-col>
                    <v-col cols="12" class="d-flex mb-2 mt-n6">
                      <loading-button
                        v-if="
                          !subscriptionIsCancelled &&
                          !isOnFreePlan &&
                          !isOnLifetimePlan
                        "
                        color="primary"
                        :loading="loading"
                        @click="updateDetails"
                      >
                        Update Plan
                      </loading-button>
                      <v-btn
                        v-else-if="!isOnLifetimePlan"
                        color="primary"
                        :href="checkoutURL"
                        >Upgrade Plan</v-btn
                      >
                      <v-spacer></v-spacer>
                      <v-dialog
                        v-if="
                          !subscriptionIsCancelled &&
                          !isOnFreePlan &&
                          !isOnLifetimePlan
                        "
                        v-model="dialog"
                        max-width="590"
                      >
                        <template #activator="{ props: activatorProps }">
                          <v-btn v-bind="activatorProps" color="error" variant="text">
                            Cancel Plan
                          </v-btn>
                        </template>
                        <v-card>
                          <v-card-text class="pt-4 mb-n6">
                            <h2 class="text-high-emphasis text-h5 mb-2">
                              Are you sure you want to cancel your subscription?
                            </h2>
                            <p>
                              You will be downgraded to the free plan at the end
                              of the current billing period on
                              <b>{{
                                new Date(
                                  store.getUser.subscription_renews_at,
                                ).toLocaleDateString()
                              }}</b>
                            </p>
                          </v-card-text>
                          <v-card-actions>
                            <v-btn color="primary" @click="dialog = false">
                              Keep Subscription
                            </v-btn>
                            <v-spacer></v-spacer>
                            <loading-button
                              v-if="!isOnFreePlan"
                              variant="text"
                              :loading="loading"
                              color="error"
                              @click="cancelPlan"
                            >
                              Cancel Plan
                            </loading-button>
                          </v-card-actions>
                        </v-card>
                      </v-dialog>
                    </v-col>
                  </v-row>
                </v-alert>
              </v-col>
            </v-row>
            <h2 v-if="isOnFreePlan" class="text-h4 mt-4 mb-2">Upgrade Plan</h2>
            <v-row v-if="isOnFreePlan">
              <v-col cols="12" md="6" xl="4">
                <v-hover v-slot="{ isHovering, props: hoverProps }">
                  <v-card
                    v-bind="hoverProps"
                    :color="isHovering ? 'black' : 'default'"
                    :href="checkoutURL"
                    variant="outlined"
                  >
                    <v-card-text>
                      <v-row align="center">
                        <v-col class="flex-grow-1">
                          <h1
                            class="subtitle-1 font-weight-bold text-uppercase mt-3"
                          >
                            Pro - Monthly
                          </h1>
                          <p class="text-medium-emphasis">5,000 messages monthly</p>
                        </v-col>
                        <v-col class="flex-shrink-1 flex-grow-0">
                          <span class="text-h5 text-high-emphasis">$10</span>/month
                        </v-col>
                      </v-row>
                    </v-card-text>
                  </v-card>
                </v-hover>
              </v-col>
              <v-col cols="12" md="6" xl="4">
                <v-hover v-slot="{ isHovering, props: hoverProps }">
                  <v-card
                    v-bind="hoverProps"
                    :color="isHovering ? 'black' : 'default'"
                    :href="checkoutURL"
                    variant="outlined"
                  >
                    <v-card-text>
                      <v-row align="center">
                        <v-col class="flex-grow-1">
                          <h1
                            class="subtitle-1 font-weight-bold text-uppercase mt-3"
                          >
                            Pro - Yearly
                            <v-chip size="small" color="primary" class="mt-n1"
                              >2 months free</v-chip
                            >
                          </h1>
                          <p class="text-medium-emphasis">5,000 messages monthly</p>
                        </v-col>
                        <v-col class="flex-shrink-1 flex-grow-0">
                          <span class="text-h5 text-high-emphasis">$100</span>/year
                        </v-col>
                      </v-row>
                    </v-card-text>
                  </v-card>
                </v-hover>
              </v-col>
              <v-col cols="12" md="6" xl="4">
                <v-hover v-slot="{ isHovering, props: hoverProps }">
                  <v-card
                    v-bind="hoverProps"
                    :color="isHovering ? 'black' : 'default'"
                    :href="enterpriseCheckoutURL"
                    variant="outlined"
                  >
                    <v-card-text>
                      <v-row align="center">
                        <v-col class="flex-grow-1">
                          <h1
                            class="subtitle-1 font-weight-bold text-uppercase mt-3"
                          >
                            100k - Monthly
                          </h1>
                          <p class="text-medium-emphasis">
                            100,000 messages monthly
                          </p>
                        </v-col>
                        <v-col class="flex-shrink-1 flex-grow-0">
                          <span class="text-h5 text-high-emphasis">$175</span>/month
                        </v-col>
                      </v-row>
                    </v-card-text>
                  </v-card>
                </v-hover>
              </v-col>
            </v-row>
            <h5 class="text-h4 mb-3 mt-8">Overview</h5>
            <p class="text-medium-emphasis">
              This is the summary of the sent messages and received messages in
              <code
                v-if="store.getBillingUsage"
                class="font-weight-bold"
                >{{
                  formatBillingPeriod(store.getBillingUsage.start_timestamp)
                }}</code
              >.
            </p>
            <v-row v-if="store.getBillingUsage">
              <v-col cols="12" md="4">
                <v-alert
                  dark
                  density="compact"
                  :icon="mdiCallMade"
                  prominent
                  type="info"
                  variant="tonal"
                >
                  <h2 class="text-h4 font-weight-bold mt-4">
                    {{ formatDecimal(store.getBillingUsage.sent_messages) }}
                  </h2>
                  <p class="text-medium-emphasis mt-n1">Messages Sent</p>
                </v-alert>
              </v-col>
              <v-col cols="12" md="4">
                <v-alert
                  dark
                  density="compact"
                  :icon="mdiCallReceived"
                  prominent
                  type="warning"
                  variant="tonal"
                >
                  <div class="d-flex">
                    <h2 class="text-h4 font-weight-bold mt-4">
                      {{
                        formatDecimal(store.getBillingUsage.received_messages)
                      }}
                    </h2>
                  </div>
                  <p class="text-medium-emphasis mt-n1">Messages Received</p>
                </v-alert>
              </v-col>
              <v-col cols="12" md="4">
                <v-alert
                  density="compact"
                  :icon="mdiCreditCard"
                  prominent
                  type="success"
                  variant="tonal"
                >
                  <h2 class="text-h4 font-weight-bold mt-4">
                    {{ formatMoney(store.getBillingUsage.total_cost) }}
                  </h2>
                  <p class="text-medium-emphasis mt-n1">Total Cost</p>
                </v-alert>
              </v-col>
            </v-row>
            <template v-if="store.getUser?.subscription_id != null">
              <h5 class="text-h4 mb-3 mt-8">Subscription Payments</h5>
              <p class="text-medium-emphasis">
                This is a list of your last 10 subscription payments made using
                our payment provider
                <a
                  class="text-decoration-none"
                  href="https://www.lemonsqueezy.com"
                  >Lemon Squeezy</a
                >.
              </p>
              <v-progress-circular
                v-if="payments == null && loadingSubscriptionPayments"
                :size="20"
                :width="2"
                color="primary"
                indeterminate
              ></v-progress-circular>
              <v-table v-if="payments">
                <thead>
                  <tr class="text-uppercase">
                    <th v-if="display.lgAndUp.value" class="text-left">
                      ID
                    </th>
                    <th class="text-left">Timestamp</th>
                    <th class="text-left">Status</th>
                    <th v-if="display.lgAndUp.value" class="text-left">
                      Tax
                    </th>
                    <th class="text-left">Total</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="payment in payments.data" :key="payment.id">
                    <td v-if="display.lgAndUp.value">
                      {{ payment.id }}
                    </td>
                    <td>
                      {{ formatTimestamp(payment.attributes.created_at) }}
                    </td>
                    <td>
                      <v-chip
                        v-if="payment.attributes.status === 'paid'"
                        color="success"
                      >
                        <v-avatar size="4" start class="bg-green-darken-4">
                          <v-icon size="small" :icon="mdiCheck" />
                        </v-avatar>
                        {{ payment.attributes.status_formatted }}
                      </v-chip>
                      <v-chip v-else color="error">
                        <v-avatar size="4" start class="bg-red-darken-4">
                          <v-icon size="small" :icon="mdiAlert" />
                        </v-avatar>
                        {{ payment.attributes.status_formatted }}
                      </v-chip>
                    </td>
                    <td v-if="display.lgAndUp.value">
                      {{ payment.attributes.tax_formatted }}
                    </td>
                    <td class="font-weight-bold">
                      {{ payment.attributes.total_formatted }}
                    </td>
                    <td class="text-right">
                      <v-btn
                        color="primary"
                        size="small"
                        @click="showInvoiceDialog(payment)"
                      >
                        <v-icon start :icon="mdiInvoice" />
                        Invoice
                      </v-btn>
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </template>
            <h5 class="text-h4 mb-3 mt-8">Usage History</h5>
            <p class="text-medium-emphasis">
              Summary of all the sent and received messages in the past 12
              months
            </p>
            <v-table>
              <thead>
                <tr class="text-uppercase">
                  <th class="text-left">Period</th>
                  <th class="text-left">
                    Sent
                    <span v-if="display.lgAndUp.value">Messages</span>
                  </th>
                  <th class="text-left">
                    Received
                    <span v-if="display.lgAndUp.value">Messages</span>
                  </th>
                  <th class="text-right">
                    <span v-if="display.lgAndUp.value">Total</span> Cost
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="billingUsage in store.getBillingUsageHistory"
                  :key="billingUsage.id"
                >
                  <td>
                    {{ formatBillingPeriod(billingUsage.start_timestamp) }}
                  </td>
                  <td>
                    {{ formatDecimal(billingUsage.sent_messages) }}
                  </td>
                  <td>
                    {{ billingUsage.received_messages }}
                  </td>
                  <td class="text-right font-weight-bold">
                    {{ formatMoney(billingUsage.total_cost) }}
                  </td>
                </tr>
              </tbody>
            </v-table>
          </v-col>
        </v-row>
      </v-container>
    </div>
    <v-dialog
      v-model="subscriptionInvoiceDialog"
      persistent
      overlay-opacity="0.9"
      max-width="600px"
    >
      <v-card>
        <v-card-title class="text-h4"> Generate Invoice </v-card-title>
        <v-card-subtitle class="mt-n1">
          Create an invoice for your
          <b>{{ selectedPayment?.attributes.total_formatted }}</b> payment on
          {{ selectedPayment ? formatTimestamp(selectedPayment.attributes.created_at) : '' }}
        </v-card-subtitle>
        <v-card-text>
          <v-container>
            <v-row>
              <v-col cols="12">
                <v-text-field
                  v-model="invoiceFormName"
                  density="compact"
                  :disabled="loading"
                  :error="errorMessages.has('name')"
                  :error-messages="errorMessages.get('name')"
                  label="Name"
                  placeholder="e.g Acme Corporation"
                  persistent-placeholder
                  variant="outlined"
                ></v-text-field>
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="invoiceFormAddress"
                  density="compact"
                  :disabled="loading"
                  :error="errorMessages.has('address')"
                  :error-messages="errorMessages.get('address')"
                  label="Address"
                  placeholder="e.g 221B Baker Street"
                  persistent-placeholder
                  variant="outlined"
                ></v-text-field>
              </v-col>
            </v-row>
            <v-row>
              <v-col cols="6">
                <v-text-field
                  v-model="invoiceFormCity"
                  density="compact"
                  :disabled="loading"
                  :error="errorMessages.has('city')"
                  :error-messages="errorMessages.get('city')"
                  label="City"
                  placeholder="e.g Los Angeles"
                  persistent-placeholder
                  variant="outlined"
                ></v-text-field>
              </v-col>
              <v-col cols="6">
                <v-text-field
                  v-if="invoiceStateOptions.length === 0"
                  v-model="invoiceFormState"
                  density="compact"
                  :disabled="loading"
                  :error="errorMessages.has('state')"
                  :error-messages="errorMessages.get('state')"
                  label="State"
                  placeholder="e.g CA"
                  persistent-placeholder
                  variant="outlined"
                ></v-text-field>
                <v-autocomplete
                  v-else
                  v-model="invoiceFormState"
                  density="compact"
                  :disabled="loading"
                  :error="errorMessages.has('state')"
                  :error-messages="errorMessages.get('state')"
                  :items="invoiceStateOptions"
                  label="State"
                  variant="outlined"
                  placeholder="e.g CA"
                  persistent-placeholder
                ></v-autocomplete>
              </v-col>
            </v-row>
            <v-row>
              <v-col cols="6">
                <v-text-field
                  v-model="invoiceFormZipCode"
                  density="compact"
                  :disabled="loading"
                  :error="errorMessages.has('zip_code')"
                  :error-messages="errorMessages.get('zip_code')"
                  label="Zip Code"
                  placeholder="e.g 46001"
                  persistent-placeholder
                  variant="outlined"
                ></v-text-field>
              </v-col>
              <v-col cols="6">
                <v-autocomplete
                  v-model="invoiceFormCountry"
                  density="compact"
                  :disabled="loading"
                  :error="errorMessages.has('country')"
                  :error-messages="errorMessages.get('country')"
                  :items="countries"
                  label="Country"
                  placeholder="e.g United States"
                  variant="outlined"
                  persistent-placeholder
                ></v-autocomplete>
              </v-col>
            </v-row>
            <v-row>
              <v-col cols="12">
                <v-textarea
                  v-model="invoiceFormNotes"
                  density="compact"
                  :disabled="loading"
                  :error="errorMessages.has('notes')"
                  :error-messages="errorMessages.get('notes')"
                  rows="3"
                  label="Notes (optional)"
                  placeholder="e.g Thanks for doing business with us!"
                  persistent-placeholder
                  variant="outlined"
                ></v-textarea>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions class="mt-n8 pb-4">
          <v-btn :loading="loading" color="primary" @click="generateInvoice">
            <v-icon start :icon="mdiDownloadOutline" />
            Download Invoice
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn color="error" variant="text" @click="subscriptionInvoiceDialog = false">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import {
  mdiArrowLeft,
  mdiDownloadOutline,
  mdiCheck,
  mdiAlert,
  mdiInvoice,
  mdiCallReceived,
  mdiCallMade,
  mdiCreditCard,
} from '@mdi/js'
import { useDisplay } from 'vuetify'
import { useAppStore } from '~/stores/app'
import type {
  RequestsUserPaymentInvoice,
  ResponsesUserSubscriptionPaymentsResponse,
} from '~/models/api'
import { ErrorMessages } from '~/utils/errors'
import {
  formatTimestamp,
  formatMoney,
  formatDecimal,
  formatBillingPeriod,
} from '~/plugins/filters'

definePageMeta({ middleware: ['auth'] })

useHead({
  title: 'Usage & Billing - httpSMS',
})

const store = useAppStore()
const router = useRouter()
const config = useRuntimeConfig()
const display = useDisplay()

// --- State ---
const loading = ref(true)
const loadingSubscriptionPayments = ref(false)
const dialog = ref(false)
const payments = ref<ResponsesUserSubscriptionPaymentsResponse | null>(null)
const selectedPayment = ref<SubscriptionPayment | null>(null)
const errorMessages = ref(new ErrorMessages())
const invoiceFormName = ref('')
const invoiceFormAddress = ref('')
const invoiceFormCity = ref('')
const invoiceFormState = ref('')
const invoiceFormZipCode = ref('')
const invoiceFormCountry = ref('')
const invoiceFormNotes = ref('')
const subscriptionInvoiceDialog = ref(false)

// --- Types ---
type PaymentPlan = {
  name: string
  id: string
  price: number
  messagesPerMonth: number
}

type SubscriptionPayment = {
  attributes: {
    created_at: string
    total_formatted: string
  }
  id: string
}

// --- Static Data ---
const countries = [
  { title: 'Afghanistan', value: 'AF' },
  { title: '\u00c5land Islands', value: 'AX' },
  { title: 'Albania', value: 'AL' },
  { title: 'Algeria', value: 'DZ' },
  { title: 'American Samoa', value: 'AS' },
  { title: 'Andorra', value: 'AD' },
  { title: 'Angola', value: 'AO' },
  { title: 'Anguilla', value: 'AI' },
  { title: 'Antarctica', value: 'AQ' },
  { title: 'Antigua and Barbuda', value: 'AG' },
  { title: 'Argentina', value: 'AR' },
  { title: 'Armenia', value: 'AM' },
  { title: 'Aruba', value: 'AW' },
  { title: 'Australia', value: 'AU' },
  { title: 'Austria', value: 'AT' },
  { title: 'Azerbaijan', value: 'AZ' },
  { title: 'Bahamas', value: 'BS' },
  { title: 'Bahrain', value: 'BH' },
  { title: 'Bangladesh', value: 'BD' },
  { title: 'Barbados', value: 'BB' },
  { title: 'Belarus', value: 'BY' },
  { title: 'Belgium', value: 'BE' },
  { title: 'Belize', value: 'BZ' },
  { title: 'Benin', value: 'BJ' },
  { title: 'Bermuda', value: 'BM' },
  { title: 'Bhutan', value: 'BT' },
  { title: 'Bolivia', value: 'BO' },
  { title: 'Bonaire', value: 'BQ' },
  { title: 'Bosnia and Herzegovina', value: 'BA' },
  { title: 'Botswana', value: 'BW' },
  { title: 'Bouvet Island', value: 'BV' },
  { title: 'Brazil', value: 'BR' },
  { title: 'British Indian Ocean', value: 'IO' },
  { title: 'Brunei Darussalam', value: 'BN' },
  { title: 'Bulgaria', value: 'BG' },
  { title: 'Burkina Faso', value: 'BF' },
  { title: 'Burundi', value: 'BI' },
  { title: 'Cabo Verde', value: 'CV' },
  { title: 'Cambodia', value: 'KH' },
  { title: 'Cameroon', value: 'CM' },
  { title: 'Canada', value: 'CA' },
  { title: 'Cayman Islands', value: 'KY' },
  { title: 'Central African Republic', value: 'CF' },
  { title: 'Chad', value: 'TD' },
  { title: 'Chile', value: 'CL' },
  { title: 'China', value: 'CN' },
  { title: 'Christmas Island', value: 'CX' },
  { title: 'Cocos (Keeling) Islands', value: 'CC' },
  { title: 'Colombia', value: 'CO' },
  { title: 'Comoros', value: 'KM' },
  { title: 'Congo', value: 'CG' },
  { title: 'Congo', value: 'CD' },
  { title: 'Cook Islands', value: 'CK' },
  { title: 'Costa Rica', value: 'CR' },
  { title: "C\u00f4te d'Ivoire", value: 'CI' },
  { title: 'Cuba', value: 'CU' },
  { title: 'Cura\u00e7ao', value: 'CW' },
  { title: 'Cyprus', value: 'CY' },
  { title: 'Czechia', value: 'CZ' },
  { title: 'Denmark', value: 'DK' },
  { title: 'Djibouti', value: 'DJ' },
  { title: 'Dominica', value: 'DM' },
  { title: 'Dominican Republic', value: 'DO' },
  { title: 'Ecuador', value: 'EC' },
  { title: 'Egypt', value: 'EG' },
  { title: 'El Salvador', value: 'SV' },
  { title: 'Equatorial Guinea', value: 'GQ' },
  { title: 'Eritrea', value: 'ER' },
  { title: 'Estonia', value: 'EE' },
  { title: 'Eswatini', value: 'SZ' },
  { title: 'Ethiopia', value: 'ET' },
  { title: 'Falkland Islands', value: 'FK' },
  { title: 'Faroe Islands', value: 'FO' },
  { title: 'Fiji', value: 'FJ' },
  { title: 'Finland', value: 'FI' },
  { title: 'France', value: 'FR' },
  { title: 'French Guiana', value: 'GF' },
  { title: 'French Polynesia', value: 'PF' },
  { title: 'French Southern Territories', value: 'TF' },
  { title: 'Gabon', value: 'GA' },
  { title: 'Gambia', value: 'GM' },
  { title: 'Georgia', value: 'GE' },
  { title: 'Germany', value: 'DE' },
  { title: 'Ghana', value: 'GH' },
  { title: 'Gibraltar', value: 'GI' },
  { title: 'Greece', value: 'GR' },
  { title: 'Greenland', value: 'GL' },
  { title: 'Grenada', value: 'GD' },
  { title: 'Guadeloupe', value: 'GP' },
  { title: 'Guam', value: 'GU' },
  { title: 'Guatemala', value: 'GT' },
  { title: 'Guernsey', value: 'GG' },
  { title: 'Guinea', value: 'GN' },
  { title: 'Guinea-Bissau', value: 'GW' },
  { title: 'Guyana', value: 'GY' },
  { title: 'Haiti', value: 'HT' },
  { title: 'Heard Island and McDonald Islands', value: 'HM' },
  { title: 'Holy See', value: 'VA' },
  { title: 'Honduras', value: 'HN' },
  { title: 'Hong Kong', value: 'HK' },
  { title: 'Hungary', value: 'HU' },
  { title: 'Iceland', value: 'IS' },
  { title: 'India', value: 'IN' },
  { title: 'Indonesia', value: 'ID' },
  { title: 'Iran', value: 'IR' },
  { title: 'Iraq', value: 'IQ' },
  { title: 'Ireland', value: 'IE' },
  { title: 'Isle of Man', value: 'IM' },
  { title: 'Israel', value: 'IL' },
  { title: 'Italy', value: 'IT' },
  { title: 'Jamaica', value: 'JM' },
  { title: 'Japan', value: 'JP' },
  { title: 'Jersey', value: 'JE' },
  { title: 'Jordan', value: 'JO' },
  { title: 'Kazakhstan', value: 'KZ' },
  { title: 'Kenya', value: 'KE' },
  { title: 'Kiribati', value: 'KI' },
  { title: 'North Korea', value: 'KP' },
  { title: 'South Korea', value: 'KR' },
  { title: 'Kuwait', value: 'KW' },
  { title: 'Kyrgyzstan', value: 'KG' },
  { title: "Lao People\u2019s Democratic Republic", value: 'LA' },
  { title: 'Latvia', value: 'LV' },
  { title: 'Lebanon', value: 'LB' },
  { title: 'Lesotho', value: 'LS' },
  { title: 'Liberia', value: 'LR' },
  { title: 'Libya', value: 'LY' },
  { title: 'Liechtenstein', value: 'LI' },
  { title: 'Lithuania', value: 'LT' },
  { title: 'Luxembourg', value: 'LU' },
  { title: 'Macao', value: 'MO' },
  { title: 'Madagascar', value: 'MG' },
  { title: 'Malawi', value: 'MW' },
  { title: 'Malaysia', value: 'MY' },
  { title: 'Maldives', value: 'MV' },
  { title: 'Mali', value: 'ML' },
  { title: 'Malta', value: 'MT' },
  { title: 'Marshall Islands', value: 'MH' },
  { title: 'Martinique', value: 'MQ' },
  { title: 'Mauritania', value: 'MR' },
  { title: 'Mauritius', value: 'MU' },
  { title: 'Mayotte', value: 'YT' },
  { title: 'Mexico', value: 'MX' },
  { title: 'Micronesia', value: 'FM' },
  { title: 'Moldova', value: 'MD' },
  { title: 'Monaco', value: 'MC' },
  { title: 'Mongolia', value: 'MN' },
  { title: 'Montenegro', value: 'ME' },
  { title: 'Montserrat', value: 'MS' },
  { title: 'Morocco', value: 'MA' },
  { title: 'Mozambique', value: 'MZ' },
  { title: 'Myanmar', value: 'MM' },
  { title: 'Namibia', value: 'NA' },
  { title: 'Nauru', value: 'NR' },
  { title: 'Nepal', value: 'NP' },
  { title: 'Netherlands', value: 'NL' },
  { title: 'New Caledonia', value: 'NC' },
  { title: 'New Zealand', value: 'NZ' },
  { title: 'Nicaragua', value: 'NI' },
  { title: 'Niger', value: 'NE' },
  { title: 'Nigeria', value: 'NG' },
  { title: 'Niue', value: 'NU' },
  { title: 'Norfolk Island', value: 'NF' },
  { title: 'North Macedonia', value: 'MK' },
  { title: 'Northern Mariana Islands', value: 'MP' },
  { title: 'Norway', value: 'NO' },
  { title: 'Oman', value: 'OM' },
  { title: 'Pakistan', value: 'PK' },
  { title: 'Palau', value: 'PW' },
  { title: 'Panama', value: 'PA' },
  { title: 'Papua New Guinea', value: 'PG' },
  { title: 'Paraguay', value: 'PY' },
  { title: 'Peru', value: 'PE' },
  { title: 'Philippines', value: 'PH' },
  { title: 'Pitcairn', value: 'PN' },
  { title: 'Poland', value: 'PL' },
  { title: 'Portugal', value: 'PT' },
  { title: 'Puerto Rico', value: 'PR' },
  { title: 'Qatar', value: 'QA' },
  { title: 'R\u00e9union', value: 'RE' },
  { title: 'Romania', value: 'RO' },
  { title: 'Russian Federation', value: 'RU' },
  { title: 'Rwanda', value: 'RW' },
  { title: 'Saint Barth\u00e9lemy', value: 'BL' },
  { title: 'Saint Helena, Ascension and Tristan da Cunha', value: 'SH' },
  { title: 'Saint Kitts and Nevis', value: 'KN' },
  { title: 'Saint Lucia', value: 'LC' },
  { title: 'Saint Martin (French part)', value: 'MF' },
  { title: 'Saint Pierre and Miquelon', value: 'PM' },
  { title: 'Saint Vincent and the Grenadines', value: 'VC' },
  { title: 'Samoa', value: 'WS' },
  { title: 'San Marino', value: 'SM' },
  { title: 'Sao Tome and Principe', value: 'ST' },
  { title: 'Saudi Arabia', value: 'SA' },
  { title: 'Senegal', value: 'SN' },
  { title: 'Serbia', value: 'RS' },
  { title: 'Seychelles', value: 'SC' },
  { title: 'Sierra Leone', value: 'SL' },
  { title: 'Singapore', value: 'SG' },
  { title: 'Slovakia', value: 'SK' },
  { title: 'Slovenia', value: 'SI' },
  { title: 'Solomon Islands', value: 'SB' },
  { title: 'Somalia', value: 'SO' },
  { title: 'South Africa', value: 'ZA' },
  { title: 'South Georgia and the South Sandwich Islands', value: 'GS' },
  { title: 'South Sudan', value: 'SS' },
  { title: 'Spain', value: 'ES' },
  { title: 'Sri Lanka', value: 'LK' },
  { title: 'Sudan', value: 'SD' },
  { title: 'Suriname', value: 'SR' },
  { title: 'Svalbard and Jan Mayen', value: 'SJ' },
  { title: 'Sweden', value: 'SE' },
  { title: 'Switzerland', value: 'CH' },
  { title: 'Syrian Arab Republic', value: 'SY' },
  { title: 'Taiwan, Province of China', value: 'TW' },
  { title: 'Tajikistan', value: 'TJ' },
  { title: 'Tanzania, United Republic of', value: 'TZ' },
  { title: 'Thailand', value: 'TH' },
  { title: 'Timor-Leste', value: 'TL' },
  { title: 'Togo', value: 'TG' },
  { title: 'Tokelau', value: 'TK' },
  { title: 'Tonga', value: 'TO' },
  { title: 'Trinidad and Tobago', value: 'TT' },
  { title: 'Tunisia', value: 'TN' },
  { title: 'Turkey', value: 'TR' },
  { title: 'Turkmenistan', value: 'TM' },
  { title: 'Turks and Caicos Islands', value: 'TC' },
  { title: 'Tuvalu', value: 'TV' },
  { title: 'Uganda', value: 'UG' },
  { title: 'Ukraine', value: 'UA' },
  { title: 'United Arab Emirates', value: 'AE' },
  { title: 'United Kingdom', value: 'GB' },
  { title: 'United States', value: 'US' },
  { title: 'United States Minor Outlying Islands', value: 'UM' },
  { title: 'Uruguay', value: 'UY' },
  { title: 'Uzbekistan', value: 'UZ' },
  { title: 'Vanuatu', value: 'VU' },
  { title: 'Venezuela', value: 'VE' },
  { title: 'Viet Nam', value: 'VN' },
  { title: 'Virgin Islands (British)', value: 'VG' },
  { title: 'Virgin Islands (U.S.)', value: 'VI' },
  { title: 'Wallis and Futuna', value: 'WF' },
  { title: 'Western Sahara', value: 'EH' },
  { title: 'Yemen', value: 'YE' },
  { title: 'Zambia', value: 'ZM' },
  { title: 'Zimbabwe', value: 'ZW' },
]

const plans: PaymentPlan[] = [
  { name: 'Free', id: 'free', messagesPerMonth: 200, price: 0 },
  { name: 'PRO - Monthly', id: 'pro-monthly', messagesPerMonth: 5000, price: 10 },
  { name: 'PRO - Yearly', id: 'pro-yearly', messagesPerMonth: 5000, price: 100 },
  { name: 'Ultra - Monthly', id: 'ultra-monthly', messagesPerMonth: 10000, price: 20 },
  { name: 'Ultra - Yearly', id: 'ultra-yearly', messagesPerMonth: 10000, price: 200 },
  { name: '20k - Monthly', id: '20k-monthly', messagesPerMonth: 20000, price: 35 },
  { name: '20k - Yearly', id: '20k-yearly', messagesPerMonth: 20000, price: 350 },
  { name: '50k - Monthly', id: '50k-monthly', messagesPerMonth: 50000, price: 89 },
  { name: '100k - Monthly', id: '100k-monthly', messagesPerMonth: 100000, price: 175 },
  { name: '200k - Monthly', id: '200k-monthly', messagesPerMonth: 200000, price: 350 },
  { name: 'PRO - Lifetime', id: 'pro-lifetime', messagesPerMonth: 10000, price: 1000 },
]

// --- Computed ---
const invoiceStateOptions = computed(() => {
  if (invoiceFormCountry.value === 'US') {
    return [
      { title: 'Alabama', value: 'AL' },
      { title: 'Alaska', value: 'AK' },
      { title: 'Arizona', value: 'AZ' },
      { title: 'Arkansas', value: 'AR' },
      { title: 'California', value: 'CA' },
      { title: 'Colorado', value: 'CO' },
      { title: 'Connecticut', value: 'CT' },
      { title: 'Delaware', value: 'DE' },
      { title: 'Florida', value: 'FL' },
      { title: 'Georgia', value: 'GA' },
      { title: 'Hawaii', value: 'HI' },
      { title: 'Idaho', value: 'ID' },
      { title: 'Illinois', value: 'IL' },
      { title: 'Indiana', value: 'IN' },
      { title: 'Iowa', value: 'IA' },
      { title: 'Kansas', value: 'KS' },
      { title: 'Kentucky', value: 'KY' },
      { title: 'Louisiana', value: 'LA' },
      { title: 'Maine', value: 'ME' },
      { title: 'Maryland', value: 'MD' },
      { title: 'Massachusetts', value: 'MA' },
      { title: 'Michigan', value: 'MI' },
      { title: 'Minnesota', value: 'MN' },
      { title: 'Mississippi', value: 'MS' },
      { title: 'Missouri', value: 'MO' },
      { title: 'Montana', value: 'MT' },
      { title: 'Nebraska', value: 'NE' },
      { title: 'Nevada', value: 'NV' },
      { title: 'New Hampshire', value: 'NH' },
      { title: 'New Jersey', value: 'NJ' },
      { title: 'New Mexico', value: 'NM' },
      { title: 'New York', value: 'NY' },
      { title: 'North Carolina', value: 'NC' },
      { title: 'North Dakota', value: 'ND' },
      { title: 'Ohio', value: 'OH' },
      { title: 'Oklahoma', value: 'OK' },
      { title: 'Oregon', value: 'OR' },
      { title: 'Pennsylvania', value: 'PA' },
      { title: 'Rhode Island', value: 'RI' },
      { title: 'South Carolina', value: 'SC' },
      { title: 'South Dakota', value: 'SD' },
      { title: 'Tennessee', value: 'TN' },
      { title: 'Texas', value: 'TX' },
      { title: 'Utah', value: 'UT' },
      { title: 'Vermont', value: 'VT' },
      { title: 'Virginia', value: 'VA' },
      { title: 'Washington', value: 'WA' },
      { title: 'West Virginia', value: 'WV' },
      { title: 'Wisconsin', value: 'WI' },
      { title: 'Wyoming', value: 'WY' },
      { title: 'District of Columbia', value: 'DC' },
    ]
  }
  if (invoiceFormCountry.value === 'CA') {
    return [
      { title: 'Alberta', value: 'AB' },
      { title: 'British Columbia', value: 'BC' },
      { title: 'Manitoba', value: 'MB' },
      { title: 'New Brunswick', value: 'NB' },
      { title: 'Newfoundland and Labrador', value: 'NL' },
      { title: 'Nova Scotia', value: 'NS' },
      { title: 'Ontario', value: 'ON' },
      { title: 'Prince Edward Island', value: 'PE' },
      { title: 'Quebec', value: 'QC' },
      { title: 'Saskatchewan', value: 'SK' },
      { title: 'Northwest Territories', value: 'NT' },
      { title: 'Nunavut', value: 'NU' },
      { title: 'Yukon', value: 'YT' },
    ]
  }
  return []
})

const checkoutURL = computed(() => {
  const url = new URL(config.public.checkoutURL as string)
  const user = store.getAuthUser
  url.searchParams.append('checkout[custom][user_id]', user?.id ?? '')
  url.searchParams.append('checkout[email]', user?.email ?? '')
  url.searchParams.append('checkout[name]', user?.displayName ?? '')
  return url.toString()
})

const enterpriseCheckoutURL = computed(() => {
  const url = new URL(config.public.enterpriseCheckoutURL as string)
  const user = store.getAuthUser
  url.searchParams.append('checkout[custom][user_id]', user?.id ?? '')
  url.searchParams.append('checkout[email]', user?.email ?? '')
  url.searchParams.append('checkout[name]', user?.displayName ?? '')
  return url.toString()
})

const plan = computed((): PaymentPlan => {
  return plans.find(
    (x) => x.id === (store.getUser?.subscription_name || 'free'),
  )!
})

const isOnFreePlan = computed((): boolean => {
  return plan.value.id === 'free'
})

const isOnLifetimePlan = computed((): boolean => {
  return plan.value.id === 'pro-lifetime'
})

const subscriptionIsCancelled = computed((): boolean => {
  return store.getUser?.subscription_status === 'cancelled'
})

const totalMessages = computed((): number => {
  if (store.getBillingUsage == null) {
    return 0
  }
  return store.getBillingUsage.sent_messages + store.getBillingUsage.received_messages
})

// --- Methods ---
async function loadData() {
  await Promise.all([
    store.loadUser(),
    store.loadBillingUsage(),
    store.loadBillingUsageHistory(),
  ])
  loading.value = false
  loadSubscriptionInvoices()
}

function loadSubscriptionInvoices() {
  loadingSubscriptionPayments.value = true
  store
    .indexSubscriptionPayments()
    .then((response: ResponsesUserSubscriptionPaymentsResponse) => {
      payments.value = response
    })
    .finally(() => {
      loadingSubscriptionPayments.value = false
    })
}

function generateInvoice() {
  errorMessages.value = new ErrorMessages()
  loading.value = true
  store
    .generateSubscriptionPaymentInvoice({
      subscriptionInvoiceId: selectedPayment.value?.id || '',
      request: {
        name: invoiceFormName.value,
        address: invoiceFormAddress.value,
        city: invoiceFormCity.value,
        state: invoiceFormState.value,
        zip_code: invoiceFormZipCode.value,
        country: invoiceFormCountry.value,
        notes: invoiceFormNotes.value,
      } as RequestsUserPaymentInvoice,
    })
    .then(() => {
      subscriptionInvoiceDialog.value = false
    })
    .catch((error: ErrorMessages) => {
      errorMessages.value = error
    })
    .finally(() => {
      loading.value = false
    })
}

function updateDetails() {
  loading.value = true
  store
    .getSubscriptionUpdateLink()
    .then((link: string) => {
      window.location.href = link
    })
    .catch(() => {
      loading.value = false
    })
}

function cancelPlan() {
  loading.value = true
  store
    .cancelSubscription()
    .then(() => {
      store.addNotification({
        message: 'Subscription cancelled successfully',
        type: 'success',
      })
      router.push('/')
    })
    .catch(() => {
      loading.value = false
    })
}

function showInvoiceDialog(payment: SubscriptionPayment) {
  selectedPayment.value = payment
  subscriptionInvoiceDialog.value = true
}

// --- Lifecycle ---
onMounted(async () => {
  await loadData()
})
</script>
