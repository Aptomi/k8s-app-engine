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
                  <th>Created By</th>
                  <th>Date</th>
                  <th>Apply Revisions</th>
                  <th>Apply Status</th>
                  <th>Last Applied</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="error">
                  <td><span class="label label-danger center">Error</span> <i class="text-red">{{ error }}</i></td>
                </tr>
                <tr v-for="p in policies">
                  <td>{{ p.metadata.generation }}</td>
                  <td>{{ p.createdBy }}</td>
                  <td>{{ p.createdOn }}</td>
                  <td>
                    <span class="label label-success">7</span>
                    <span class="label label-success">8</span>
                    <span class="label label-danger">9</span>
                    <span class="label label-success">10</span>
                    <span class="label label-primary">11</span>
                  </td>
                  <td class="align-middle">
                    <div class="progress-group">
                      <div class="progress progress-xs progress-striped active">
                        <div class="progress-bar progress-bar-primary" style="width: 40%"></div>
                      </div>
                      <span class="progress-number"><b>160</b>/200</span>
                    </div>
                  </td>
                  <td>{{ p.lastApplied }}</td>
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
                <tr>
                  <td>3</td>
                  <td>Bob</td>
                  <td>11-7-2014</td>
                  <td><span class="label label-success center">Success</span></td>
                  <td>
                    <div class="progress progress-xs active">
                      <div class="progress-bar progress-bar-success" style="width: 100%"></div>
                    </div>
                  </td>
                  <td>10 min ago</td>
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
                <tr>
                  <td>2</td>
                  <td>Frank</td>
                  <td>11-7-2014</td>
                  <td><span class="label label-danger">Failed</span></td>
                  <td>
                    <div class="progress progress-xs active">
                      <div class="progress-bar progress-bar-danger" style="width: 100%"></div>
                    </div>
                  </td>
                  <td>1 hour ago</td>
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
  import {getAllPolicies} from 'lib/api.js'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        policies: null,
        error: null
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchData()
    },
    methods: {
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
    }
  }
</script>

<style>

</style>
