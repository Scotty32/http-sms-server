<template>
  <v-app dark>
    <v-app-bar elevate-on-scroll color="#121212" height="70" fixed>
      <v-container>
        <v-row>
          <v-col class="w-full d-flex">
            <NuxtLink
              to="/"
              class="text-decoration-none d-flex"
              :class="{ 'mt-5': display.mdAndUp.value }"
            >
              <v-avatar tile size="33" class="mt-1">
                <v-img contain src="~/assets/img/logo.svg"></v-img>
              </v-avatar>
              <h3
                v-if="display.lgAndUp.value"
                class="text-h4 ml-1 text--primary"
              >
                httpSMS
              </h3>
            </NuxtLink>
            <v-spacer></v-spacer>
            <v-btn
              v-show="display.lgAndUp.value"
              size="large"
              variant="text"
              color="primary"
              class="my-5 mr-2"
              @click="goToPricing"
            >
              Pricing
            </v-btn>
            <v-btn
              v-show="display.lgAndUp.value"
              size="large"
              variant="text"
              color="primary"
              class="my-5 mr-2"
              :to="{ name: 'blog' }"
            >
              Blog
            </v-btn>
            <v-btn
              v-show="
                display.lgAndUp.value &&
                store.getAuthUser === null
              "
              size="large"
              variant="text"
              color="primary"
              class="my-5 mr-2"
              :to="{ name: 'login' }"
            >
              Login
            </v-btn>
            <v-btn
              v-show="store.getAuthUser === null"
              exact-path
              color="primary"
              :class="{
                'mt-5': display.mdAndUp.value,
                'mt-1': !display.mdAndUp.value,
              }"
              :size="display.lgAndUp.value ? 'large' : 'default'"
              :to="{ name: 'login' }"
            >
              Get Started
              <span v-show="display.lgAndUp.value">&nbsp;For Free</span>
            </v-btn>
            <v-btn
              v-show="store.getAuthUser !== null"
              exact-path
              color="primary"
              :class="{
                'mt-5': display.mdAndUp.value,
                'mt-1': !display.mdAndUp.value,
              }"
              :size="display.lgAndUp.value ? 'large' : 'default'"
              :to="{ name: 'threads' }"
            >
              Dashboard
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-app-bar>
    <v-main>
      <Toast />
      <NuxtPage />
    </v-main>
    <v-footer class="pt-4">
      <v-container>
        <v-row>
          <v-col cols="12" md="3">
            <NuxtLink to="/" class="text-decoration-none d-flex">
              <v-avatar tile size="33" class="mt-1">
                <v-img contain src="~/assets/img/logo.svg"></v-img>
              </v-avatar>
              <h3 class="text-h4 ml-1 text--primary">httpSMS</h3>
            </NuxtLink>
            <div class="subtitle-2 mb-4 text--secondary">
              Made With <v-icon color="#cf1112">{{ mdiHeart }}</v-icon> in
              Tallinn
              <v-img
                class="d-inline-block"
                width="20"
                src="https://upload.wikimedia.org/wikipedia/commons/8/8f/Flag_of_Estonia.svg"
              ></v-img>
            </div>
            <p class="mt-n3">
              <v-btn href="https://twitter.com/httpsmsHQ" icon color="#1DA1F2">
                <v-icon>{{ mdiTwitter }}</v-icon>
              </v-btn>
              <v-btn
                :href="store.getAppData.githubUrl"
                icon
                size="large"
                color="#ffffff"
              >
                <v-icon>{{ mdiGithub }}</v-icon>
              </v-btn>
              <v-btn
                href="https://discord.gg/kGk8HVqeEZ"
                icon
                size="large"
                color="#5865f2"
              >
                <v-img
                  contain
                  height="24"
                  width="24"
                  src="~/assets/img/discord-logo-blue.svg"
                ></v-img>
              </v-btn>
            </p>
            <a
              href="https://www.saashub.com/httpsms?utm_source=badge&utm_campaign=badge&utm_content=httpsms&badge_variant=color&badge_kind=approved"
              target="_blank"
            >
              <img
                src="https://cdn-b.saashub.com/img/badges/approved-color.png?v=1"
                alt="httpSMS badge"
                style="max-width: 150px"
              />
            </a>
          </v-col>
          <v-col cols="12" md="3">
            <h2 class="text-h6 mb-2">Resources</h2>
            <ul style="list-style: none" class="pa-0">
              <li class="mb-2">
                <v-hover v-slot="{ isHovering }">
                  <a
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                    @click.stop="goToPricing"
                  >
                    Pricing
                    <v-icon size="small">{{ mdiCreditCardOutline }}</v-icon>
                  </a>
                </v-hover>
              </li>
              <li class="mb-2">
                <v-hover v-slot="{ isHovering }">
                  <a
                    href="https://httpsms.lemonsqueezy.com/affiliates"
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                  >
                    Affiliates
                    <v-icon color="warning" size="small">{{ mdiShieldStar }}</v-icon>
                  </a>
                </v-hover>
              </li>
              <li class="mb-2">
                <v-hover v-slot="{ isHovering }">
                  <a
                    href="https://status.httpsms.com"
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                  >
                    Site status
                    <v-icon color="success" size="x-small">{{ mdiCircle }}</v-icon>
                  </a>
                </v-hover>
              </li>
              <li class="mb-2">
                <v-hover v-slot="{ isHovering }">
                  <NuxtLink
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                    to="/blog"
                  >
                    Blog
                    <v-icon size="small">{{ mdiPost }}</v-icon>
                  </NuxtLink>
                </v-hover>
              </li>
            </ul>
          </v-col>
          <v-col cols="12" md="3">
            <h2 class="text-h6 mb-2">Developers</h2>
            <ul style="list-style: none" class="pa-0">
              <li class="mb-2">
                <v-hover v-slot="{ isHovering }">
                  <a
                    :href="store.getAppData.documentationUrl"
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                  >
                    Documentation
                    <v-icon size="small">{{ mdiBookOpenVariant }}</v-icon>
                  </a>
                </v-hover>
              </li>
              <li class="mb-2">
                <v-hover v-slot="{ isHovering }">
                  <a
                    :href="store.getAppData.githubUrl"
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                  >
                    Github
                    <v-icon size="small">{{ mdiGithub }}</v-icon>
                  </a>
                </v-hover>
              </li>
              <li class="mb-2">
                <v-hover v-slot="{ isHovering }">
                  <a
                    href="https://sandbox.httpsms.com"
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                  >
                    Sandbox
                    <v-icon size="small" color="pink">{{ mdiCreation }}</v-icon>
                  </a>
                </v-hover>
              </li>
              <li class="mb-2">
                <v-hover v-slot="{ isHovering }">
                  <a
                    href="https://httpsms.featurebase.app"
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                  >
                    Request Feature
                    <v-icon size="small" color="yellow">{{ mdiLightbulbOn50 }}</v-icon>
                  </a>
                </v-hover>
              </li>
            </ul>
          </v-col>
          <v-col cols="12" md="3">
            <h2 class="text-h6 mb-2">Legal</h2>
            <ul style="list-style: none" class="pa-0">
              <li class="mb-2">
                <v-hover v-slot="{ isHovering }">
                  <NuxtLink
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                    to="/terms-and-conditions"
                  >
                    Terms & Conditions
                    <v-icon size="small">{{ mdiScaleBalance }}</v-icon>
                  </NuxtLink>
                </v-hover>
              </li>
              <li>
                <v-hover v-slot="{ isHovering }">
                  <NuxtLink
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                    to="/privacy-policy"
                  >
                    Privacy Policy
                    <v-icon size="small">{{ mdiEyeOffOutline }}</v-icon>
                  </NuxtLink>
                </v-hover>
              </li>
              <li class="mt-2">
                <v-hover v-slot="{ isHovering }">
                  <a
                    class="text--primary text-decoration-none"
                    :class="{ 'text-decoration-underline': isHovering }"
                    href="mailto:support@httpsms.com"
                  >
                    Contact Support
                    <v-icon size="small">{{ mdiEmailOutline }}</v-icon>
                  </a>
                </v-hover>
              </li>
            </ul>
          </v-col>
        </v-row>
      </v-container>
    </v-footer>
  </v-app>
</template>

<script setup lang="ts">
import {
  mdiLoginVariant,
  mdiArrowRight,
  mdiGithub,
  mdiCircle,
  mdiTwitter,
  mdiHeart,
  mdiShieldStar,
  mdiLightbulbOn50,
  mdiCreation,
  mdiDomain,
  mdiEyeOffOutline,
  mdiPost,
  mdiCreditCardOutline,
  mdiScaleBalance,
  mdiEmailOutline,
  mdiBookOpenVariant,
} from '@mdi/js'
import { useDisplay } from 'vuetify'

const store = useAppStore()
const route = useRoute()
const router = useRouter()
const display = useDisplay()

function goToPricing() {
  if (route.name === 'index') {
    document.querySelector('#pricing')?.scrollIntoView({ behavior: 'smooth' })
  } else {
    router.push('/#pricing')
  }
}
</script>

<style lang="scss">
.v-application {
  .logo-badge {
    .v-badge__wrapper {
      span {
        margin-bottom: -8px;
      }
    }
  }
}
</style>
