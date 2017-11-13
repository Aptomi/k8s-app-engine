<template>
  <div>

    <div class="box box-default">

      <div class="overlay" v-if="loading">
        <i class="fa fa-refresh fa-spin"></i>
      </div>

      <div class="box-body">
        <div class="row">
          <div class="col-xs-2">
            <div class="form-group">
              <label>Policy Version</label>
              <v-select placeholder="Select Policy Version" v-model="selectedPolicyVersion" :options.sync="policyVersions"></v-select>
            </div>
            <!-- /.form-group -->
          </div>
          <!-- /.col -->
          <div class="col-xs-2">
            <div class="form-group">
              <input type="checkbox" id="checkbox" v-model="compareEnabled"> <label>Compare Against</label>
              <v-select v-if="compareEnabled" placeholder="Select Policy Version" v-model="selectedPolicyVersionBase" :options.sync="policyVersions"></v-select>
            </div>
            <!-- /.form-group -->
          </div>
          <!-- /.col -->
          <div v-if="false" class="col-xs-4">
            <div class="form-group">
              <label>Namespace</label>
              <v-select placeholder="Select namespace" v-model="selectedNamespace" :options.sync="namespaces"></v-select>
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

    <v-diagram v-if="selectedPolicyVersion" :policyGen="selectedPolicyVersion" :policyGenBase="selectedPolicyVersionBaseComputed"></v-diagram>
  </div>
</template>

<script>
  import vSelect from 'vue-select'
  import vDiagram from 'pages/components/Diagram'
  import { getPolicy, getPolicyGeneration, getPolicyObjects, getNamespaces } from 'lib/api.js'

  export default {
    data () {
      return {
        compareEnabled: false,
        loading: false,
        policy: null,
        error: null,
        policyVersions: [],
        namespaces: [],
        selectedPolicyVersion: null,
        selectedPolicyVersionBase: null,
        selectedNamespace: null
      }
    },
    computed: {
      selectedPolicyVersionBaseComputed: function () {
        if (this.compareEnabled) {
          return this.selectedPolicyVersionBase
        }
        return null
      }
    },
    watch: {
      policy: function (data) {
        // once policy is loaded, create the list of namespaces for the dropdown
        this.namespaces = getNamespaces(getPolicyObjects(data))
        if (this.namespaces.length > 0) {
          // select first namespace
          this.selectedNamespace = this.namespaces[0]
        }

        // once policy is loaded, create the list of versions for the dropdown
        const generation = getPolicyGeneration(data)
        this.policyVersions = []
        for (let i = generation; i > 0; i--) {
          this.policyVersions.push(i.toString())
        }

        // pre-select dropdown values
        this.selectedPolicyVersion = this.policyVersions[0]
        if (this.policyVersions.length > 1) {
          this.selectedPolicyVersionBase = this.policyVersions[1]
        }
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchPolicy()
    },
    methods: {
      fetchPolicy () {
        this.loading = true
        this.policy = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          this.policy = data
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getPolicy(fetchSuccess, fetchError)
      }
    },
    components: {vSelect, vDiagram}
  }

</script>

<style>

</style>
