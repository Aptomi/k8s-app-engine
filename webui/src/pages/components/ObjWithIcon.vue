<template>
  <div>
    <img style="float: left; height: 20px; margin-right: 5px" :src="imagePath" title="Kubernetes YAMLs"/>
    <span>{{nameComputed}}</span>
  </div>
</template>

<script>
export default {
  props: ['kind', 'name', 'obj'],
  computed: {
    object () {
      if (this.obj) {
        return this.obj
      }
      return {'kind': this.kind, 'name': this.name}
    },
    nameComputed () {
      return this.object['name']
    },
    image () {
      const o = this.object
      switch (o['kind']) {
        case 'cluster':
          if (o['type'].indexOf('kubernetes') >= 0) return 'k8s'
          break
        case 'aclrule':
          return 'rule'
        default:
          return o['kind']
      }
      return 'unknown'
    },
    imagePath () {
      return '/static/img/' + this.image + '-icon.png'
    }
  }
}
</script>
