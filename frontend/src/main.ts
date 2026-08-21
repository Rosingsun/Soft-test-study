import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import { setRouter } from './api/request'
import './style.css'

const app = createApp(App)
// OPT-22: 把 router 实例注入到 request 模块，让 401 走 SPA 路由跳转
setRouter(router)
app.use(createPinia())
app.use(router)
app.mount('#app')
