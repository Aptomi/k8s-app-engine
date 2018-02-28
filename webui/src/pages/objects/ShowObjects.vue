<template>
  <div>

    <div class="box box-default">

      <div class="overlay" v-if="loading">
        <i class="fa fa-refresh fa-spin"></i>
      </div>

      <div class="box-body">
        <div class="row">
          <div class="col-xs-4">
            <div class="form-group">
              <label>Namespace</label>
              <!-- :multiple="false" :allowempty="false"  -->
              <v-select placeholder="Select namespace" v-model="selectedNamespace" :options.sync="namespaces" :allow-empty="false" deselect-label="Selected"></v-select>
            </div>
            <!-- /.form-group -->
          </div>
          <!-- /.col -->
          <div class="col-xs-8">
            <div class="form-group">
              <label>Object</label>
              <v-select placeholder="Select object" v-model="selectedObject" :options.sync="objectList" label="label" track-by="label" group-label="kind" group-values="list" :allow-empty="false" deselect-label="Selected"></v-select>
            </div>
            <!-- /.form-group -->
          </div>
          <!-- /.col -->
        </div>
        <!-- /.row -->

        <div class="row" v-if="error">
          <div class="col-xs-12">
            <span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i>
          </div>
        </div>
        <!-- /.row -->

      </div>
    </div>

    <object-yaml v-if="selectedObject" :obj="selectedObject" height="100%"></object-yaml>

  </div>
</template>

<script>
  import vSelect from 'vue-multiselect'
  import objectYaml from 'pages/components/ObjectYAML'
  import { getPolicy, getPolicyObjects, getNamespaces, filterObjects } from 'lib/api.js'

  export default {
    data () {
      return {
        loading: false,
        policyObjects: null,
        error: null,
        namespaces: [],
        selectedNamespace: null,
        objectList: [],
        selectedObject: null
      }
    },
    watch: {
      policyObjects: function (data) {
        // once policy objects are loaded, create the list of namespaces for the first dropdown
        this.namespaces = getNamespaces(data)
        if (this.namespaces.length > 0) {
          // select first namespace
          this.selectedNamespace = this.namespaces[0]
        }
      },
      selectedNamespace: function (ns) {
        // once namespace is selected, create the list of objects for the second dropdown
        this.selectedObject = null
        this.filteredObjects = filterObjects(this.policyObjects, ns)

        // group objects by kind
        let byKind = {}
        for (const idx in this.filteredObjects) {
          let obj = this.filteredObjects[idx]
          obj['label'] = [obj['kind'], obj['name']].join('/')
          if (!(obj['kind'] in byKind)) {
            byKind[obj['kind']] = []
          }
          byKind[obj['kind']].push(obj)
        }

        // add them to the object list
        this.objectList = []
        for (const kind in byKind) {
          this.objectList.push({
            'kind': kind,
            'list': byKind[kind]
          })
        }

        // select first object
        if (this.filteredObjects.length > 0) {
          // select first namespace
          this.selectedObject = this.filteredObjects[0]
        }
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchObjectList()
    },
    methods: {
      fetchObjectList () {
        this.loading = true
        this.policyObjects = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          this.policyObjects = getPolicyObjects(data)
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getPolicy(fetchSuccess, fetchError)
      }
    },
    components: {
      vSelect,
      objectYaml
    }
  }
</script>

<style>

</style>
