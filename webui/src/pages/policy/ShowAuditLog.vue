<template>
  <div>

    <div class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Policy</h3>
          </div>
          <!-- /.box-header -->
          <div class="overlay" v-if="loading">
            <i class="fa fa-refresh fa-spin"></i>
          </div>
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <thead>
                <tr>
                  <th>Version</th>
                  <th>Who</th>
                  <th>When</th>
                  <th>Apply Revisions</th>
                  <th>Apply Status</th>
                  <th>Last Apply Run</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="error">
                  <td><span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i></td>
                </tr>
                <tr v-for="p in policies">
                  <td>{{ p['metadata']['generation'] }}</td>
                  <td>{{ p['metadata']['updatedby'] }}</td>
                  <td>{{ p['metadata']['updatedat'] | formatDateAgo }} <small>({{ p['metadata']['updatedat'] | formatDate }})</small></td>
                  <td class="col-xs-4">
                    <span v-if="Object.keys(p['revisions']).length <= 0" class="label label-warning">N/A</span>
                    <span v-for="r in p['revisions']" class="label" v-bind:class="{ 'label-success': r['status'] === 'success', 'label-primary': r['status'] === 'inprogress', 'label-danger': r['status'] === 'error' }" style="float:left; margin-right: 2px; margin-bottom: 2px">{{ r.metadata.generation }}</span>
                  </td>
                  <td class="align-middle">
                    <div v-for="r, index in p['revisions']" v-if="index === p['revisions'].length - 1" class="progress-group">
                      <div v-if="r['status'] === 'inprogress'" class="progress progress-xs progress-striped active">
                        <div class="progress-bar progress-bar-primary" v-bind:style="{ width: percent(r) + '%' }"></div>
                      </div>
                      <div v-if="r['status'] === 'success'" class="progress progress-xs active">
                        <div class="progress-bar progress-bar-success" style="width: 100%"></div>
                      </div>
                      <div v-if="r['status'] === 'error'" class="progress progress-xs active">
                        <div class="progress-bar progress-bar-danger" style="width: 100%"></div>
                      </div>
                      <span class="progress-number"><b>{{ percent(r) }}%</b> ({{r['progress']['current']}}/{{r['progress']['total']}})</span>
                    </div>
                  </td>
                  <td v-if="p['revisions'].length <= 0">
                    <div>never</div>
                  </td>
                  <td v-for="r, index in p['revisions']" v-if="index === p['revisions'].length - 1">
                    <div v-if="r['status'] !== 'inprogress'">{{ r['appliedat'] | formatDateAgo }}</div>
                  </td>
                  <td>
                    <div class="btn-group btn-group-xs">
                      <button type="button" class="btn btn-default btn-flat">Action</button>
                      <button type="button" class="btn btn-default btn-flat dropdown-toggle" data-toggle="dropdown">
                        <span class="caret"></span>
                        <span class="sr-only">Toggle Dropdown</span>
                      </button>
                      <ul class="dropdown-menu" role="menu">
                        <li><a href="#">Browse Policy</a></li>
                        <li><a href="#">Compare With Previous</a></li>
                        <li class="divider"></li>
                        <li><a href="#">View Resolution Log</a></li>
                        <li><a href="#">View Apply Log</a></li>
                      </ul>
                    </div>
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
    <!-- /.row -->

  </div>
</template>

<script>
  import {getAllPolicies, fetchPolicyRevisions} from 'lib/api.js'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        policies: null,
        error: null,
        interval: null
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchData()
      this.interval = setInterval($.proxy(function () {
        // continue to fetch progress information for all the recent policies (unprocessed policies ... last policy with revisions)
        if (this.policies != null) {
          for (const idx in this.policies) {
            const p = this.policies[idx]
            const hasRevisions = (p['revisions'] != null) && (p['revisions'].length > 0)
            fetchPolicyRevisions(p['metadata']['generation'], p)
            if (hasRevisions) {
              break
            }
          }
        }
      }, this), 5000)
    },
    methods: {
      percent (r) {
        return Math.round(100.0 * r['progress']['current'] / r['progress']['total'])
      },
      fetchData () {
        this.loading = true
        this.policies = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          this.policies = data
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getAllPolicies(fetchSuccess, fetchError)
      }
    },
    beforeDestroy: function () {
      clearInterval(this.interval)
    }
  }
</script>

<style>

</style>
