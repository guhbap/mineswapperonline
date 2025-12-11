import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Импорты компонентов
import RouterPage from '@/routerPage.vue'
import MainPage from '@/components/pages/main/MainPage.vue'
import LoginPage from '@/components/pages/auth/LoginPage.vue'
import RegisterPage from '@/components/pages/auth/RegisterPage.vue'

// Типизированный массив маршрутов
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: LoginPage
  },
  {
    path: '/register',
    name: 'Register',
    component: RegisterPage
  },
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

// Защита роутов
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const isAuthenticated = authStore.isAuthenticated

  // Публичные маршруты (доступны без авторизации)
  const publicRoutes = ['/login', '/register', '/main']
  const isPublicRoute = publicRoutes.includes(to.path)

  // Страницы входа/регистрации - перенаправляем на главную, если уже авторизован
  if (isAuthenticated && (to.path === '/login' || to.path === '/register')) {
    next('/main')
  } else {
    // Все остальные маршруты доступны
    next()
  }
})

export default router
