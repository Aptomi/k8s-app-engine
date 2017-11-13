<template>
  <div>
    <div class="box">
      <div class="box-header">
        <h3 class="box-title">Endpoints: <b>{{ dependency.namespace }} / {{ dependency.kind }} / {{ dependency.name }}</b></h3>
      </div>
      <!-- /.box-header -->
      <div class="overlay" v-if="loading">
        <i class="fa fa-refresh fa-spin"></i>
      </div>
      <div class="box-body table-responsive no-padding">
        <table class="table table-hover">
          <thead>
          <tr>
            <th>Component</th>
            <th>URL</th>
          </tr>
          </thead>
          <tbody>
          <tr v-if="error">
            <td><span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i></td>
          </tr>
          <tr v-if="!endpoints">
            <td>No Endpoints Available</td>
          </tr>
          <tr v-for="e in endpoints">
            <td>{{e.component}}</td>
            <td><a :href="e.url">{{e.url}}</a></td>
          </tr>
          </tbody>
        </table>
      </div>
      <!-- /.box-body -->
    </div>
    <!-- /.box -->

  </div>
</template>
<script>
  import { getEndpoints } from 'lib/api.js'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        endpoints: null,
        error: null
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchData()
    },
    props: {
      'dependency': {
        type: Object,
        validator: function (value) {
          return true
        }
      }
    },
    watch: {
      'dependency': 'fetchData'
    },
    methods: {
      fetchData () {
        this.loading = true
        this.endpoints = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          this.endpoints = data
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getEndpoints(fetchSuccess, fetchError)
      }
    }
  }
</script>
<style>
</style>
