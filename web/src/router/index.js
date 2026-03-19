import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/documentation',
    name: 'Documentation',
    component: () => import('@/views/Documentation.vue'),
    meta: { title: 'API 文档', requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: 'collections',
        name: 'Collections',
        component: () => import('@/views/Collections.vue'),
        meta: { title: '集合管理' }
      },
      {
        path: 'collections/:name/detail',
        name: 'CollectionDetail',
        component: () => import('@/views/CollectionDetail.vue'),
        meta: { title: '集合详情' }
      },
      {
        path: 'collections/:name',
        name: 'RecordList',
        component: () => import('@/views/RecordList.vue'),
        meta: { title: '记录列表' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '系统设置' }
      },
      {
        path: 'dictionaries',
        name: 'Dictionaries',
        component: () => import('@/views/Dictionaries.vue'),
        meta: { title: '字典管理' }
      },
      {
        path: 'admins',
        name: 'Admins',
        component: () => import('@/views/Admins.vue'),
        meta: { title: '管理员' }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('@/views/Logs.vue'),
        meta: { title: '操作日志' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    next({ name: 'Login' })
  } else if (to.name === 'Login' && authStore.isLoggedIn) {
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
