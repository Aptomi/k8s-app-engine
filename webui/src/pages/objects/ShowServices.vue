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
                  <obj-with-icon :obj="d"/>
                </td>
                <td>
                  <div v-for="c in d.components" v-if="c.contract != null">
                    <obj-with-icon kind="contract" :obj="c"/>
                  </div>
                </td>
                <td>
                  <div v-for="c in d.components" v-if="c.code != null" style="margin-right: 5px">
                    <obj-with-icon kind="code" :obj="c"/>
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
