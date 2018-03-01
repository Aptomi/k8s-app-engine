<template>
  <div id="app">
    <div class="wrapper" v-if="loggedIn">
      <modals-container/>
      <notifications group="main" position="bottom right" />

      <va-navibar></va-navibar>
      <va-menu :menuItems="menuItems"></va-menu>
      <va-content></va-content>

      <!-- Main Footer -->
      <footer class="main-footer">
        <!-- Default to the left -->
        <strong>Copyright &copy; 2018 <a href="http://aptomi.io">Aptomi</a>.</strong> All rights reserved.
      </footer>
    </div>
    <div v-else>
      <section class="content container-fluid">
        <transition name="page" mode="out-in">
          <router-view></router-view>
        </transition>
      </section>
    </div>
  </div>
</template>

<script>
import VANaviBar from './pages/components/NaviBar'
import VAMenu from './pages/components/Menu'
import VAAppContent from 'AppContent.vue'
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
    'va-content': VAAppContent
  },
  store
}
</script>

<style>

  .multiselect__option--selected.multiselect__option--highlight {
    background: #41B883;
  }
  .multiselect__option--selected.multiselect__option--highlight:after {
    background: #41B883;
  }

</style>
