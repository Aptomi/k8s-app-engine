<template>
    <div>
        <div class="row animated fadeInUp">
            <div class="col-lg-12">
                <div class="ibox float-e-margins">
                    <div class="ibox-content">
                        <div class="row m-b-xs">
                            <div class="col-sm-10">
                                <span class="m-l-xs m-r-sm">View: </span>
                                <details-filter v-bind:items="views" v-bind:current_title="curr_view_title" target_view=""/>

                                <span v-if="items.length > 0" class="m-l-sm m-r-sm">{{ filter_name }}: </span>
                                <details-filter v-if="items.length > 0"
                                                v-bind:items="items" v-bind:current_title="curr_filter_title" v-bind:target_view="curr_view"/>

                            </div>
                            <div class="col-sm-2">
                                <button type="button" class="btn btn-primary" id="rule-log-button"
                                        data-toggle="modal" data-target="#details-modal" style="display: none;">
                                    Rule log
                                </button>
                            </div>
                        </div>
                        <div id="mygraph"></div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    module.exports = {
        data: function () {
            return {
                curr_view: "",
                curr_filter: "",

                curr_view_title: "",
                curr_filter_title: "",

                filter_name: "",

                views: [],

                items: []
            }
        },
        created: function () {
            this.fetch_data()
        },
        watch: {
            '$route': 'fetch_data'
        },
        methods: {
            fetch_data: function () {
                var ctx = this;

                $("#mygraph").html("");

                loadJSON("/api/details", function (jsonData) {
                    var view = ctx.$route.params.view;
                    var filter = ctx.$route.params.filter;

                    app.obj_view = [];
                    ctx.views = jsonData.Views;

                    if (view === ":view") {
                        router.push({ name: 'details', params: { view: ctx.views[0].name }});
                        return
                    }

                    ctx.curr_view = view;
                    ctx.curr_filter = filter;

                    var userId = "";
                    var dependencyId = "";
                    var serviceName = "";

                    if (view === "service") {
                        ctx.items = jsonData.Services;
                        ctx.filter_name = "Service";

                        // service-view = view from the standpoint of service owner
                        serviceName = filter
                    } else if (view === "consumer") {
                        ctx.items = jsonData.Dependencies;
                        ctx.filter_name = "Dependency";

                        // consumer-view = view from the standpoint of service consumer
                        userId = jsonData.UserId
                        dependencyId = filter === "all" ? "" : filter
                    } else if (view === "globalops"){
//                        ctx.items = jsonData.Users;
//                        ctx.filter_name = "User";
//                        userId = filter === "all" ? "" : filter

                        ctx.items = jsonData.AllDependencies;
                        ctx.filter_name = "Dependency";
                        dependencyId = filter === "all" ? "" : filter

                        // globalops-view = view from the standpoint of global IT ops

                    }

                    ctx.views.forEach(function(item) {
                        if (item.name === ctx.curr_view) {
                            ctx.curr_view_title = item.title
                        }
                    });

                    ctx.items.forEach(function(item) {
                        if (item.name === ctx.curr_filter) {
                            ctx.curr_filter_title = item.title
                        }
                    });

                    if (filter === ":filter") {
                        if (ctx.items.length > 0) {
                            router.push({name: 'details', params: {view: view, filter: ctx.items[0].name}})
                        }
                        return
                    }

                    drawChart(ctx.curr_view, userId, dependencyId, serviceName)
                }, function (err) {
                    console.log("/api/details not loaded with err:");
                    console.log(err)
                });
            }
        }
    }
</script>
