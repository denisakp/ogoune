import { createApp } from 'vue'
import { createPinia } from 'pinia'
import VueGtag from 'vue-gtag-next'

import StatusApp from './StatusApp.vue'
import statusRouter from './router/status-router'
import './style.css'

const app = createApp(StatusApp)

app.use(createPinia())
app.use(statusRouter)
app.use(VueGtag, {
  property: {
    id: '',
  },
})

app.mount('#app')
