<template>
  <header class="main-header">
    <!-- Logo -->
    <router-link to="/" class="logo">
      <!-- mini logo for sidebar mini 50x50 pixels -->
      <span class="logo-mini"><img src="/static/img/aptomi-logo-square.png" width="50%"/></span>
      <!-- logo for regular state and mobile devices -->
      <span class="logo-lg"><img src="/static/img/aptomi-logo.png" width="50%"/></span>
    </router-link>

    <!-- Header Navbar: style can be found in header.less -->
    <nav class="navbar navbar-static-top">
      <!-- Sidebar toggle button-->
      <a href="#" class="sidebar-toggle" data-toggle="push-menu" role="button">
        <span class="sr-only">Toggle navigation</span>
      </a>
      <!-- Navbar Right Menu -->
      <div class="navbar-custom-menu">
        <ul class="nav navbar-nav">
          <!-- User Account: style can be found in dropdown.less -->
          <li class="dropdown user user-menu">
            <a href="#" class="dropdown-toggle" data-toggle="dropdown">
              <img src="/static/img/user-navbar.jpg" class="user-image" alt="User Image">
              <span class="hidden-xs">{{ username }}</span>
            </a>
            <ul class="dropdown-menu">
              <!-- User image -->
              <li class="user-header">
                <img src="/static/img/user-navbar.jpg" class="img-circle" alt="User Image">

                <p>
                  {{ username }}
                </p>
              </li>
              <!-- Menu Footer-->
              <li class="user-footer">
                <div class="pull-right">
                  <router-link to="/logout" class="btn btn-default btn-flat">Sign out</router-link>
                </div>
              </li>
            </ul>
          </li>
        </ul>
      </div>
    </nav>
  </header>
</template>

<script>
import auth from 'lib/auth'

export default {
  data () {
    return {
      username: auth.getUsername()
    }
  }
}
</script>
