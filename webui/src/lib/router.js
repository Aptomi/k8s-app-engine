import Vue from 'vue'
import Router from 'vue-router'

// Library for changing body classes
import vbclass from 'vue-body-class'

// Authentication library
import auth from 'lib/auth'

// Aptomi pages
import Login from 'pages/auth/Login.vue'
import ShowServices from 'pages/objects/ShowServices.vue'
import ShowContracts from 'pages/objects/ShowContracts.vue'
import ShowClusters from 'pages/objects/ShowClusters.vue'
import ShowRules from 'pages/objects/ShowRules.vue'
import ShowDependencies from 'pages/objects/ShowDependencies.vue'
import ShowUserRoles from 'pages/objects/ShowUserRoles.vue'
import ShowAuditLog from 'pages/deployment/ShowAuditLog.vue'
import BrowsePolicy from 'pages/debug/BrowsePolicy.vue'

Vue.use(Router)

const Passthrough = {
  template: '<router-view></router-view>'
}

const router = new Router({
  mode: 'hash',
  routes: [
    {
      path: '/',
      name: 'Home',
      redirect: '/objects/services'
    },
    {
      path: '/objects',
      name: 'Objects',
      component: Passthrough,
      children: [
        {
          path: 'services',
          name: 'ShowServices',
          component: ShowServices
        },
        {
          path: 'contracts',
          name: 'ShowContracts',
          component: ShowContracts
        },
        {
          path: 'dependencies',
          name: 'ShowDependencies',
          component: ShowDependencies
        },
        {
          path: 'rules',
          name: 'ShowRules',
          component: ShowRules
        },
        {
          path: 'clusters',
          name: 'ShowClusters',
          component: ShowClusters
        },
        {
          path: 'users',
          name: 'ShowUserRoles',
          component: ShowUserRoles
        }
      ]
    },
    {
      path: '/deployment',
      name: 'Deployment',
      component: Passthrough,
      children: [
        {
          path: 'audit',
          name: 'ShowAuditLog',
          component: ShowAuditLog
        }
      ]
    },
    {
      path: '/debug',
      name: 'Debug',
      component: Passthrough,
      children: [
        {
          path: 'browse',
          name: 'BrowsePolicy',
          component: BrowsePolicy,
          props: true
        }
      ]
    },
    {
      path: '/help',
      name: 'Help',
      component: Passthrough,
      children: [
        {
          path: 'website',
          name: 'Website',
          beforeEnter (to, from, next) {
            window.location.href = 'http://aptomi.io'
          }
        },
        {
          path: 'documentation',
          name: 'Documentation',
          beforeEnter (to, from, next) {
            window.location.href = 'https://godoc.org/github.com/Aptomi/aptomi'
          }
        },
        {
          path: 'slack',
          name: 'Slack',
          beforeEnter (to, from, next) {
            window.location.href = 'http://slack.aptomi.io'
          }
        },
        {
          path: 'github',
          name: 'Github',
          beforeEnter (to, from, next) {
            window.location.href = 'https://github.com/Aptomi/aptomi'
          }
        }
      ]
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
        window.location.href = '/'
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
