<template>
  <div id="app">
    <div class="login-box" v-if="!loggedIn">
      <div class="login-logo">
        <img src="static/img/aptomi-logo.png"/>
      </div>
      <!-- /.login-logo -->
      <div class="login-box-body">
        <p class="login-box-msg">Sign in</p>

        <form @submit.prevent="login">
          <div class="form-group has-feedback">
            <input class="form-control" v-model="username" placeholder="Username">
            <span class="glyphicon glyphicon-user form-control-feedback"></span>
          </div>
          <div class="form-group has-feedback">
            <input type="password" class="form-control" v-model="password" placeholder="Password">
            <span class="glyphicon glyphicon-lock form-control-feedback"></span>
          </div>
          <div class="row">
            <div class="col-xs-8">
              <div class="checkbox">
                <label>
                  <input type="checkbox"> Remember Me
                </label>
              </div>
            </div>
            <!-- /.col -->
            <div class="col-xs-4">
              <button type="submit" class="btn btn-primary btn-block btn-flat">Sign In</button>
            </div>
            <!-- /.col -->
          </div>
          <div class="row" v-if="error">
            <div class="col-xs-12">
              <p class="error">Unable to log in (check username/password)</p>
            </div>
          </div>
        </form>

      </div>
      <!-- /.login-box-body -->
    </div>
  </div>
</template>

<script>
  import auth from 'lib/auth'

  export default {
    data () {
      return {
        username: '',
        password: '',
        error: false,
        loggedIn: auth.loggedIn()
      }
    },
    methods: {
      login () {
        this.error = false

        auth.login(this.username, this.password, loggedIn => {
          if (!loggedIn) {
            this.error = true
          } else {
            window.location.href = '/'
          }
        })
      }
    }
  }
</script>

<style>
  .error {
    color: red;
  }
</style>
