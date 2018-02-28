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
                <th>Kind</th>
                <th>Name</th>
                <th>Uses</th>
                <th>Code</th>
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
                <td>{{d.namespace}}</td>
                <td>{{d.kind}}</td>
                <td>
                  <img style="float: left; height: 20px; margin-right: 5px" src="/static/img/service-icon.png" alt="Service"/>
                  <span>{{d.name}}</span>
                </td>
                <td>
                  <div v-for="c in d.components" v-if="c.contract != null">
                    <img style="float: left; height: 20px; margin-right: 5px" src="/static/img/contract-icon.png" alt="Contract"/>
                    {{c.name}}
                  </div>
                </td>
                <td>
                  <div v-for="c in d.components" v-if="c.code != null && c.code.type.indexOf('helm') >= 0" style="margin-right: 5px">
                    <img style="float: left; height: 20px; margin-right: 5px" src="/static/img/helm-logo.png" alt="Helm"/>
                    <span><b>{{c.code.params.chartName}}</b></span>
                    <span v-if="c.code.params.chartVersion != null">{{c.code.params.chartVersion}}</span>
                    <span v-else>latest</span>
                  </div>
                </td>
                <td>
                  <button type="button" class="btn btn-default btn-xs" @click="">Diagram</button>
                  <button type="button" class="btn btn-default btn-xs" @click="openYaml(d)">YAML</button>
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
  import objectYAML from 'pages/components/ObjectYAML'

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
      openYaml (obj) {
        this.$modal.show(objectYAML, {
          obj: obj,
          height: '465px'
        }, {
          width: '60%',
          height: '550px'
        })
      },
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
