<template>
  <v-btn
    :disabled="disabled"
    :color="color"
    :size="display.smAndDown.value ? 'small' : 'default'"
    :block="block"
    :large="large"
    @click="copy"
  >
    <v-icon start :icon="mdiContentCopy" />
    {{ copyText }}
  </v-btn>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { mdiContentCopy } from '@mdi/js'
import { useDisplay } from 'vuetify'
import { useAppStore } from '~/stores/app'

const props = withDefaults(defineProps<{
  value: string
  color?: string
  block?: boolean
  large?: boolean
  copyText?: string
  notificationText?: string
}>(), {
  color: 'default',
  block: false,
  large: false,
  copyText: 'Copy',
  notificationText: 'Copied',
})

const store = useAppStore()
const display = useDisplay()
const disabled = ref(false)

async function copy() {
  disabled.value = true
  await navigator.clipboard.writeText(props.value)

  store.addNotification({
    message: props.notificationText,
    type: 'success',
  })

  setTimeout(() => {
    disabled.value = false
  }, 5000)
}
</script>
