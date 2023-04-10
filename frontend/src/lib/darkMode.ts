import { ref } from 'vue'

const matcher = window.matchMedia('(prefers-color-scheme: dark)')
const browserInDarkMode = ref<boolean>(matcher.matches)

matcher.onchange = (ev) => {
  browserInDarkMode.value = ev.matches
}

export default browserInDarkMode
