export default defineNuxtRouteMiddleware(async () => {
  const store = useAppStore()
  if (store.getAuthUser !== null) {
    return navigateTo('/threads')
  }
  try {
    await store.loadUser()
    return navigateTo('/threads')
  } catch {
    // not authenticated — stay on login page
  }
})
