<template>
  <div>
    <div class="box box-default">
      <div class="box-header">
        <h3 class="box-title">View/Edit: <b>{{ obj.namespace }} / {{ obj.kind }} / {{ obj.name }}</b></h3>
        <button class="btn btn-sm btn-default" style="float: right" @click="editorCancel">Cancel</button>
        <button class="btn btn-sm btn-primary" style="float: right; margin-right: 5px" @click="editorSave">Save</button>
        <button class="btn btn-sm btn-danger" style="float: right; margin-right: 5px" @click="editorSave">Delete</button>
      </div>
      <!-- /.box-header -->
      <div class="overlay" v-if="loading">
        <i class="fa fa-refresh fa-spin"></i>
      </div>
      <!-- /.box-header -->
      <div class="box-body">
        <div class="row" v-if="error">
          <div class="col-xs-12">
            <span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i>
          </div>
        </div>
        <div class="row">
          <div class="col-md-12">
            <editor v-model="obj.yaml" @init="editorInit" lang="yaml" theme="tomorrow_night_eighties" :height="height"></editor>
          </div>
        </div>
        <!-- /.row -->
      </div>

      <!-- /.box-body -->
    </div>

  </div>
</template>
<script>
  import { fetchObjectProperties, savePolicyObjects } from 'lib/api.js'
  import hljs from 'highlight.js'
  import 'highlight.js/styles/agate.css'
  import yaml from 'js-yaml'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        error: null
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchData()
    },
    mounted () {
      // highlight
      $('pre code').each(function (i, block) {
        hljs.highlightBlock(block)
      })
    },
    props: ['obj', 'height'],
    watch: {
      'obj': 'fetchData'
    },
    methods: {
      editorInit: function () {
        require('vue2-ace-editor/node_modules/brace/mode/yaml')
        require('vue2-ace-editor/node_modules/brace/theme/tomorrow_night_eighties')
      },
      editorCancel: function () {
        this.$emit('close')
      },
      editorDelete: function () {
        alert(this.obj.yaml)
        this.$emit('close')
      },
      editorSave: function () {
        // trying to parse YAML first
        var parsedObj
        try {
          parsedObj = yaml.safeLoad(this.obj.yaml)
        } catch (e) {
          // keep plain alert here for now, so user can read the error for as long as needed before closing it
          alert('Invalid YAML: ' + e)
          return
        }

        const saveSuccess = $.proxy(function (data) {
          var message = data['policychanged'] ? 'Changed: version ' + (data['policygeneration'] - 1) + ' -> ' + data['policygeneration'] : 'No changes'
          this.$notify({
            group: 'main',
            type: 'success',
            title: 'Saved Successfully',
            text: message
          })
          this.$emit('close')
        }, this)

        const saveError = $.proxy(function (err) {
          // keep plain alert here for now, so user can read the error for as long as needed before closing it
          alert(err)
        }, this)

        savePolicyObjects(saveSuccess, saveError, parsedObj)
      },
      fetchData () {
        this.loading = true
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false

          // highlight
          $('pre code').each(function (i, block) {
            block.textContent = data['yaml']
            hljs.highlightBlock(block)
          })
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        fetchObjectProperties(this.obj, fetchSuccess, fetchError)
      }
    },
    components: {
      editor: require('vue2-ace-editor')
    }
  }
</script>
<style>

</style>
