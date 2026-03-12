import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('../layouts/DefaultLayout.vue'),
      children: [
        { path: '', redirect: '/providers' },
        { path: 'providers', name: 'providers', component: () => import('../views/ProvidersView.vue') },
        { path: 'keys', name: 'keys', component: () => import('../views/KeysView.vue') },
        { path: 'usage', name: 'usage', component: () => import('../views/UsageView.vue') },
      ],
    },
    { path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('admin_token')
  if (!token && to.name !== 'login') return { name: 'login' }
})

export default router
