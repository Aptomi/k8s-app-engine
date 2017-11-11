<template>
  <div>

    <!-- /.row -->
    <div class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Declared Dependencies</h3>
          </div>
          <!-- /.box-header -->
          <div class="overlay" v-if="loading">
            <i class="fa fa-refresh fa-spin"></i>
          </div>
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>User</th>
                  <th>Contract</th>
                  <th>Status</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="error">
                  <td><span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i></td>
                </tr>
                <tr>
                  <td>alice-stage</td>
                  <td>Alice</td>
                  <td>twitter-stats</td>
                  <td><span class="label label-success">Deployed</span></td>
                  <td>
                    <button type="button" class="btn btn-default btn-xs">Show Endpoints</button>
                  </td>
                </tr>
                <tr>
                  <td>bob-stage</td>
                  <td>Bob</td>
                  <td>twitter-stats</td>
                  <td><span class="label label-primary center">Processing</span></td>
                  <td>
                  </td>
                </tr>
                <tr>
                  <td>carol-stage</td>
                  <td>Carol</td>
                  <td>twitter-stats</td>
                  <td><span class="label label-danger center">Not Resolved</span></td>
                  <td>
                  </td>
                </tr>
                <tr>
                  <td>prod</td>
                  <td>John</td>
                  <td>twitter-stats</td>
                  <td><span class="label label-success">Deployed</span></td>
                  <td>
                    <button type="button" class="btn btn-default btn-xs">Show Endpoints</button>
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
    <div class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Endpoints for <b>alice-stage</b></h3>
          </div>
          <!-- /.box-header -->
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <thead>
                <tr>
                  <th>Component</th>
                  <th>URL</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>zookeeper</td>
                  <td><a href="#">http://127.0.0.1:12345</a></td>
                </tr>
                <tr>
                  <td>hdfs</td>
                  <td><a href="#">http://127.0.0.1:23456</a></td>
                </tr>
              </tbody>
            </table>
          </div>
          <!-- /.box-body -->
        </div>
        <!-- /.box -->
      </div>
    </div>

  </div>
</template>

<script>
import { callAPI } from 'lib/api.js'

export default {
  name: 'show-dependencies',
  data () {
    // empty data
    return {
      loading: false,
      dependencies: null,
      error: 'test'
    }
  },
  created () {
    // fetch the data when the view is created and the data is already being observed
    this.fetchData()
  },
  watch: {
    // call again the method if the route changes
    '$route': 'fetchData'
  },
  methods: {
    fetchData () {
      this.loading = true
      this.dependencies = null
      this.error = null

      var fetchSuccess = $.proxy(function (data) {
        this.loading = false
        this.dependencies = data
        console.log('Data: ' + data)
      }, this)

      var fetchError = $.proxy(function (err) {
        this.loading = false
        this.error = err
        // console.log('Error: ' + err)
      }, this)

      callAPI('policy', fetchSuccess, fetchError)
    }
  }
}
</script>

<style>

</style>
