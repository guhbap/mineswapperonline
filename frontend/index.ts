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
    component: LoginPage,
    meta: {
      title: 'Вход - Сапер Онлайн',
      description: 'Войдите в свой аккаунт, чтобы играть в Сапера онлайн с друзьями',
      keywords: 'вход, авторизация, сапер онлайн, войти в игру'
    }
  },
  {
    path: '/register',
    name: 'Register',
    component: RegisterPage,
    meta: {
      title: 'Регистрация - Сапер Онлайн',
      description: 'Зарегистрируйтесь, чтобы играть в Сапера онлайн и сохранять свой прогресс',
      keywords: 'регистрация, создать аккаунт, сапер онлайн, зарегистрироваться'
    }
  },
  {
    path: '/',
    component: RouterPage,
    redirect: '/main',
    children: [
      { 
        path: '/main', 
        name: 'Main', 
        component: MainPage,
        meta: {
          title: 'Сапер Онлайн - Играй в Сапера с Друзьями',
          description: 'Играйте в Сапера онлайн с друзьями в реальном времени! Создавайте комнаты, соревнуйтесь и наслаждайтесь классической игрой Сапер в многопользовательском режиме.',
          keywords: 'сапер онлайн, сапер игра, minesweeper online, играть в сапера, сапер с друзьями, многопользовательский сапер'
        }
      }
    ]
  }
]

// Создание роутера с типами
const router = createRouter({
  history: createWebHistory('/'),
  routes
})

// Защита роутов и обновление мета-тегов
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const isAuthenticated = authStore.isAuthenticated

  // Публичные маршруты (доступны без авторизации)
  const publicRoutes = ['/login', '/register', '/main']
  const isPublicRoute = publicRoutes.includes(to.path)

  // Страницы входа/регистрации - перенаправляем на главную, если уже авторизован
  if (isAuthenticated && (to.path === '/login' || to.path === '/register')) {
    next('/main')
    return
  }

  // Обновление мета-тегов для SEO
  if (to.meta.title) {
    document.title = to.meta.title as string
  }

  // Обновление meta description
  let metaDescription = document.querySelector('meta[name="description"]')
  if (!metaDescription) {
    metaDescription = document.createElement('meta')
    metaDescription.setAttribute('name', 'description')
    document.head.appendChild(metaDescription)
  }
  if (to.meta.description) {
    metaDescription.setAttribute('content', to.meta.description as string)
  }

  // Обновление Open Graph тегов
  const ogTitle = document.querySelector('meta[property="og:title"]')
  if (ogTitle && to.meta.title) {
    ogTitle.setAttribute('content', to.meta.title as string)
  }

  const ogDescription = document.querySelector('meta[property="og:description"]')
  if (ogDescription && to.meta.description) {
    ogDescription.setAttribute('content', to.meta.description as string)
  }

  const ogUrl = document.querySelector('meta[property="og:url"]')
  if (ogUrl) {
    ogUrl.setAttribute('content', window.location.origin + to.fullPath)
  }

  // Обновление canonical URL
  let canonical = document.querySelector('link[rel="canonical"]')
  if (!canonical) {
    canonical = document.createElement('link')
    canonical.setAttribute('rel', 'canonical')
    document.head.appendChild(canonical)
  }
  canonical.setAttribute('href', window.location.origin + to.fullPath)

  next()
})

export default router
