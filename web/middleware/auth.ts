export default defineNuxtRouteMiddleware(async (to) => {
  const store = useAppStore()

  // Only fetch user data if not already loaded in the store
  if (store.getAuthUser === null) {
    try {
      await store.loadUser()
    } catch (err: any) {
      if (err?.response?.status === 401) {
        return navigateTo(`/login?to=${to.path}`)
      }
    }
  }

  // Load phones if not yet fetched (guard in loadPhones prevents refetch)
  if (store.getAuthUser !== null) {
    store.loadPhones().catch(() => {})
  }
})
