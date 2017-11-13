<template>
  <div>
    <div class="box box-default">
      <div class="box-header">
        <h3 class="box-title">View Object: <b>{{ obj.namespace }} / {{ obj.kind }} / {{ obj.name }}</b></h3>
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
            <pre><code class="yaml">{{ obj.yaml }}</code></pre>
          </div>
        </div>
        <!-- /.row -->
      </div>

      <!-- /.box-body -->
    </div>
  </div>
</template>
<script>
  import { fetchObjectProperties } from 'lib/api.js'
  import hljs from 'highlight.js'
  import 'highlight.js/styles/agate.css'

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
    props: {
      'obj': {
        type: Object,
        validator: function (value) {
          return true
        }
      }
    },
    watch: {
      'obj': 'fetchData'
    },
    methods: {
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
    }
  }
</script>
<style>

</style>
