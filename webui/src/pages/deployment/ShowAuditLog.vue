<template>
  <div>

    <div class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Audit Log</h3>
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
                  <th>Last Apply Status</th>
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
                    <!-- <span v-if="p['revisions'] == null || p['revisions'].length <= 0" class="label label-success" style="float:left; margin-right: 2px; margin-bottom: 2px">skipped</span> -->
                    <span @click="showEventLog('apply', r)" v-for="r in p['revisions']" class="label" v-bind:class="{ 'label-success': r['status'] === 'completed' && r['result']['failed'] <= 0, 'label-primary': r['status'] === 'inprogress', 'label-danger': r['status'] === 'error' || (r['status'] === 'completed' && r['result']['failed'] > 0) }" style="float:left; margin-right: 2px; margin-bottom: 2px">{{ r.metadata.generation }}</span>
                  </td>
                  <td class="align-middle">
                    <div v-for="r, index in p['revisions']" v-if="index === p['revisions'].length - 1" class="progress-group">
                      <!--
                      <div v-if="r['status'] === 'inprogress'" class="progress progress-xs progress-striped active">
                        <div class="progress-bar progress-bar-primary" v-bind:style="{ width: percent(r) + '%' }"></div>
                      </div>
                      <div v-if="r['status'] === 'completed'" class="progress progress-xs active">
                        <div class="progress-bar progress-bar-success" style="width: 100%"></div>
                      </div>
                      <div v-if="r['status'] === 'error'" class="progress progress-xs active">
                        <div class="progress-bar progress-bar-danger" style="width: 100%"></div>
                      </div>
                      <span class="progress-number"><b>{{ percent(r) }}%</b> ({{r['progress']['current']}}/{{r['progress']['total']}})</span>
                      -->

                      <!--
                      {{ r['result'] }}
                      {{ percents(r) }}
                      -->

                      <div class="progress">
                        <progress-bar v-bind:now="percents(r)['success']" v-bind:label="true" type="success" v-bind:striped="r['status'] === 'inprogress'" v-bind:animated="r['status'] === 'inprogress'"></progress-bar>
                        <progress-bar v-bind:now="percents(r)['failed']" v-bind:label="true" type="danger" v-bind:striped="r['status'] === 'inprogress'" v-bind:animated="r['status'] === 'inprogress'"></progress-bar>
                        <progress-bar v-bind:now="percents(r)['skipped']" v-bind:label="true" type="warning" v-bind:striped="r['status'] === 'inprogress'" v-bind:animated="r['status'] === 'inprogress'"></progress-bar>
                        <progress-bar v-bind:now="percents(r)['remaining']" v-bind:label="true" type="primary" v-bind:striped="r['status'] === 'inprogress'" v-bind:animated="r['status'] === 'inprogress'"></progress-bar>
                      </div>
                    </div>

                  </td>
                  <td v-if="p['revisions'].length <= 0">
                    <div>never</div>
                  </td>
                  <td v-for="r, index in p['revisions']" v-if="index === p['revisions'].length - 1">
                    <div v-if="r['status'] !== 'inprogress'">{{ r['appliedat'] | formatDateAgo }}</div>
                  </td>
                  <td class="col-xs-1">
                    <div class="btn-group btn-group-xs" style="position: absolute">
                      <button type="button" class="btn btn-default btn-flat">Action</button>
                      <button type="button" class="btn btn-default btn-flat dropdown-toggle" data-toggle="dropdown">
                        <span class="caret"></span>
                        <span class="sr-only">Toggle Dropdown</span>
                      </button>
                      <ul class="dropdown-menu" style="left: auto !important; right: 0 !important;" v-for="r, index in p['revisions']" v-if="index === p['revisions'].length - 1">
                        <li><router-link :to="{ name: 'BrowsePolicy', params: { inMode: 'policy', inPolicyVersion: p['metadata']['generation'].toString() }}">Browse Policy</router-link></li>
                        <li><router-link :to="{ name: 'BrowsePolicy', params: { inMode: 'policy', inPolicyVersion: p['metadata']['generation'].toString(), inCompareEnabled: true, inPolicyVersionBase: (p['metadata']['generation'] - 1).toString() }}">Compare With Previous</router-link></li>
                        <li class="divider"></li>
                        <li><a href="#" @click="showEventLog('resolve', r)">View Resolve Log</a></li>
                        <li><a href="#" @click="showEventLog('apply', r)">View Apply Log</a></li>
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
  import {getAllPolicies, fetchPolicy, fetchPolicyRevisions} from 'lib/api.js'
  import eventLog from 'pages/components/EventLog'
  import progressBar from 'vue-strap/src/Progressbar'

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
        // if there are no policies, keep refreshing data
        if (this.policies == null || this.policies.length <= 0) {
          this.fetchData()
          return
        }

        // if recent policy has changed, then re-fetch the data
        let pCur = this.policies[0]
        let pNext = {}
        fetchPolicy(0, pNext)
        if (pCur['metadata']['generation'] !== pNext['metadata']['generation']) {
          this.fetchData()
          return
        }

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
      percents (r) {
        if (r['result']['total'] <= 0) {
          return {
            'success': 100,
            'failed': 0,
            'skipped': 0,
            'remaining': 0
          }
        }
        let p1 = Math.floor(100.0 * r['result']['success'] / r['result']['total'])
        let p2 = Math.floor(100.0 * r['result']['failed'] / r['result']['total'])
        let p3 = Math.floor(100.0 * r['result']['skipped'] / r['result']['total'])
        let p4 = Math.max(0, 100 - (p1 + p2 + p3))

        if (p4 > 0 && (r['status'] === 'completed' || r['status'] === 'failed')) {
          // there should not be remaining percentage left
          if (p1 > 0) {
            p1 += p4
          } else if (p2 > 0) {
            p2 += p4
          } else if (p3 > 0) {
            p3 += p4
          }
          p4 = 0
        }

        return {
          'success': p1,
          'failed': p2,
          'skipped': p3,
          'remaining': p4
        }
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
      },
      showEventLog (type, obj) {
        this.$modal.show(eventLog, {
          revision: obj,
          type: type
        }, {
          width: '60%',
          height: 'auto'
        })
      }
    },
    beforeDestroy: function () {
      clearInterval(this.interval)
    },
    components: {
      progressBar: progressBar
    }
  }
</script>

<style>

</style>
