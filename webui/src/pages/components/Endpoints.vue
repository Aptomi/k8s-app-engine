<template>
  <div>
    <div class="box">
      <div class="box-header">
        <h3 class="box-title">Endpoints: <b>{{ dependency.namespace }} / {{ dependency.kind }} / {{ dependency.name }}</b></h3>
      </div>
      <div class="box-body table-responsive no-padding">
        <table class="table table-hover">
          <thead>
          <tr>
            <th>Component</th>
            <th>URL</th>
          </tr>
          </thead>
          <tbody>
          <tr v-if="endpoints == null || endpoints.length <= 0">
            <td>No Endpoints Available</td>
          </tr>
          <tr v-for="e, key in endpoints">
            <td>{{ key }}</td>
            <td>
              <ul v-for="url, proto in e">
                <li><a :href="url">{{proto}} - {{ url }}</a></li>
              </ul>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
      <!-- /.box-body -->
    </div>
    <!-- /.box -->

  </div>
</template>
<script>
  export default {
    data () {
      // empty data
      return {
        endpoints: null
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchData()
    },
    props: {
      'dependency': {
        type: Object,
        validator: function (value) {
          return true
        }
      }
    },
    methods: {
      fetchData () {
        this.endpoints = this.dependency['status']['endpoints']
      }
    }
  }
</script>
<style>
</style>
