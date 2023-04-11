import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import { createVuetify, type ThemeDefinition } from 'vuetify'

const light: ThemeDefinition = {
  dark: false,
  colors: {
    primary: '#0D47A1',
    secondary: '#DD2C00'
  }
}

const dark: ThemeDefinition = {
  dark: true,
  colors: {
    primary: '#131b26',
    secondary: '#E53935'
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
