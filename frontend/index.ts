import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

// Импорты компонентов
import RouterPage from '@/routerPage.vue'
import MainPage from '@/components/pages/main/MainPage.vue'

// Типизированный массив маршрутов
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: RouterPage,
    redirect: '/main',
    children: [
      { path: '/main', name: 'Main', component: MainPage }
    ]
  }
]

// Создание роутера с типами
const router = createRouter({
  history: createWebHistory('/'),
  routes
})

export default router
