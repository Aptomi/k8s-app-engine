<template>
  <div>
    <div class="box">
      <div class="box-header">
        <h3 class="box-title">Logs for revision: <b>{{ revision.metadata.generation }}</b></h3>
      </div>
      <!-- /.box-header -->
      <div class="overlay" v-if="loading">
        <i class="fa fa-refresh fa-spin"></i>
      </div>
      <div class="box-body table-responsive no-padding">
        <table class="table table-hover">
          <thead>
          <tr>
            <th>Time</th>
            <th>Level</th>
            <th>Message</th>
          </tr>
          </thead>
          <tbody>
          <tr v-if="error">
            <td><span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i></td>
          </tr>
          <tr v-if="log == null || log.length <= 0">
            <td>No Logs Available</td>
          </tr>
          <tr v-for="entry in log">
            <td>{{ entry.time | formatDateAgo }}</td>
            <td>{{ entry.level }}</td>
            <td>{{ entry.message }}</td>
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
  import { getEventLogs } from 'lib/api.js'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        log: null,
        error: null
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchData()
    },
    props: {
      'revision': {
        type: Object,
        validator: function (value) {
          return true
        }
      },
      'type': {
        type: String,
        validator: function (value) {
          return true
        }
      }
    },
    watch: {
      'revision': 'fetchData',
      'type': 'fetchData'
    },
    methods: {
      fetchData () {
        this.loading = true
        this.log = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          if (this.type === 'resolve') {
            this.log = data.resolvelog
          } else if (this.type === 'apply') {
            this.log = data.applylog
          }
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getEventLogs(this.revision, fetchSuccess, fetchError)
      }
    }
  }
</script>
<style>
</style>
