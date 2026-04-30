import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/boards',
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/pages/LoginPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/pages/RegisterPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/boards',
      name: 'Boards',
      component: () => import('@/pages/BoardsPage.vue'),
      meta: { auth: true },
    },
    {
      path: '/boards/:id',
      name: 'BoardDetail',
      component: () => import('@/pages/BoardDetailPage.vue'),
      meta: { auth: true },
    },
    {
      path: '/members',
      name: 'Members',
      component: () => import('@/pages/MembersPage.vue'),
      meta: { auth: true },
    },
    {
      path: '/profile',
      name: 'Profile',
      component: () => import('@/pages/ProfilePage.vue'),
      meta: { auth: true },
    },
  ],
})

router.beforeEach((to) => {
  const { isAuthenticated } = useAuth()
  if (to.meta.auth && !isAuthenticated.value) {
    return { name: 'Login' }
  }
  if (to.meta.guest && isAuthenticated.value) {
    return { name: 'Boards' }
  }
})

export default router
