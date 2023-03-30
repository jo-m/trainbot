import { createApp } from 'vue'
import App from '@/App.vue'
import router from '@/plugins/router'
import vuetify from '@/plugins/vuetify'
import { loadFonts } from '@/plugins/webfontloader'

const app = createApp(App)
loadFonts()

app.use(router).use(vuetify)

app.mount('#app')
