import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { UNAUTHORIZED_EVENT } from './api/http'
import { useIdentityStore } from './stores/identity'
import './style.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)

// 拦截到后端 401 时，清会话并跳登录；避免用户卡在过期 token 的状态。
window.addEventListener(UNAUTHORIZED_EVENT, () => {
  const identity = useIdentityStore(pinia)
  if (!identity.isLoggedIn) return
  identity.signOut()
  const current = router.currentRoute.value
  if (current.meta?.public) return
  router.replace({
    path: '/login',
    query: current.fullPath !== '/' ? { redirect: current.fullPath } : {},
  })
})

app.mount('#app')
