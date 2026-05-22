import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import i18n from './locales'

const app = createApp(App)

const epLocales = { 'zh-CN': zhCn, en }
app.use(ElementPlus, { locale: epLocales[i18n.global.locale.value] || en })
app.use(createPinia())
app.use(router)
app.use(i18n)
app.mount('#app')
