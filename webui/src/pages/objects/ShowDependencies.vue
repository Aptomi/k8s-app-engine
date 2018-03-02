<template>
  <div>

    <div class="row" v-if="loading">
      <div class="col-xs-12">
        <div class="box">
          <div class="overlay">
            <i class="fa fa-refresh fa-spin"></i>
          </div>
        </div>
      </div>
    </div>

    <div class="row" v-if="error">
      <div class="col-xs-12">
        <div class="box">
          <table class="table table-hover">
            <tbody>
            <tr>
              <td><span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i></td>
            </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="row" v-if="!loading && !error && (dataMapByNs == null || Object.keys(dataMapByNs).length <= 0)">
      <div class="col-xs-12">
        <div class="box">
          <table class="table table-hover">
            <tbody>
              <tr>
                <td>No Dependencies Defined</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- /.row -->
    <div v-for="(objList, ns) in dataMapByNs" class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Dependencies: <b>{{ ns }}</b></h3>
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
              <tr v-for="d in objList">
                <td>
                  <img style="float: left; height: 20px; margin-right: 5px" src="/static/img/dependency-icon.png" title="Dependency"/>
                  <span>{{d.name}}</span>
                </td>
                <td v-if="!d.error">{{d.user}}</td>
                <td v-else><span class="label label-danger center">Error</span></td>
                <td v-if="!d.error">{{d.contract}}</td>
                <td v-else><span class="label label-danger center">Error</span></td>
                <td v-if="!d.status_error">
                  <span class="label" v-bind:class="{ 'label-success': d['status'] === 'Active', 'label-warning': d['status'] !== 'Active'}">{{d.status}}</span>
                </td>
                <td v-else><span class="label label-danger center">Error</span></td>
                <td>
                  <button type="button" class="btn btn-default btn-xs" @click="showEndpoints(d)">Endpoints</button>
                  <button type="button" class="btn btn-default btn-xs" @click="showResources(d)">Resources</button>
                  <button type="button" class="btn btn-default btn-xs" @click="showDiagram(d)">Diagram</button>
                  <button type="button" class="btn btn-default btn-xs" @click="editYaml(d)">Edit</button>
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
  import {getPolicyObjectsWithProperties, getObjectMapByNamespace} from 'lib/api.js'
  import objectEditYAML from 'pages/components/ObjectEditYAML'
  import objectDiagram from 'pages/components/ObjectDiagram'
  import endpoints from 'pages/components/Endpoints'
  import resources from 'pages/components/Resources'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        dataMapByNs: null,
        error: null
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchData()
    },
    methods: {
      showEndpoints (obj) {
        this.$modal.show(endpoints, {
          dependency: obj
        }, {
          width: '60%',
          height: 'auto'
        })
      },
      showDiagram (obj) {
        this.$modal.show(objectDiagram, {
          obj: obj,
          height: '465px'
        }, {
          width: '60%',
          height: '550px'
        })
      },
      showResources (obj) {
        this.$modal.show(resources, {
          dependency: obj
        }, {
          width: '60%',
          height: 'auto'
        })
      },
      editYaml (obj) {
        this.$modal.show(objectEditYAML, {
          obj: obj,
          height: '465px'
        }, {
          width: '60%',
          height: '550px'
        })
      },
      fetchData () {
        this.loading = true
        this.dataMapByNs = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          this.dataMapByNs = getObjectMapByNamespace(data)
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getPolicyObjectsWithProperties(fetchSuccess, fetchError, 'dependency')
      }
    }
  }
</script>

<style>

</style>
