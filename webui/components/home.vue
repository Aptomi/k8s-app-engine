<template>
    <div>
        <div class="row animated fadeInUp">
            <div class="col-lg-12">
                <div class="ibox float-e-margins" v-if="summary.globalDependencies.length > 0">
                    <div class="ibox-title">
                        <h5>Global List of services in use</h5>
                    </div>
                    <div class="ibox-content">
                        <div class="table-responsive">
                            <table class="table table-striped">

                                <!--{-->
                                <!--"cluster": "cluster-us-west",-->
                                <!--"context": "stage/dev-Alice-stage",-->
                                <!--"dependencyId": "alice_tests_twitter_stats_in_stage",-->
                                <!--"id": "Alice",-->
                                <!--"resolved": true,-->
                                <!--"serviceName": "twitter_stats",-->
                                <!--"stats": "-1 containers/1 hour running",-->
                                <!--"userName": "Alice"-->
                                <!--},-->

                                <thead>
                                <tr>
                                    <th>Status</th>
                                    <th>User</th>
                                    <th>Consumes</th>
                                    <th>Context</th>
                                    <th>Cluster</th>
                                    <th>Uptime</th>
                                    <th>Action</th>
                                </tr>
                                </thead>
                                <tbody>
                                <tr v-for="item in summary.globalDependencies">
                                    <td v-if="item.resolved"><span class="label label-primary">Running</span></td>
                                    <td v-else><span class="label label-danger">Not Running</span></td>
                                    <td>{{ item.userName }}</td>
                                    <td>{{ item.serviceName }}</td>
                                    <td>{{ item.context }}</td>
                                    <td>{{ item.cluster }}</td>
                                    <td>{{ item.stats }}</td>
                                    <td>
                                        <!--<router-link v-bind:to="{ name: 'details', params: { view: 'globalops', filter: item.dependencyId } }">-->
                                            <!--[Open In Policy Explorer]-->
                                        <!--</router-link>-->
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                        </div>

                    </div>
                </div>
            </div>
        </div>

        <div class="row animated fadeInUp">
            <div class="col-lg-12">
                <div class="ibox float-e-margins" v-if="summary.globalRules.length > 0">
                    <div class="ibox-title">
                        <h5>Global Rules</h5>
                    </div>
                    <div class="ibox-content">
                        <div class="table-responsive">
                            <table class="table table-striped">

                                <!--{-->
                                <!--"appliedTo": "-1 instances",-->
                                <!--"id": "compromised_users_forbid_all_services",-->
                                <!--"ruleName": "compromised_users_forbid_all_services",-->
                                <!--"ruleObject": {-->
                                <!--"Cluster": null,-->
                                <!--"Labels": null,-->
                                <!--"User": {-->
                                <!--"Accept": [-->
                                <!--"deactivated"-->
                                <!--],-->
                                <!--"Reject": []-->
                                <!--}-->
                                <!--}-->
                                <!--},-->

                                <thead>
                                <tr>
                                    <th>Status</th>
                                    <th>Rule</th>
                                    <th>Conditions</th>
                                    <th>Actions</th>
                                    <th>Applied to</th>
                                </tr>
                                </thead>
                                <tbody>
                                <tr v-for="item in summary.globalRules">
                                    <td v-if="item.matchedUsers.length > 0"><span class="label label-primary">Active</span></td>
                                    <td v-else><span class="label label-warning">Inactive</span></td>
                                    <td>{{ item.ruleName }}</td>
                                    <td>
                                        <p v-for="(details, condition) in item.conditions">
                                            {{ condition }}
                                        <ul>
                                            <li v-for="detail in details">{{ detail }}</li>
                                        </ul>
                                        </p>
                                    </td>
                                    <td>
                                        <p v-for="item in item.actions">{{ item }}</p>
                                    </td>
                                    <td>
                                        <p v-for="item in item.matchedUsers">{{ item.Name }} (ID: {{ item.ID }})</p>
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                        </div>

                    </div>
                </div>
            </div>
        </div>

        <div class="row animated fadeInUp">
            <div class="col-lg-12">
                <div class="ibox float-e-margins" v-if="summary.servicesOwned.length > 0">
                    <div class="ibox-title">
                        <h5>Services I own</h5>
                    </div>
                    <div class="ibox-content">
                        <div class="table-responsive">
                            <table class="table table-striped">

                                <!--{-->
                                <!--"cluster": "cluster-us-east",-->
                                <!--"context": "prod/production",-->
                                <!--"id": "twitter_stats.prod.production.root",-->
                                <!--"serviceName": "twitter_stats",-->
                                <!--"stats": "-1 containers/1 hour running"-->
                                <!--}-->

                                <thead>
                                <tr>
                                    <th>Status</th>
                                    <th>Service</th>
                                    <th>Context</th>
                                    <th>Cluster</th>
                                    <th>Uptime</th>
                                    <th>Action</th>
                                </tr>
                                </thead>
                                <tbody>
                                <tr v-for="item in summary.servicesOwned">
                                    <td><span class="label label-primary">Running</span></td>
                                    <td>{{ item.serviceName }}</td>
                                    <td>{{ item.context }}</td>
                                    <td>{{ item.cluster }}</td>
                                    <td>{{ item.stats }}</td>
                                    <td>
                                        <!--<router-link v-bind:to="{ name: 'details', params: { view: 'service', filter: item.serviceName } }">-->
                                            <!--[Open In Policy Explorer]-->
                                        <!--</router-link>-->
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                        </div>

                    </div>
                </div>
            </div>
        </div>

        <div class="row animated fadeInUp">
            <div class="col-lg-12">
                <div class="ibox float-e-margins" v-if="summary.servicesUsing.length > 0">
                    <div class="ibox-title">
                        <h5>Services I consume</h5>
                    </div>
                    <div class="ibox-content">
                        <div class="table-responsive">
                            <table class="table table-striped">

                                <!--{-->
                                <!--"cluster": "cluster-us-west",-->
                                <!--"context": "stage/dev-Alice-stage",-->
                                <!--"dependencyId": "alice_tests_twitter_stats_in_stage",-->
                                <!--"id": "alice_tests_twitter_stats_in_stage",-->
                                <!--"resolved": true,-->
                                <!--"serviceName": "twitter_stats",-->
                                <!--"stats": "-1 containers/1 hour running"-->
                                <!--},-->

                                <thead>
                                <tr>
                                    <th>Status</th>
                                    <th>Service</th>
                                    <th>Context</th>
                                    <th>Cluster</th>
                                    <th>Uptime</th>
                                    <th>Action</th>
                                </tr>
                                </thead>
                                <tbody>
                                <tr v-for="item in summary.servicesUsing">
                                    <td v-if="item.resolved"><span class="label label-primary">Running</span></td>
                                    <td v-else><span class="label label-danger">Not Running</span></td>
                                    <td>{{ item.serviceName }}</td>
                                    <td>{{ item.context }}</td>
                                    <td>{{ item.cluster }}</td>
                                    <td>{{ item.stats }}</td>
                                    <td>
                                        <!--<router-link v-bind:to="{ name: 'details', params: { view: 'consumer', filter: item.dependencyId } }">-->
                                            <!--[Open In Policy Explorer]-->
                                        <!--</router-link>-->
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                        </div>

                    </div>
                </div>
            </div>
        </div>

        <div class="row animated fadeInUp" v-for="user_endpoints in all_endpoints">
            <div class="col-lg-12">
                <div class="ibox float-e-margins">
                    <div class="ibox-title">
                        <h5 v-if="all_endpoints.length > 1 && user_endpoints.Endpoints.length > 0">
                            Services Endpoints for user {{ user_endpoints.User.Name }}
                        </h5>
                        <h5 v-if="all_endpoints.length > 1 && user_endpoints.Endpoints.length == 0">
                            No Services Endpoints for user {{ user_endpoints.User.Name }}
                        </h5>
                        <h5 v-if="all_endpoints.length == 1 && user_endpoints.Endpoints.length > 0">
                            Services Endpoints
                        </h5>
                        <h5 v-if="all_endpoints.length == 1 && user_endpoints.Endpoints.length == 0">
                            No Services Endpoints
                        </h5>
                    </div>
                    <div class="ibox-content" v-if="user_endpoints.Endpoints.length > 0">
                        <div class="table-responsive">
                            <table class="table table-striped">
                                <thead>
                                <tr>
                                    <th>Status</th>
                                    <th>Service</th>
                                    <th>Context</th>
                                    <th>Component</th>
                                    <th>Links</th>
                                </tr>
                                </thead>
                                <tbody>
                                <tr v-for="endpoint in user_endpoints.Endpoints">
                                    <td><span class="label label-primary">HTTP OK</span></td>
                                    <td>{{ endpoint.Service }}</td>
                                    <td>{{ endpoint.Context }}/{{ endpoint.Allocation }}</td>
                                    <td>{{ endpoint.Component }}</td>
                                    <td>
                                        <div v-for="link in endpoint.Links">
                                            {{ link.Name }}: <a v-bind:href="link.Link">{{ link.Link }}</a>
                                        </div>
                                    </td>
                                </tr>
                                </tbody>
                            </table>
                        </div>
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
                all_endpoints: {},
                summary: {
                    globalDependencies: [],
                    globalRules: [],
                    servicesOwned: [],
                    servicesUsing: []
                }
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

                loadJSON("/api/summary-view", function (jsonData) {
                    ctx.summary = jsonData;
                    console.log(ctx);
                }, function (err) {
                    console.log("/api/summary-view not loaded with err:");
                    console.log(err);
                });

                loadJSON("/api/endpoints", function (jsonData) {
                    ctx.all_endpoints = jsonData.Endpoints;
                    console.log(ctx.all_endpoints);
                }, function (err) {
                    console.log("/api/endpoints not loaded with err:");
                    console.log(err);
                })
            }
        }
    }
</script>
