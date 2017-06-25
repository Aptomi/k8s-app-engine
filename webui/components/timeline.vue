<template>
    <div class="row animated fadeInRight">
        <div class="col-lg-12">
            <div class="ibox float-e-margins">
                <div class="" id="ibox-content">
                    <div id="vertical-timeline" class="vertical-container light-timeline center-orientation"
                         v-if="timeline.length > 0">

                        <div class="vertical-timeline-block">
                            <div class="vertical-timeline-icon navy-bg">
                                <i class="fa fa-briefcase"></i>
                            </div>

                            <div class="vertical-timeline-content">
                                <h2>Now</h2>
                                <p>This Timeline visualizes all changes applied to Aptomi policies.</p>

                                <a class="btn btn-sm btn-default" v-on:click="toggle_style"> Toggle style</a>

                                <span class="vertical-date">
                                        Now <br/>
                                    <!--<small>Dec 24</small>-->
                                    </span>
                            </div>
                        </div>

                        <div class="vertical-timeline-block" v-for="item in timeline">
                            <div class="vertical-timeline-icon blue-bg">
                                <i class="fa fa-file-text"></i>
                            </div>

                            <div class="vertical-timeline-content">
                                <h2>
                                    Revision: {{ item.revisionNumber }} &nbsp;
                                    ({{ time_ago(item.createdOn) }}) &nbsp;
                                    (<a v-bind:href="'/run/' + item.dir + '/graphics/graph_delta.png'" target="_blank">delta graph</a>)
                                </h2>
                                <p><pre>{{ item.diff }}</pre></p>
                                <span class="vertical-date">
                                        <!--Today <br/>-->
                                        <small>{{ time_nice(item.createdOn) }}</small>
                                    </span>
                            </div>
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
                timeline: []
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

                loadJSON("/api/timeline-view", function (jsonData) {
                    ctx.timeline = jsonData;
                    console.log(ctx);
                }, function (err) {
                    console.log("/api/timeline-view not loaded with err:");
                    console.log(err);
                });
            },
            toggle_style: function (event) {
                $('#vertical-timeline').toggleClass('center-orientation');
            },
            time_ago: function (t) {
                return moment(t).fromNow()
            },
            time_nice: function(t) {
                return moment(t).format("dddd, MMMM Do YYYY, h:mm:ss a");
            }
        }
    }
</script>

<style>
    .hello {
        background-color: #ffe;
    }
</style>
