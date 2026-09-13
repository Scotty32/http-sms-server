export default defineNuxtPlugin(async () => {
  const store = useAppStore()
  store.restoreAuthToken()
  try {
    await store.loadUser()
    store.loadPhones().catch(() => {})
  } catch {
    // Not authenticated — store stays empty
  }
})
