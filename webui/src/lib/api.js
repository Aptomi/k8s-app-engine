import Vue from 'vue'

const yaml = require('js-yaml')

const basePath = 'http://127.0.0.1:27866/api/v1/'

/*
 * Exported functions, which can be used in pages/components
 */

// loads all dependencies
export function getDependencies (successFunc, errorFunc) {
  callAPI('policy', function (data) {
    var dependencies = getObjectsByKind(data['objects'], 'dependency')
    for (var idx in dependencies) {
      fetchDependency(dependencies[idx])
    }
    successFunc(dependencies)
  }, function (err) {
    errorFunc(err)
  })
}

/*
 * Utility/helper functions
 */

// makes an API call to Aptomi
function callAPI (handler, successFunc, errorFunc) {
  var path = basePath + handler
  var xhr = new XMLHttpRequest()
  xhr.onreadystatechange = function () {
    if (xhr.readyState === 4) {
      if (xhr.status === 200) {
        successFunc(yaml.safeLoad(xhr.responseText))
      } else {
        if (xhr.statusText) {
          errorFunc(xhr.status + ' ' + xhr.statusText)
        } else {
          errorFunc('unable to load data from ' + path)
        }
      }
    }
  }
  xhr.open('GET', path, true)
  xhr.send()
}

// receives a map[namespace][kind][name] -> generation and returns the list of objects with a given kind
function getObjectsByKind (data, kindFilter) {
  var result = []
  for (var ns in data) {
    for (var kind in data[ns]) {
      if (kind === kindFilter) {
        for (var name in data[ns][kind]) {
          var entry = {
            'namespace': ns,
            'kind': kind,
            'name': name,
            'generation': data[ns][kind][name]
          }
          fetchObjectProperties(entry)
          result.push(entry)
        }
      }
    }
  }
  return result
}

// receives a bare entry with populated fields (namespace, kind, name, generation), loads the corresponding object
// from the database and populates the corresponding fields in obj
function fetchObjectProperties (obj) {
  callAPI('object', function (data) {
    Vue.set(d, 'user', 'user_test')
    Vue.set(d, 'contract', 'contract_test')
  }, function (err) {
    // can't fetch object properties
    Vue.set(d, 'error', 'unable to fetch object properties: ' + err)
  })
}

// fetches data for a single dependency
function fetchDependency (d) {
  callAPI('object', function (data) {
    Vue.set(d, 'status', 'Deployed')
  }, function (err) {
    Vue.set(d, 'status', 'Crap')
    // can't fetch dependency properties
    Vue.set(d, 'error', 'unable to fetch dependency properties: ' + err)
  })
}
