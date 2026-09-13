// Standalone filter functions (Vue.filter removed in Vue 3)
// Import these directly in components instead of using pipe syntax
import { intervalToDuration, formatDuration } from 'date-fns'
import { parsePhoneNumber, isValidPhoneNumber } from 'libphonenumber-js'

export const formatPhoneNumber = (value: string): string => {
  if (!isValidPhoneNumber(value)) return value
  const phoneNumber = parsePhoneNumber(value)
  return phoneNumber ? phoneNumber.formatInternational() : value
}

export const formatPhoneCountry = (value: string): string => {
  const phoneNumber = parsePhoneNumber(value)
  if (phoneNumber && phoneNumber.country) {
    const regionNames = new Intl.DisplayNames(undefined, { type: 'region' })
    return regionNames.of(phoneNumber.country) ?? 'Earth'
  }
  return 'Earth'
}

export const formatTimestamp = (value: string): string =>
  new Date(value).toLocaleString()

export const formatMoney = (value: string): string =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(
    parseInt(value),
  )

export const formatDecimal = (value: string): string =>
  new Intl.NumberFormat('en-US', { style: 'decimal' }).format(parseInt(value))

export const formatBillingPeriod = (value: string): string =>
  new Date(value).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
  } as Intl.DateTimeFormatOptions)

export const formatHumanizeTime = (value: string): string => {
  const durations = intervalToDuration({
    start: new Date(),
    end: new Date(value),
  })
  return formatDuration(durations)
}

export const formatCapitalize = (value: string): string =>
  value.charAt(0).toUpperCase() + value.slice(1)

// Nuxt plugin — provides $filters globally in templates
export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.config.globalProperties.$filters = {
    phoneNumber: formatPhoneNumber,
    phoneCountry: formatPhoneCountry,
    timestamp: formatTimestamp,
    money: formatMoney,
    decimal: formatDecimal,
    billingPeriod: formatBillingPeriod,
    humanizeTime: formatHumanizeTime,
    capitalize: formatCapitalize,
  }
})
