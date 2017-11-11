<template>
  <div>

    <!-- /.row -->
    <div class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Declared Dependencies</h3>
          </div>
          <!-- /.box-header -->
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <tr>
                <th>Name</th>
                <th>User</th>
                <th>Contract</th>
                <th>Status</th>
                <th>Action</th>
              </tr>
              <tr>
                <td>alice-stage</td>
                <td>Alice</td>
                <td>twitter-stats</td>
                <td><span class="label label-success">Deployed</span></td>
                <td>2 min ago</td>
                <td>
                  <button type="button" class="btn btn-default btn-xs">Show Endpoints</button>
                </td>
              </tr>
              <tr>
                <td>bob-stage</td>
                <td>Bob</td>
                <td>twitter-stats</td>
                <td><span class="label label-primary center">Processing</span></td>
                <td>10 min ago</td>
                <td>
                </td>
              </tr>
              <tr>
                <td>carol-stage</td>
                <td>Carol</td>
                <td>twitter-stats</td>
                <td><span class="label label-danger center">Not Resolved</span></td>
                <td>10 min ago</td>
                <td>
                </td>
              </tr>
              <tr>
                <td>prod</td>
                <td>John</td>
                <td>twitter-stats</td>
                <td><span class="label label-success">Deployed</span></td>
                <td>2 hours ago</td>
                <td>
                  <button type="button" class="btn btn-default btn-xs">Show Endpoints</button>
                </td>
              </tr>
            </table>
          </div>
          <!-- /.box-body -->
        </div>
        <!-- /.box -->
      </div>
    </div>

    <!-- /.row -->
    <div class="row">
      <div class="col-xs-12">
        <div class="box">
          <div class="box-header">
            <h3 class="box-title">Endpoints for <b>alice-stage</b></h3>
          </div>
          <!-- /.box-header -->
          <div class="box-body table-responsive no-padding">
            <table class="table table-hover">
              <tr>
                <th>Component</th>
                <th>URL</th>
              </tr>
              <tr>
                <td>zookeeper</td>
                <td><a href="#">http://127.0.0.1:12345</a></td>
              </tr>
              <tr>
                <td>hdfs</td>
                <td><a href="#">http://127.0.0.1:23456</a></td>
              </tr>
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
export default {
  name: 'show-dependencies',
  created () {

  }
}
</script>

<style>

</style>
