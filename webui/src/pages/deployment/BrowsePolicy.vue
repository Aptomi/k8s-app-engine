<template>
  <div>

    <div class="box box-default">

      <div class="overlay" v-if="loading">
        <i class="fa fa-refresh fa-spin"></i>
      </div>

      <div class="box-body">
        <div class="row">
          <div class="col-xs-3">
            <div class="form-group">
              <label>Show</label>
              <v-select placeholder="Select Mode" v-model="selectedMode" :options="modes" track-by="mode" label="name" :searchable="false" :allow-empty="false" deselect-label="Selected"></v-select>
            </div>
            <!-- /.form-group -->
          </div>
          <div class="col-xs-3">
            <div class="form-group" v-if="selectedMode['mode'] !== 'actual'">
              <label>Policy Version</label>
              <v-select placeholder="Select Policy Version" v-model="selectedPolicyVersion" :options.sync="policyVersions" :allow-empty="false" deselect-label="Selected"></v-select>
            </div>
            <!-- /.form-group -->
          </div>
          <!-- /.col -->
          <div class="col-xs-3">
            <div class="form-group" v-if="selectedMode['mode'] !== 'actual'">
              <input type="checkbox" id="checkbox" v-model="compareEnabled"> <label>Compare With</label>
              <v-select v-if="compareEnabled" placeholder="Select Policy Version" v-model="selectedPolicyVersionBase" :options.sync="policyVersions" :allow-empty="false" deselect-label="Selected"></v-select>
            </div>
            <!-- /.form-group -->
          </div>
          <!-- /.col -->
          <div v-if="false" class="col-xs-3">
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

    <v-diagram v-if="selectedPolicyVersion" :mode="selectedMode['mode']" :policyGen="selectedPolicyVersion" :policyGenBase="selectedPolicyVersionBaseComputed"></v-diagram>
  </div>
</template>

<script>
  import vSelect from 'vue-multiselect'
  import vDiagram from 'pages/components/Diagram'
  import { getPolicy, getPolicyGeneration, getPolicyObjectRefMap, getNamespacesByRefMap } from 'lib/api.js'

  export default {
    data () {
      return {
        loading: false,
        policy: null,
        error: null,
        policyVersions: [],
        namespaces: [],
        selectedMode: this.inMode,
        selectedPolicyVersion: this.inPolicyVersion,
        compareEnabled: this.inCompareEnabled,
        selectedPolicyVersionBase: this.inPolicyVersionBase,
        selectedNamespace: this.inNamespace
      }
    },
    computed: {
      modes: function () {
        return [
          { name: 'Policy', mode: 'policy' },
          { name: 'Desired State', mode: 'desired' },
          { name: 'Actual State', mode: 'actual' }
        ]
      },
      inModeComputed: function () {
        for (const idx in this.modes) {
          if (this.modes[idx]['mode'] === this.inMode) {
            return this.modes[idx]
          }
        }
        return this.modes[0]
      },
      selectedPolicyVersionBaseComputed: function () {
        if (this.compareEnabled && this.selectedMode['mode'] !== 'actual') {
          return this.selectedPolicyVersionBase
        }
        return null
      }
    },
    props: {
      'inMode': {
        type: String
      },
      'inPolicyVersion': {
        type: String
      },
      'inCompareEnabled': {
        type: Boolean
      },
      'inPolicyVersionBase': {
        type: String
      },
      'inNamespace': {
        type: String
      }
    },
    watch: {
      compareEnabled: function (data) {
        // one checkbox is checked, pre-select policy version to compare with
        if (data) {
          this.selectedPolicyVersionBase = (this.selectedPolicyVersion - 1).toString()
        }
      },
      policy: function (data) {
        // once policy is loaded, create the list of namespaces for the dropdown
        this.namespaces = getNamespacesByRefMap(getPolicyObjectRefMap(data))

        // once policy is loaded, create the list of versions for the dropdown
        const generation = getPolicyGeneration(data)
        this.policyVersions = []
        for (let i = generation; i > 0; i--) {
          this.policyVersions.push(i.toString())
        }

        if (this.policyVersions.length > 0) {
          // pre-select dropdown values -> policy version
          if (this.selectedPolicyVersion == null) {
            this.selectedPolicyVersion = this.policyVersions[0]
          }
        }

        // pre-select dropdown values -> namespace
        if (this.namespaces.length > 0) {
          if (this.selectedNamespace == null) {
            this.selectedNamespace = this.namespaces[0]
          }
        }
      }
    },
    created () {
      // pre-select dropdown values -> mode
      this.selectedMode = this.inModeComputed

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
    components: {
      vSelect,
      vDiagram
    }
  }
</script>

<style>
</style>
