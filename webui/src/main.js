// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import './lib/css'
import './lib/script'

import Vue from 'vue'
import App from './App'
import router from './lib/router'
import EventBus from './lib/eventBus.js'
import axios from 'axios'
import moment from 'moment'
import VModal from 'vue-js-modal'

Vue.prototype.$bus = EventBus
Vue.prototype.$http = axios

Vue.filter('formatDateAgo', function (value) {
  if (value) {
    return moment(String(value)).fromNow()
  }
})

Vue.filter('formatDate', function (value) {
  if (value) {
    return moment(String(value)).format('MM/DD/YYYY hh:mm:ss')
  }
})

Vue.use(VModal, { dynamic: true })

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  template: '<App/>',
  components: {
    App
  }
})
