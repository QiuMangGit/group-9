import { createRouter, createWebHistory } from 'vue-router'
import mainVue from '../view/main.vue'
import LoginVue from '../view/Login.vue'
import RegisterVue from '../view/Register.vue'

const routes = [
  { path: '/', redirect: '/login' },
  { path: '/main', component: mainVue, meta: { requiresAuth: true }},
  { path: '/login', component: LoginVue },
  { path: '/register', component: RegisterVue }
]

const router = createRouter({
  history: createWebHistory(),
  routes: routes
})

router.beforeEach((to, from, next) => {
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth);
  const user = sessionStorage.getItem('user');

  if (requiresAuth && !user) {
    next('/login');
  } else {
    next();
  }
})

export default router;