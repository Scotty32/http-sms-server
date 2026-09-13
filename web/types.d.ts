import { Framework } from 'vuetify'

export interface SelectItem {
  text: string
  value: string | number
}

declare module 'vue/types/vue' {
  interface Vue {
    $vuetify: Framework
  }
}
