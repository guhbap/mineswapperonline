import 'air-datepicker/air-datepicker.css'
// import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'
import 'choices.js/public/assets/styles/choices.css'
import 'vue-final-modal/style.css'

import { createApp, type App as VueApp } from 'vue'
import { createVfm } from 'vue-final-modal'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from '../index.js'
import { useAuthStore } from './stores/auth'

const app: VueApp = createApp(App)
// app.config.devtools = false
app.use(createVfm())
const pinia = createPinia()
app.use(pinia)

app.use(router)

// Инициализация auth store
const authStore = useAuthStore()
authStore.init()

app.mount('#app')
