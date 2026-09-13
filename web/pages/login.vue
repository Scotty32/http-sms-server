<template>
  <v-container fluid style="height: 100vh">
    <v-row align="center" justify="center" style="height: 100%">
      <v-col
        cols="12"
        md="4"
        xl="3"
        :class="{ 'mt-n16': display.lgAndUp.value }"
      >
        <div class="text-center mb-5">
          <v-avatar tile size="45" class="mt-n8 mr-4">
            <v-img contain src="@/assets/img/logo.svg" />
          </v-avatar>
          <span class="text-h3">Welcome</span>
        </div>
        <v-card max-width="360" class="mx-auto">
          <v-card-text>
            <v-tabs v-model="tab" align-tabs="center">
              <v-tab value="login">Login</v-tab>
              <v-tab value="register">Register</v-tab>
            </v-tabs>
            <v-window v-model="tab">
              <v-window-item value="login">
                <v-form class="mt-4" @submit.prevent="onLogin">
                  <v-text-field
                    v-model="email"
                    label="Email"
                    type="email"
                    variant="outlined"
                    density="compact"
                    :disabled="loading"
                  />
                  <v-text-field
                    v-model="password"
                    label="Password"
                    type="password"
                    variant="outlined"
                    density="compact"
                    :disabled="loading"
                  />
                  <v-alert v-if="error" type="error" density="compact" class="mb-3">
                    {{ error }}
                  </v-alert>
                  <v-btn type="submit" color="primary" block :loading="loading">
                    Login
                  </v-btn>
                </v-form>
              </v-window-item>
              <v-window-item value="register">
                <v-form class="mt-4" @submit.prevent="onRegister">
                  <v-text-field
                    v-model="name"
                    label="Name"
                    variant="outlined"
                    density="compact"
                    :disabled="loading"
                  />
                  <v-text-field
                    v-model="email"
                    label="Email"
                    type="email"
                    variant="outlined"
                    density="compact"
                    :disabled="loading"
                  />
                  <v-text-field
                    v-model="password"
                    label="Password"
                    type="password"
                    variant="outlined"
                    density="compact"
                    :disabled="loading"
                  />
                  <v-alert v-if="error" type="error" density="compact" class="mb-3">
                    {{ error }}
                  </v-alert>
                  <v-btn type="submit" color="primary" block :loading="loading">
                    Create Account
                  </v-btn>
                </v-form>
              </v-window-item>
            </v-window>

          </v-card-text>
        </v-card>
        <div class="text-center mt-4">
          <BackButton />
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { useDisplay } from 'vuetify'

definePageMeta({ middleware: ['guest'] })
useHead({ title: 'Login To Your Account - Http SMS' })

const store = useAppStore()
const router = useRouter()
const route = useRoute()
const display = useDisplay()

const tab = ref('login')
const name = ref('')
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

async function onLogin() {
  error.value = null
  loading.value = true
  try {
    await store.login({ email: email.value, password: password.value })
    const to = (route.query.to as string) || '/threads'
    router.push(to)
  } catch (e: any) {
    error.value = e?.response?.data?.message || 'Invalid email or password'
  } finally {
    loading.value = false
  }
}

async function onRegister() {
  error.value = null
  loading.value = true
  try {
    await store.register({ name: name.value, email: email.value, password: password.value })
    router.push('/threads')
  } catch (e: any) {
    error.value = e?.response?.data?.message || 'Registration failed'
  } finally {
    loading.value = false
  }
}
</script>
