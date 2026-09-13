<template>
  <v-container fluid class="pa-0" :style="{ height: display.lgAndUp.value ? '100vh' : 'auto' }">
    <div class="w-full h-full">
      <v-app-bar height="60" :density="display.mdAndDown.value ? 'compact' : 'default'" location="top">
        <v-btn :icon="mdiArrowLeft" to="/threads" />
        <v-toolbar-title>
          Heartbeats
          <v-icon size="x-small" class="mx-2" color="primary" :icon="mdiCircle" />
          <span v-if="store.getOwner">{{ formatPhoneNumber(store.getOwner) }}</span>
        </v-toolbar-title>
      </v-app-bar>
      <v-container class="mt-16">
        <v-row>
          <v-col cols="12">
            <p>
              Every 15 minutes, the httpSMS app on your Android phone sends a
              heartbeat event to the httpsms API to show that it is alive.
            </p>
          </v-col>
          <v-col v-if="display.mdAndUp.value" cols="12" class="px-0">
            <BarChart :data="chartData" :options="chartOptions" />
          </v-col>
          <v-col cols="12">
            <v-data-table
              :headers="dataTableHeaders"
              :items="dataTableItems"
              :sort-by="[{ key: 'timestamp', order: 'desc' }]"
              :items-per-page="100"
              class="heartbeat--table"
            >
              <template #[`item.interval`]="{ item }">
                {{ formatDurationItem(item.interval) }}
              </template>
              <template #[`item.owner`]="{ item }">
                {{ formatPhoneNumber(item.owner) }}
              </template>
              <template #[`item.timestamp`]="{ item }">
                {{ formatTimestamp(item.timestamp) }}
              </template>
            </v-data-table>
          </v-col>
        </v-row>
      </v-container>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'
import { mdiArrowLeft, mdiCircle } from '@mdi/js'
import 'chartjs-adapter-moment'
import { formatDuration, intervalToDuration } from 'date-fns'
import { Bar as BarChart } from 'vue-chartjs'
import {
  Chart as ChartJS,
  BarElement,
  CategoryScale,
  LinearScale,
  TimeScale,
  Tooltip,
  Legend,
} from 'chart.js'
import { formatPhoneNumber, formatTimestamp } from '~/plugins/filters'

ChartJS.register(BarElement, CategoryScale, LinearScale, TimeScale, Tooltip, Legend)

definePageMeta({ middleware: ['auth'] })
useHead({ title: 'Heartbeats - Http SMS' })

const store = useAppStore()
const display = useDisplay()

const heartbeats = ref<any[]>([])

const dataTableHeaders = [
  { title: 'HEARTBEAT ID', key: 'id', sortable: false },
  { title: 'PHONE NUMBER', key: 'owner', sortable: false },
  { title: 'RECEIVED AT', key: 'timestamp' },
  { title: 'TIME INTERVAL', key: 'interval' },
]

const dataTableItems = computed(() =>
  heartbeats.value.map((heartbeat, index) => {
    let interval = 0
    if (index < heartbeats.value.length - 1) {
      interval = getDiff(heartbeat.timestamp, heartbeats.value[index + 1].timestamp)
    }
    return { id: heartbeat.id, timestamp: heartbeat.timestamp, owner: heartbeat.owner, interval }
  }),
)

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label(context: any) {
          if (context.dataIndex === context.dataset.data.length - 1) return '-'
          const duration = intervalToDuration({
            start: new Date(context.dataset.data[context.dataIndex + 1].x),
            end: new Date(context.dataset.data[context.dataIndex].x),
          })
          return formatDuration(duration)
        },
      },
    },
  },
  scales: {
    x: { type: 'time' },
    y: { display: false },
  },
}))

const chartData = computed(() => {
  const data = heartbeats.value.map((h) => ({
    x: new Date(h.timestamp).toISOString(),
    y: 1,
  }))
  if (!data.length) return { datasets: [{ data, backgroundColor: '#2196f3' }] }

  let prev = new Date(data[0].x)
  const newData = []
  for (let i = 1; i < data.length; i++) {
    const current = new Date(data[i].x)
    const diff = prev.getTime() - current.getTime()
    if (diff > 600000) {
      newData.push(data[i])
      prev = current
    }
  }
  return { datasets: [{ data: newData, backgroundColor: '#2196f3' }] }
})

function getDiff(a: string, b: string): number {
  return new Date(a).getTime() - new Date(b).getTime()
}

function formatDurationItem(duration: number): string {
  if (duration === 0) return '-'
  const start = new Date()
  start.setMilliseconds(start.getMilliseconds() + duration)
  return formatDuration(intervalToDuration({ start: new Date(), end: start })) || '0 seconds'
}

onMounted(async () => {
  await store.loadUser()
  await store.loadPhones()
  store.getHeartbeatAction(100).then((h) => { heartbeats.value = h })
})
</script>

<style lang="scss">
.v-application {
  .heartbeat--table.v-data-table tbody tr.v-data-table__selected {
    background: #b71c1c;
  }
}
</style>
