<template>
  <v-btn
    color="default"
    :size="display.smAndDown.value ? 'small' : 'default'"
    :block="block"
    @click="goBack"
  >
    <v-icon :icon="mdiArrowLeft" />
    Go Back
  </v-btn>
</template>

<script setup lang="ts">
import { mdiArrowLeft } from '@mdi/js'
import type { RouteLocationRaw } from 'vue-router'
import { useDisplay } from 'vuetify'

const props = defineProps<{
  route?: RouteLocationRaw
  block?: boolean
}>()

const router = useRouter()
const display = useDisplay()

function goBack() {
  if (props.route) {
    router.push(props.route)
    return
  }
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.push({ name: 'index' })
}
</script>
