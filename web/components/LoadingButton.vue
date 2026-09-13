<template>
  <v-btn
    :block="block"
    :type="type"
    :size="btnSize"
    :color="color"
    :variant="text ? 'text' : undefined"
    :rounded="tile ? '0' : undefined"
    exact
    :disabled="isLoading"
    @click.prevent="onClick"
  >
    <v-progress-circular
      v-if="isClicked"
      :size="small ? 20 : 25"
      color="grey"
      class="mr-2"
      indeterminate
    ></v-progress-circular>
    <v-icon v-if="icon && !isLoading" start :icon="icon" />
    <slot></slot>
  </v-btn>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = withDefaults(defineProps<{
  type?: string
  block?: boolean
  large?: boolean
  xLarge?: boolean
  tile?: boolean
  text?: boolean
  small?: boolean
  color?: string
  icon?: string | null
}>(), {
  type: 'submit',
  block: false,
  large: false,
  xLarge: false,
  tile: false,
  text: false,
  small: false,
  color: 'primary',
  icon: null,
})

const isLoading = defineModel<boolean>('loading', { required: true })

const isClicked = ref(false)

const btnSize = computed(() => {
  if (props.xLarge) return 'x-large'
  if (props.large) return 'large'
  if (props.small) return 'small'
  return 'default'
})

const emit = defineEmits<{
  click: []
}>()

watch(isLoading, (submitting) => {
  if (!submitting && isClicked.value) {
    isClicked.value = false
  }
})

function onClick() {
  isClicked.value = true
  emit('click')
}
</script>
