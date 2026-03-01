import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../store/auth'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'

const routes = [
  { 
    path: '/login', 
    name: 'Login', 
    component: Login,
    meta: { public: true }
  },
  { 
    path: '/', 
    name: 'Dashboard', 
    component: Dashboard 
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _, next) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isAuthenticated) {
    next({ name: 'Login' })
  } else if (to.name === 'Login' && auth.isAuthenticated) {
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
