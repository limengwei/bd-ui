import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import en from './en'

const saved = localStorage.getItem('beads-ui.locale') || 'zh-CN'

const i18n = createI18n({
  legacy: false,
  locale: saved,
  fallbackLocale: 'en',
  messages: {
    'zh-CN': zhCN,
    en,
  },
})

export default i18n
