import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ui from '@nuxt/ui/vue-plugin'

import StatusApp from './StatusApp.vue'
import statusRouter from './router/status-router'
import './style.css'

const app = createApp(StatusApp)

app.use(createPinia())
app.use(statusRouter)
app.use(ui)

app.mount('#app')
