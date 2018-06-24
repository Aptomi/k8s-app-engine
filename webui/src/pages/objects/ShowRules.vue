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
                <td>No Rules Defined</td>
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
            <h3 class="box-title">Rules: <b>{{ ns }}</b></h3>
          </div>
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <thead>
              <tr>
                <th>Name</th>
                <th>Weight</th>
                <th>Criteria</th>
                <th>Rule Actions</th>
                <th>Action</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="d in sorted(objList)">
                <td>
                  <obj-with-icon :obj="d"/>
                </td>
                <td>
                  {{d.weight}}
                </td>
                <td>
                  <div v-if="d.criteria == null || d.criteria.length <= 0">
                    <span class="label label-success">require-all</span>
                    <span>true</span>
                  </div>
                  <div v-else>
                    <div v-for="(cList, cType) in d.criteria" >
                      <div v-for="c in cList">
                        <span class="label" v-bind:class="{ 'label-success': cType === 'require-all', 'label-info': cType === 'require-any', 'label-danger': cType === 'require-none'}">{{cType}}</span>
                        <span>{{ c }}</span>
                      </div>
                    </div>
                  </div>
                </td>
                <td>
                  <div v-for="(aData, aType) in d.actions" >
                    <div v-if="aType === 'change-labels'">
                      <div v-for="(v,k) in aData.set">
                        <span class="label label-success">[+] label</span>
                        <span>{{ k }} = {{ v }}</span>
                      </div>
                      <div v-for="(v,k) in aData.remove">
                        <span class="label label-danger">[-] label</span>
                        <span>{{ k }}</span>
                      </div>
                    </div>
                    <div v-else-if="aType === 'claim'">
                      <div v-if="aData === 'reject'">
                        <span class="label label-danger">{{ aType }}</span>
                        <span>{{ aData }}</span>
                      </div>
                      <div v-else>
                        <span class="label label-danger">Unknown claim action</span>
                      </div>
                    </div>
                    <div v-else>
                      <span class="label label-danger">Unknown rule type</span>
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
      editYaml (obj) {
        this.$modal.show(objectEditYAML, {
          obj: obj,
          height: '465px'
        }, {
          width: '60%',
          height: '550px'
        })
      },
      sorted: function (objList) {
        return objList.slice().sort(function (a, b) { return a.weight - b.weight })
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

        getPolicyObjectsWithProperties(fetchSuccess, fetchError, 'rule')
      }
    }
  }
</script>

<style>

</style>
