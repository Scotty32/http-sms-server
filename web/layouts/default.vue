<template>
  <v-app>
    <v-divider v-if="store.isLocal" class="py-1 bg-warning"></v-divider>
    <v-navigation-drawer
      v-if="display.lgAndUp.value && hasDrawer"
      :width="400"
      permanent
    >
      <template #prepend>
        <v-divider v-if="store.isLocal" class="py-1 bg-warning"></v-divider>
        <MessageThreadHeader />
        <div class="overflow-y-auto v-navigation-drawer__message-thread">
          <MessageThread />
        </div>
      </template>
    </v-navigation-drawer>
    <v-main :class="{ 'has-drawer': hasDrawer && display.lgAndUp.value }">
      <Toast />
      <NuxtPage v-if="store.isAuthStateChanged" />
      <LoadingDashboard v-else />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'
import Pusher from 'pusher-js'

const store = useAppStore()
const route = useRoute()
const display = useDisplay()
const config = useRuntimeConfig()

const hasDrawer = computed(() =>
  ['threads', 'threads-id'].includes(route.name as string ?? ''),
)

let poller: ReturnType<typeof setInterval> | null = null
let canPoll = false

onMounted(() => {
  setTimeout(() => {
    if (!store.getAuthUser) return

    const pusher = new Pusher(config.public.pusherKey as string, {
      cluster: config.public.pusherCluster as string,
    })

    const channel = pusher.subscribe(store.getAuthUser.id)
    channel.bind('phone.updated', () => {
      canPoll = true
    })

    startPoller()
  }, 10_000)
})

onBeforeUnmount(() => {
  if (poller) clearInterval(poller)
})

function startPoller() {
  poller = setInterval(async () => {
    if (!canPoll || store.getAuthUser == null) return

    store.setPolling(true)

    const promises: Promise<any>[] = []
    if (store.getAuthUser && store.getOwner) {
      promises.push(
        store.loadPhones(true),
        store.loadThreads(),
        store.getHeartbeatAction(),
      )
    }
    canPoll = false
    await Promise.all(promises)

    setTimeout(() => store.setPolling(false), 1000)
  }, 10_000)
}
</script>

<style lang="scss">
.v-application {
  .w-full {
    width: 100%;
  }
  .h-full {
    height: 100%;
  }

  .has-drawer {
    .v-snackbar {
      padding-left: 400px;
    }
  }

  .v-navigation-drawer__message-thread {
    height: calc(100vh - 120px);

    &::-webkit-scrollbar {
      width: 8px;
    }
    &::-webkit-scrollbar-track {
      background: #363636;
    }
    &::-webkit-scrollbar-thumb {
      background: #666666;
      border-radius: 8px;
    }
  }

  code.hljs {
    font-size: 16px;
  }
}

.feedback-btn {
  position: fixed;
  z-index: 15;
  right: -56px;
  margin: 0;
  top: 45%;
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
  transform: rotate(-90deg);
}
</style>
