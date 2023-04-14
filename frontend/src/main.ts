import { createApp } from 'vue'
import App from '@/App.vue'
import router from '@/plugins/router'
import vuetify from '@/plugins/vuetify'
import pinia from '@/plugins/pinia'
import { loadFonts } from '@/plugins/webfontloader'

const app = createApp(App)
loadFonts()

app.use(router).use(vuetify).use(pinia)

app.mount('#app')
