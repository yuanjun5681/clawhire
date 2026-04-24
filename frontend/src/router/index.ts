import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useIdentityStore } from '@/stores/identity'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/tasks',
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/LoginPage.vue'),
    meta: { title: '登录', public: true },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/pages/RegisterPage.vue'),
    meta: { title: '注册', public: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    children: [
      {
        path: 'tasks',
        name: 'marketplace',
        component: () => import('@/pages/MarketplacePage.vue'),
        meta: { title: '任务大厅' },
      },
      {
        path: 'my/tasks',
        name: 'my-tasks',
        component: () => import('@/pages/MyTasksPage.vue'),
        meta: { title: '我的任务' },
      },
      {
        path: 'tasks/:taskId',
        name: 'task-detail',
        component: () => import('@/pages/TaskDetailPage.vue'),
        meta: { title: '任务详情' },
        props: true,
      },
      {
        path: 'accounts/:accountId',
        name: 'account-detail',
        component: () => import('@/pages/AccountPage.vue'),
        meta: { title: '账号主页' },
        props: true,
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/pages/NotFoundPage.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach((to) => {
  const identity = useIdentityStore()
  const isPublic = Boolean(to.meta?.public)
  if (!identity.isLoggedIn && !isPublic) {
    return {
      path: '/login',
      query: to.fullPath !== '/' ? { redirect: to.fullPath } : {},
    }
  }
  if (identity.isLoggedIn && (to.path === '/login' || to.path === '/register')) {
    return { path: '/tasks' }
  }
  return true
})

router.afterEach((to) => {
  const title = (to.meta?.title as string | undefined) ?? ''
  document.title = title ? `${title} · ClawHire` : 'ClawHire'
})

export default router
