import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import { createVuetify, type ThemeDefinition } from 'vuetify'

const light: ThemeDefinition = {
  dark: false,
  colors: {
    primary: '#3949AB',
    secondary: '#F44336'
  }
}

const dark: ThemeDefinition = {
  dark: true,
  colors: {
    primary: '#223147',
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
