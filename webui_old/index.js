// JS for /timeline
const CompTimeline = Vue.component('timeline', {
    template: '#template-timeline',
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
});

// JS for /details/filter
const CompDetailsFilter = Vue.component('details-filter', {
        template: '#template-details-filter',
        props: ['items', 'current_title', 'target_view']
});

// JS for drawing charts on /details
const drawChart = function (inputView, userId, dependencyId, serviceName) {
    // can be service-view, consumer-view, or globalops-view
    var view = inputView + "-view";

    var apiPath = "/api/" + view;
    if (view === "service-view") {
        // service-view = view from the standpoint of service owner
        apiPath += "?serviceName=" + serviceName;
    } else if (view === "consumer-view") {
        // consumer-view = view from the standpoint of service consumer
        apiPath += "?userId=" + userId + "&dependencyId=" + dependencyId;
    } else {
        // globalops-view = view from the standpoint of global IT ops
        apiPath += "?userId=" + userId + "&dependencyId=" + dependencyId;
    }

    loadJSON(apiPath, jsonLoaded, function (err) {
        console.log(err)
    });

    // create a network
    var container = document.getElementById('details_graph');
    var options = {
        nodes: {
            font: {
                size: 12,
                color: 'white'
            },
            borderWidth: 2,
            chosen: {
                label: false,
                node: chosenNode
            }
        },
        edges: {
            width: 1,
            font: {
                size: 12,
                strokeWidth: 0,
                color: 'white',
                align: 'top'
            }
        },
        groups: {
            service: {
                shape: 'icon',
                icon: {
                    face: 'FontAwesome',
                    code: '\uf1b2',
                    size: 50,
                    color: 'red'
                },
                color: {
                    border: 'red'
                }
            },
            serviceInstancePrimary: {
                font: {
                    color: 'black',
                    multi: 'html'
                },
                color: {background: 'rgb(0,255,140)', border: 'darkslategrey'},
                shape: 'box'
            },
            serviceInstance: {
                font: {
                    color: 'black',
                    multi: 'html'
                },
                color: {background: 'rgb(0,123,199)', border: 'darkslategrey'},
                shape: 'box'
            },
            dependencyShort: {
                shape: 'icon',
                icon: {
                    face: 'FontAwesome',
                    code: '\uf007',
                    size: 50,
                    color: 'orange'
                }
            },
            dependencyLongResolved: {
                shape: 'icon',
                font: {
                    multi: 'html'
                },
                icon: {
                    face: 'FontAwesome',
                    code: '\uf234',
                    size: 50,
                    color: 'orange'
                }
            },
            dependencyLongNotResolved: {
                shape: 'icon',
                font: {
                    multi: 'html'
                },
                icon: {
                    face: 'FontAwesome',
                    code: '\uf235',
                    size: 50,
                    color: 'red'
                }
            },
            mints: {color: 'rgb(0,255,140)'},
            source: {
                color: {border: 'white'}
            }
        },
        layout: {
            randomSeed: 239,
            hierarchical: {
                direction: "LR",
                levelSeparation: 220
            }
        },
        interaction: {
            hover: true,
            navigationButtons: true,
            keyboard: true
        },
        physics: false
    };

    // hack to better display consumer and globalops views
    if (view === "consumer-view" || view === "globalops-view") {
        options.layout.hierarchical.levelSeparation = 130;
    }

    function chosenNode(values, id, selected, hovering) {
        values.color = "#ffdd88";
        values.borderColor = "#ff0000";
    }

    function clickedNode(params) {
        params.event = "[original event]";
        var node = this.getNodeAt(params.pointer.DOM);
        var edge = this.getEdgeAt(params.pointer.DOM);

        var id = "";
        if (node) {
            id = node
        } else if (edge) {
            id = edge
        }
        if (id) {
            loadJSON("/api/object-view?id=" + id, objectLoaded, objectNotLoaded);

            $("#rule-log-button").click();
        } else {
            app.obj_view = [];
        }
    }

    function jsonLoaded(jsonData) {
        var data = {
            nodes: jsonData.nodes,
            edges: jsonData.edges
        };

        var network = new vis.Network(container, data, options);
        network.on("click", clickedNode);
        network.fit();
    }

    function objectLoaded(jsonData) {
        if (jsonData) {
            app.obj_view = [jsonData];
        } else {
            app.obj_view = [];
        }
    }

    function objectNotLoaded(err) {
        app.obj_view = [];
    }
};

// JS for /details
const CompDetails = Vue.component('details', {
        template: '#template-details',
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

                $("#details_graph").html("");

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

                        serviceName = filter
                    } else if (view === "consumer") {
                        ctx.items = jsonData.Dependencies;
                        ctx.filter_name = "Dependency";

                        userId = jsonData.UserId;
                        dependencyId = filter === "all" ? "" : filter
                    } else if (view === "globalops"){
                        // ctx.items = jsonData.Users;
                        // ctx.filter_name = "User";
                        ctx.items = jsonData.AllDependencies;
                        ctx.filter_name = "Dependency";

                        // userId = filter === "all" ? "" : filter
                        dependencyId = filter === "all" ? "" : filter
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
});

// JS for /home
const CompHome = Vue.component('home', {
    template: '#template-home',
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
            }, function (err) {
                console.log("/api/summary-view not loaded with err:");
                console.log(err);
            });

            loadJSON("/api/endpoints", function (jsonData) {
                ctx.all_endpoints = jsonData.Endpoints;
            }, function (err) {
                console.log("/api/endpoints not loaded with err:");
                console.log(err);
            })
        }
    }
});

// Main Vue JS
const routes = [
    { name: 'home', path: '/home', title: "Home", component: CompHome },
    { name: 'details', path: '/details/:view/:filter', title: 'Policy Explorer', component: CompDetails },
    { name: 'timeline', path: '/timeline', title: 'Audit Log', component: CompTimeline },
    { path: '*', redirect: { name: 'home' } }
];

const router = new VueRouter({
    routes: routes,
    linkActiveClass: 'active',
    linkExactActiveClass: 'active'
});

const app = new Vue({
    el: '#wrapper',
    router: router,
    data: {
        routes: routes,
        obj_view: []
    },
    computed: {
        user: function () {
            return {
                id:  Cookies.get("logUserId"),
                name: Cookies.get("logUserName"),
                descr:  Cookies.get("logUserDescr")
            }
        }
    }
});
