<template>
  <div>
    {{prefix}}<img style="height: 20px; margin-right: 5px" :src="imagePath" :title="kindComputed"/>{{caption}}
  </div>
</template>

<script>
export default {
  props: ['obj', 'kind', 'prefix'],
  computed: {
    kindComputed () {
      if (this.kind != null) {
        return this.kind
      }
      return this.obj['kind']
    },
    image () {
      const o = this.obj
      const kind = this.kindComputed
      switch (kind) {
        case 'cluster':
          if (o['type'].indexOf('kubernetes') >= 0) return 'k8s'
          break
        case 'aclrule':
          return 'rule'
        case 'code':
          if (o['code']['type'].indexOf('helm') >= 0) return 'helm'
          if (o['code']['type'].indexOf('raw') >= 0) return 'k8s'
          break
      }
      return kind
    },
    caption () {
      const o = this.obj
      const kind = this.kindComputed
      if (kind === 'code') {
        if (o['code']['type'].indexOf('helm') >= 0) {
          var chartName = o['code']['params']['chartName']
          var chartVersion = o['code']['params']['chartVersion']
          if (chartVersion == null) {
            chartVersion = 'latest'
          }
          return chartName + ' / ' + chartVersion
        }
        if (o['code']['type'].indexOf('raw') >= 0) {
          return o['name']
        }
      }

      return this.obj['name']
    },
    imagePath () {
      return '/static/img/' + this.image + '-icon.png'
    }
  }
}
</script>
