import { createApp } from 'vue'
import { createPinia } from 'pinia'
import VueGtag from 'vue-gtag-next'

import App from './App.vue'
import router from './router'
import './style.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(VueGtag, {
  property: {
    id: '', // Will be set dynamically when settings are loaded
  },
})

app.mount('#app')
