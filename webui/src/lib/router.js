import Vue from 'vue'
import Router from 'vue-router'

// Library for changing body classes
import vbclass from 'vue-body-class'

// Authentication library
import auth from 'lib/auth'

// Aptomi pages
import Login from 'pages/auth/Login.vue'
import ShowObjects from 'pages/policy/ShowObjects.vue'
import BrowsePolicy from 'pages/policy/BrowsePolicy.vue'
import ShowDependencies from 'pages/policy/ShowDependencies.vue'
import ShowUserRoles from 'pages/policy/ShowUserRoles.vue'
import ShowAuditLog from 'pages/policy/ShowAuditLog.vue'

Vue.use(Router)

const router = new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'Home',
      redirect: '/policy/objects'
    },
    {
      path: '/policy/objects',
      name: 'ShowObjects',
      component: ShowObjects
    },
    {
      path: '/policy/browse',
      name: 'BrowsePolicy',
      component: BrowsePolicy
    },
    {
      path: '/policy/dependencies',
      name: 'ShowDependencies',
      component: ShowDependencies
    },
    {
      path: '/policy/users',
      name: 'ShowUserRoles',
      component: ShowUserRoles
    },
    {
      path: '/policy/audit',
      name: 'ShowAuditLog',
      component: ShowAuditLog
    },
    {
      path: '/login',
      component: Login,
      meta: { bodyClass: 'hold-transition login-page' }
    },
    {
      path: '/logout',
      beforeEnter (to, from, next) {
        auth.logout()
        window.location.href = '/login'
      }
    }
  ],
  linkActiveClass: 'active'
})

// change body classes for certain pages (such as /login)
Vue.use(vbclass, router)

// enforce authentication
router.beforeEach((to, from, next) => {
  if (to.path === '/login' || to.path === '/logout') {
    // we are on login or logout pages
    next()
  } else {
    if (!auth.loggedIn()) {
      next('/login')
    } else {
      next()
    }
  }
})

export default router
