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
                <td>No Services Defined</td>
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
            <h3 class="box-title">Services: <b>{{ ns }}</b></h3>
          </div>
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <thead>
              <tr>
                <th>Name</th>
                <th>Uses</th>
                <th>Code</th>
                <th>Action</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="d in objList">
                <td>
                  <img style="float: left; height: 20px; margin-right: 5px" src="/static/img/service-icon.png" title="Service"/>
                  <span>{{d.name}}</span>
                </td>
                <td>
                  <div v-for="c in d.components" v-if="c.contract != null">
                    <img style="float: left; height: 20px; margin-right: 5px" src="/static/img/contract-icon.png" title="Contract"/>
                    {{c.name}}
                  </div>
                </td>
                <td>
                  <div v-for="c in d.components" v-if="c.code != null" style="margin-right: 5px">
                    <div v-if="c.code.type.indexOf('helm') >= 0">
                      <img style="float: left; height: 20px; margin-right: 5px" src="/static/img/helm-icon.png" title="Helm Chart"/>
                      <span>{{c.code.params.chartName}} /</span>
                      <span v-if="c.code.params.chartVersion != null">{{c.code.params.chartVersion}}</span>
                      <span v-else>latest</span>
                    </div>
                    <div v-else-if="c.code.type.indexOf('raw') >= 0">
                      <img style="float: left; height: 20px; margin-right: 5px" src="/static/img/k8s-icon.png" title="Kubernetes YAMLs"/>
                      <span>{{c.name}}</span>
                    </div>
                    <div v-else>
                      <span class="label label-danger">Unknown code type</span>
                    </div>
                  </div>
                </td>
                <td>
                  <!-- <button type="button" class="btn btn-default btn-xs" @click="showDiagram(d)">Diagram</button> -->
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
      showDiagram (obj) {
        this.$modal.show(objectDiagram, {
          obj: obj,
          height: '465px'
        }, {
          width: '60%',
          height: '550px'
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

        getPolicyObjectsWithProperties(fetchSuccess, fetchError, 'service')
      }
    }
  }
</script>

<style>

</style>
