// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,

  app: {
    head: {
      titleTemplate: '%s',
      title: 'Convert your android phone into an SMS gateway - httpSMS',
      htmlAttrs: { lang: 'en' },
      script: [
        { src: '/integrations.js', async: true, defer: true },
        {
          src: 'https://lmsqueezy.com/affiliate.js',
          async: true,
          defer: true,
        },
        {
          src: 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit',
        },
      ],
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        {
          name: 'description',
          content:
            'Use your android phone to send and receive SMS messages using a simple HTTP API.',
        },
        { name: 'format-detection', content: 'telephone=no' },
        { name: 'twitter:site', content: '@NdoleStudio' },
        { name: 'twitter:card', content: 'summary_large_image' },
        {
          name: 'og:title',
          content:
            'Convert your android phone into an SMS gateway - httpSMS',
        },
        {
          name: 'og:description',
          content:
            'Use your android phone to send and receive SMS messages using a simple HTTP API.',
        },
        { name: 'og:image', content: 'https://httpsms.com/header.png' },
      ],
      link: [{ rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }],
    },
  },

  runtimeConfig: {
    public: {
      apiBaseUrl: process.env.API_BASE_URL || 'http://localhost:8000',
      appUrl: process.env.APP_URL || '',
      appName: process.env.APP_NAME || 'httpSMS',
      appEnv: process.env.APP_ENV || 'local',
      appDownloadUrl: process.env.APP_DOWNLOAD_URL || '',
      appDocumentationUrl: process.env.APP_DOCUMENTATION_URL || '',
      appGithubUrl: process.env.APP_GITHUB_URL || '',
      checkoutURL: process.env.CHECKOUT_URL || '',
      enterpriseCheckoutURL: process.env.ENTERPRISE_CHECKOUT_URL || '',
      cloudflareTurnstileSiteKey:
        process.env.CLOUDFLARE_TURNSTILE_SITE_KEY || '',
      pusherKey: process.env.PUSHER_KEY || '',
      pusherCluster: process.env.PUSHER_CLUSTER || '',
    },
  },

  modules: ['@pinia/nuxt', 'vuetify-nuxt-module'],

  vuetify: {
    vuetifyOptions: {
      theme: {
        defaultTheme: 'dark',
      },
      icons: {
        defaultSet: 'mdi-svg',
      },
    },
  },

  build: {
    transpile: ['chart.js', 'vue-chartjs'],
  },

  compatibilityDate: '2025-05-22',
})
