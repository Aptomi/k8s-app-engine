<template>
  <div>

    <div class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Policy Revisions</h3>
          </div>
          <!-- /.box-header -->
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Created By</th>
                  <th>Date</th>
                  <th>Apply Status</th>
                  <th>Apply Progress</th>
                  <th>Last Apply Run</th>
                  <th>Action</th>
                </tr>
                <tr>
                  <td>4</td>
                  <td>Alice</td>
                  <td>11-7-2014</td>
                  <td><span class="label label-primary">Applying (41/100)</span></td>
                  <td>
                    <div class="progress progress-xs progress-striped active">
                      <div class="progress-bar progress-bar-primary" style="width: 40%"></div>
                    </div>
                  </td>
                  <td>2 min ago</td>
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
              </thead>
              <tbody>
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
                <tr>
                  <td>1</td>
                  <td>John</td>
                  <td>11-7-2014</td>
                  <td><span class="label label-success">Success</span></td>
                  <td>
                    <div class="progress progress-xs active">
                      <div class="progress-bar progress-bar-success" style="width: 100%"></div>
                    </div>
                  </td>
                  <td>2 hours ago</td>
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
  import {getUsersAndRoles} from 'lib/api.js'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        users: null,
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
        this.users = null
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          console.log(data)
          this.loading = false
          this.users = data
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        getUsersAndRoles(fetchSuccess, fetchError)
      }
    }
  }
</script>

<style>

</style>
