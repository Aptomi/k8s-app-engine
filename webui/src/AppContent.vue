<template>
  <!-- Content Wrapper. Contains page content -->
  <div id="content-wrap" class="content-wrapper">
    <!-- Content Header (Page header) -->
    <section class="content-header">
      <h1>
        {{ $route.matched[0].name }}
        <small></small>
      </h1>
      <ol class="breadcrumb">
        <li v-for="route in $route.matched">
          {{route.name}}
        </li>
      </ol>
    </section>

    <section class="content container-fluid">
      <transition name="page" mode="out-in">
        <router-view></router-view>
      </transition>
    </section>
    <!-- /.content -->
  </div>
  <!-- /.content-wrapper -->
</template>

<script>
export default {
  name: 'va-content-wrap',
  created () {

  }
}
</script>

<style>
.page-enter-active, .page-leave-active {
  transition: opacity 0s, transform 0s;
}
.page-enter, .page-leave-to {
  opacity: 0;
}
</style>
