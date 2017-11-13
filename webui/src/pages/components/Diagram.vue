<template>
  <div>

    <div class="box box-default">
      <div class="overlay" v-if="loading">
        <i class="fa fa-refresh fa-spin"></i>
      </div>
      <div class="box-body">
        <div >
          <div class="col-xs-12">

            <div id="details_graph"></div>

          </div>
        </div>
        <!-- /.row -->
      </div>
    </div>

  </div>
</template>
<script>
  import { getPolicyDiagram, getPolicyDiagramCompare } from 'lib/api.js'
  import vis from 'vis'

  export default {
    data () {
      // empty data
      return {
        loading: false,
        error: null
      }
    },
    created () {
      // fetch the data when the view is created and the data is already being observed
      this.fetchData()
    },
    props: {
      'policyGen': {
        type: String
      },
      'policyGenBase': {
        type: String
      }
    },
    watch: {
      'policyGen': 'fetchData',
      'policyGenBase': 'fetchData'
    },
    methods: {
      fetchData () {
        this.loading = true
        this.error = null

        const fetchSuccess = $.proxy(function (data) {
          this.loading = false
          let container = document.getElementById('details_graph')
          let network = new vis.Network(container, data, options)
          // network.on("click", clickedNode)
          network.fit()
        }, this)

        const fetchError = $.proxy(function (err) {
          this.loading = false
          this.error = err
        }, this)

        if (this.policyGenBase) {
          getPolicyDiagramCompare(this.policyGen, this.policyGenBase, fetchSuccess, fetchError)
        } else {
          getPolicyDiagram(this.policyGen, fetchSuccess, fetchError)
        }
      }
    }
  }

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
      component: {
        font: {
          color: 'black',
          multi: 'html'
        },
        color: {background: 'rgb(250,250,80)', border: 'darkslategrey'},
        shape: 'box'
      },
      contract: {
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
      dependency: {
        shape: 'icon',
        icon: {
          face: 'FontAwesome',
          code: '\uf007',
          size: 50,
          color: 'orange'
        }
      },
      dependencyNotResolved: {
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
      error: {
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
        direction: 'LR',
        levelSeparation: 220
      }
    },
    interaction: {
      hover: true,
      navigationButtons: true,
      keyboard: true
    },
    physics: false
  }

  function chosenNode (values, id, selected, hovering) {
    values.color = '#ffdd88'
    values.borderColor = '#ff0000'
  }
/*
  function clickedNode (params) {
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
      // loadJSON("/api/object-view?id=" + id, objectLoaded, objectNotLoaded);
      // $("#rule-log-button").click();
    } else {
      // app.obj_view = [];
    }
  }
*/
</script>
<style>
  #details_graph {
    color: #d3d3d3;
    background-color: #222222;
    border: 1px solid #444444;
    font: 12pt arial;
    width: 100%;
    height: 800px;
  }
</style>
