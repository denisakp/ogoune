import { createApp } from 'vue'
import { createPinia } from 'pinia'
import VueGtag from 'vue-gtag-next'
import ui from '@nuxt/ui/vue-plugin'

import App from './App.vue'
import router from './router'
import './style.css'
import { loadRuntimeConfig } from '@/composables/useRuntimeConfig'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ui)
app.use(VueGtag, {
  property: {
    id: '', // Will be set dynamically when settings are loaded
  },
})

// Prefetch runtime config (SSL provider + edition) so first paint can render
// the correct UI wording (spec 059 FR-030). Failure falls back to safe defaults.
loadRuntimeConfig().finally(() => {
  app.mount('#app')
})
