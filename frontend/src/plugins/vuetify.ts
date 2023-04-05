import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import { createVuetify, type ThemeDefinition } from 'vuetify'

const light: ThemeDefinition = {
  dark: false,
  colors: {
    primary: '#1867C0',
    secondary: '#5CBBF6'
  }
}

const dark: ThemeDefinition = {
  dark: true,
  colors: {
    primary: '#212121',
    secondary: '#F4511E'
  }
}

export default createVuetify({
  theme: {
    themes: {
      light,
      dark
    }
  }
})
