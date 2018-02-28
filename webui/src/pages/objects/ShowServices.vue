<template>
  <div>

    <!-- /.row -->
    <div class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Services</h3>
          </div>
          <!-- /.box-header -->
          <div class="overlay" v-if="loading">
            <i class="fa fa-refresh fa-spin"></i>
          </div>
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <thead>
              <tr>
                <th>Namespace</th>
                <th>Name</th>
                <th>Components</th>
                <th>Action</th>
              </tr>
              </thead>
              <tbody>
              <tr v-if="error">
                <td><span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i></td>
              </tr>
              <tr v-if="services == null || services.length <= 0">
                <td>No Services Defined</td>
              </tr>
              <tr v-for="d in services">
                <td>
                  <span class="label label-primary">{{d.namespace}}</span>
                </td>
                <td>{{d.name}}</td>
                <td>
                  <span v-for="c in d.components" class="label label-success" style="margin-right: 5px">{{c.name}}</span>
                </td>
                <td>
                  <button type="button" class="btn btn-default btn-xs" @click="">Diagram</button>
                  <button type="button" class="btn btn-default btn-xs" @click="">YAML</button>
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

  </div>
</template>

<script>
  import {getPolicyObjectsWithProperties} from 'lib/api.js'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        services: null,
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
        this.services = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          this.services = data
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getPolicyObjectsWithProperties(fetchSuccess, fetchError, 'service')
      }
    }
  }
</script>

<style>

</style>
