<template>
  <div>

    <div class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Users</h3>
          </div>
          <!-- /.box-header -->
          <div class="overlay" v-if="loading">
            <i class="fa fa-refresh fa-spin"></i>
          </div>
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <thead>
                <tr>
                  <th>User</th>
                  <th>Domain Admin</th>
                  <th>Namespace Admin</th>
                  <th>Service Consumer</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="error">
                  <td><span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i></td>
                </tr>
                <tr v-for="roleMap, username in users">
                  <td>{{ username }}</td>
                  <td>
                    <span v-if="roleMap['domain-admin']" class="label label-success"><label class="fa fa-check"></label></span>
                  </td>
                  <td>
                    <span v-for="flag, namespace in roleMap['namespace-admin']" class="label label-primary">{{ namespace }}</span>
                  </td>
                  <td>
                    <span v-for="flag, namespace in roleMap['service-consumer']" class="label label-primary">{{ namespace }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <!-- /.box-body -->
        </div>
        <!-- /.box -->
      </div>
    </div>
    <!-- /.row -->

  </div>
</template>

<script>
  import {getUsersAndRoles} from 'lib/api.js'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        users: null,
        error: null
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchData()
    },
    methods: {
      fetchData () {
        this.loading = true
        this.users = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          this.users = data
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getUsersAndRoles(fetchSuccess, fetchError)
      }
    }
  }
</script>

<style>

</style>
