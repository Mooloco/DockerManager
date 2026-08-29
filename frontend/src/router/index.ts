import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import MainLayout from '../layouts/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true, title: '登录' },
    },
    {
      path: '/',
      component: MainLayout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('../views/DashboardView.vue'),
          meta: { title: '总览' },
        },
        {
          path: 'containers',
          name: 'containers',
          component: () => import('../views/ContainersView.vue'),
          meta: { title: '容器' },
        },
        {
          path: 'containers/:id',
          name: 'container-detail',
          component: () => import('../views/ContainerDetailView.vue'),
          meta: { title: '容器详情' },
        },
        {
          path: 'projects',
          name: 'projects',
          component: () => import('../views/ProjectsView.vue'),
          meta: { title: '项目 · Docker Compose' },
        },
        {
          path: 'projects/:name',
          name: 'project-detail',
          component: () => import('../views/ProjectDetailView.vue'),
          meta: { title: '项目详情' },
        },
        {
          path: 'images',
          name: 'images',
          component: () => import('../views/ImagesView.vue'),
          meta: { title: '镜像' },
        },
        {
          path: 'networks',
          name: 'networks',
          component: () => import('../views/NetworksView.vue'),
          meta: { title: '网络' },
        },
        {
          path: 'networks/:id',
          name: 'network-detail',
          component: () => import('../views/NetworkDetailView.vue'),
          meta: { title: '网络详情' },
        },
        {
          path: 'volumes',
          name: 'volumes',
          component: () => import('../views/VolumesView.vue'),
          meta: { title: '卷' },
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('../views/SettingsView.vue'),
          meta: { title: '设置' },
        },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
  ],
})

// 全局守卫:未登录跳转登录页
router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    // 已登录访问登录页 → 回首页
    if (auth.isLoggedIn && to.name === 'login') return { name: 'dashboard' }
    return true
  }
  if (!auth.isLoggedIn) {
    // 尝试恢复会话(刷新后)
    const ok = await auth.restore()
    if (!ok) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
  }
  return true
})

router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} · Docker Manager` : 'Docker Manager'
})

export default router
