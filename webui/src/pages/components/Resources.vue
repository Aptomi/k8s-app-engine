<template>
  <div>
    <div class="box">
      <div class="box-header">
        <h3 class="box-title">Resources: <b>{{ dependency.namespace }} / {{ dependency.kind }} / {{ dependency.name }}</b></h3>
      </div>
      <!-- /.box-header -->
      <div class="overlay" v-if="loading">
        <i class="fa fa-refresh fa-spin"></i>
      </div>
      <div v-if="error">
        <span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i>
      </div>
      <div v-if="resources == null || Object.keys(resources).length <= 0">
        No Resources Found
      </div>

      <div class="box-body table-responsive no-padding" v-for="table, type in resources">
        <div class="box-header">
          <h3 class="box-title">{{ type }}</h3>
        </div>
        <table class="table table-hover">
          <thead>
          <tr>
            <th v-for="header in table.headers">{{ header }}</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="item in table.items">
            <td v-for="column in item">{{ column }}</td>
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
  import { getResources } from 'lib/api.js'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        resources: null,
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
        this.resources = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          this.resources = data
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getResources(this.dependency, fetchSuccess, fetchError)
      }
    }
  }
</script>
<style>
</style>
