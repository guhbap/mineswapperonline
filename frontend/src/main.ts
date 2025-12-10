import 'air-datepicker/air-datepicker.css'
// import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'
import 'choices.js/public/assets/styles/choices.css'
import 'vue-final-modal/style.css'

import { createApp, type App as VueApp } from 'vue'
import { createVfm } from 'vue-final-modal'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from '../index.js'

const app: VueApp = createApp(App)
// app.config.devtools = false
app.use(createVfm())
app.use(createPinia())

app.use(router)
app.mount('#app')
