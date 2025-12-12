import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Импорты компонентов
import RouterPage from '@/routerPage.vue'
import MainPage from '@/components/pages/main/MainPage.vue'
import LoginPage from '@/components/pages/auth/LoginPage.vue'
import RegisterPage from '@/components/pages/auth/RegisterPage.vue'
import ProfilePage from '@/components/pages/profile/ProfilePage.vue'
import RatingPage from '@/components/pages/rating/RatingPage.vue'
import FAQPage from '@/components/pages/faq/FAQPage.vue'

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
    children: [
      { 
        path: '', 
        name: 'Home', 
        component: MainPage,
        meta: {
          title: 'Сапер Онлайн - Играй в Сапера с Друзьями',
          description: 'Играйте в Сапера онлайн с друзьями в реальном времени! Создавайте комнаты, соревнуйтесь и наслаждайтесь классической игрой Сапер в многопользовательском режиме.',
          keywords: 'сапер онлайн, сапер игра, minesweeper online, играть в сапера, сапер с друзьями, многопользовательский сапер'
        }
      },
      {
        path: 'room/:id',
        name: 'Room',
        component: MainPage,
        meta: {
          title: 'Комната - Сапер Онлайн',
          description: 'Подключитесь к игровой комнате Сапера',
          keywords: 'комната сапера, подключиться к комнате, играть в сапера'
        }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: ProfilePage,
        meta: {
          title: 'Профиль - Сапер Онлайн',
          description: 'Профиль пользователя и статистика игр',
          keywords: 'профиль, статистика, игры сапера'
        }
      },
      {
        path: 'profile/:username',
        name: 'UserProfile',
        component: ProfilePage,
        meta: {
          title: 'Профиль пользователя - Сапер Онлайн',
          description: 'Профиль пользователя и статистика игр',
          keywords: 'профиль, статистика, игры сапера'
        }
      },
      {
        path: 'rating',
        name: 'Rating',
        component: RatingPage,
        meta: {
          title: 'Рейтинг игроков - Сапер Онлайн',
          description: 'Рейтинг всех игроков по рейтинговым очкам',
          keywords: 'рейтинг, топ игроков, лидеры, сапер онлайн'
        }
      },
      {
        path: 'faq',
        name: 'FAQ',
        component: FAQPage,
        meta: {
          title: 'FAQ - Часто задаваемые вопросы - Сапер Онлайн',
          description: 'Ответы на часто задаваемые вопросы об игре Сапер Онлайн',
          keywords: 'faq, вопросы, помощь, сапер онлайн, как играть'
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
  const publicRoutes = ['/login', '/register', '/']
  const isPublicRoute = publicRoutes.includes(to.path) || to.path.startsWith('/room/')

  // Защищенные маршруты (требуют авторизации)
  const protectedRoutes = ['/profile']
  if (protectedRoutes.includes(to.path) && !isAuthenticated) {
    next('/login')
    return
  }

  // Страницы входа/регистрации - перенаправляем на главную, если уже авторизован
  if (isAuthenticated && (to.path === '/login' || to.path === '/register')) {
    next('/')
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
