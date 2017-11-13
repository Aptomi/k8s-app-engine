<template>
  <div id="app">
    <div class="wrapper" v-if="loggedIn">
      <va-navibar></va-navibar>
      <va-menu :menuItems="menuItems"></va-menu>
      <va-content-wrap></va-content-wrap>

      <!-- Main Footer -->
      <footer class="main-footer">
        <!-- Default to the left -->
        <strong>Copyright &copy; 2017 <a href="http://aptomi.io">Aptomi</a>.</strong> All rights reserved.
      </footer>

      <Modal></Modal>
    </div>
    <div v-else>
      <section class="content container-fluid">
        <transition name="page" mode="out-in">
          <router-view></router-view>
        </transition>
      </section>

      <Modal></Modal>
    </div>
  </div>
</template>

<script>
import VANaviBar from 'NaviBar.vue'
import VAMenu from 'Menu.vue'
import VAContentWrap from 'ContentWrap.vue'
import Modal from './components/Modal.vue'
import store from './vuex/store.js'
import menuItems from './lib/menuItems.js'
import auth from 'lib/auth'

export default {
  name: 'app',
  data () {
    return {
      menuItems: menuItems,
      loggedIn: auth.loggedIn()
    }
  },
  components: {
    'va-navibar': VANaviBar,
    'va-menu': VAMenu,
    'va-content-wrap': VAContentWrap,
    Modal
  },
  store
}
</script>

<style>

</style>
