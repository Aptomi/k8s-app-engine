const async = true
const sync = false

const yaml = require('js-yaml')
const delayMs = 0
const basePath = process.env.API_BASEPATH

/*
 * Exported functions, which can be used in pages/components
 */

// returns the list of namespaces, given a map of policy object references (map[namespace][kind][name] -> generation)
export function getNamespacesByRefMap (policyObjectsRefMap) {
  const namespaces = []
  for (const ns in policyObjectsRefMap) {
    namespaces.push(ns)
  }
  return namespaces.sort()
}

// returns map from a namespace to a list of objects, given a list of objects
export function getObjectMapByNamespace (objectList) {
  const result = {}
  for (const idx in objectList) {
    const ns = objectList[idx]['namespace']
    if (!(ns in result)) {
      result[ns] = []
    }
    result[ns].push(objectList[idx])
  }
  return result
}

// filters the list of objects by a given namespace and/or kind
export function filterObjects (policyObjects, nsFilter = null, kindFilter = null) {
  const result = []
  for (const ns in policyObjects) {
    if (nsFilter === null || ns === nsFilter) {
      for (const kind in policyObjects[ns]) {
        if (kindFilter === null || kind === kindFilter) {
          for (const name in policyObjects[ns][kind]) {
            const entry = {
              'namespace': ns,
              'kind': kind,
              'name': name,
              'generation': policyObjects[ns][kind][name]
            }
            result.push(entry)
          }
        }
      }
    }
  }
  return result
}

// loads the object diagram
export async function getObjectDiagram (obj, successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy', 'diagram', 'object', obj['namespace'], obj['kind'], obj['name']].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data['data'])
  }, function (err) {
    errorFunc(err)
  })
}

// loads the policy diagram
export async function getPolicyDiagram (mode, generation, successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy', 'diagram', 'mode', mode, 'gen', generation].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data['data'])
  }, function (err) {
    errorFunc(err)
  })
}

// loads the policy diagram, comparing two policies
export async function getPolicyDiagramCompare (mode, generation, generationBase, successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy', 'diagram', 'compare', 'mode', mode, 'gen', generation, 'genBase', generationBase].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data['data'])
  }, function (err) {
    errorFunc(err)
  })
}

// loads all users and their roles
export async function getUsersAndRoles (successFunc, errorFunc) {
  await makeDelay()
  const handler = ['user', 'roles'].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data['data'])
  }, function (err) {
    errorFunc(err)
  })
}

// authenticates the user
export async function authenticateUser (username, password, successFunc, errorFunc) {
  const handler = ['user', 'login'].join('/')
  var authReq = {
    'kind': 'auth-request',
    'username': username,
    'password': password
  }
  callAPI(handler, async, function (data) {
    successFunc(data)
  }, function (err) {
    errorFunc(err)
  }, authReq)
}

// loads the latest policy
export async function getPolicy (successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy'].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data)
  }, function (err) {
    errorFunc(err)
  })
}

// updates/saves objects in the latest policy
export async function savePolicyObjects (successFunc, errorFunc, policyObjects) {
  await makeDelay()
  const handler = ['policy'].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data)
  }, function (err) {
    errorFunc(err)
  }, policyObjects)
}

// deletes objects in the latest policy
export async function deletePolicyObjects (successFunc, errorFunc, policyObjects) {
  await makeDelay()
  const handler = ['policy'].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data)
  }, function (err) {
    errorFunc(err)
  }, policyObjects, true)
}

// loads all policies and returns revision information for each and every of them
export async function getAllPolicies (successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy'].join('/')
  callAPI(handler, async, function (data) {
    // here we retrieved just the latest policy and we know its generation, so let's retrieve everything else
    const policies = []
    let lastGen = getPolicyGeneration(data)
    for (let i = lastGen; i > 0; i--) {
      let p = {}
      fetchPolicy(i.toString(), p)
      fetchPolicyRevisions(i.toString(), p)
      policies.push(p)
    }
    successFunc(policies)
  }, function (err) {
    errorFunc(err)
  })
}

// returns policy generation
export function getPolicyGeneration (policy) {
  return policy['metadata']['generation']
}

// returns map of references to policy objects (map[namespace][kind][name] -> generation), given the loaded policy
export function getPolicyObjectRefMap (policy) {
  return policy['objects']
}

// loads all objects and fetch their properties
export async function getPolicyObjectsWithProperties (successFunc, errorFunc, kindFilter = null) {
  await makeDelay()
  const handler = ['policy'].join('/')
  callAPI(handler, async, function (data) {
    // 1. once all objects are received, filter them by kind
    const objectList = filterObjects(data['objects'], null, kindFilter)

    // 2. fetch properties for every object
    for (const idx in objectList) {
      fetchObjectProperties(objectList[idx])
    }

    // 3. fetch status for all dependencies
    fetchDependenciesStatus(objectList)

    successFunc(objectList)
  }, function (err) {
    errorFunc(err)
  })
}

// fetches status of all dependencies
export function fetchDependenciesStatus (objectList) {
  const depIds = []
  for (const idx in objectList) {
    if (objectList[idx]['kind'] === 'dependency') {
      depIds.push(objectList[idx]['namespace'] + '^' + objectList[idx]['name'])
    }
  }

  const handler = ['policy', 'dependency', 'status', 'deployed', depIds.join(',')].join('/')
  callAPI(handler, sync, function (data) {
    for (const idx in objectList) {
      if (objectList[idx]['kind'] === 'dependency') {
        const key = [objectList[idx]['namespace'], objectList[idx]['kind'], objectList[idx]['name']].join('/')
        if (key in data['status'] && data['status'][key]['found']) {
          objectList[idx]['status'] = data['status'][key]
        } else {
          objectList[idx]['status_error'] = 'unable to retrieve status'
        }
      }
    }
  }, function (err) {
    for (const idx in objectList) {
      if (objectList[idx]['kind'] === 'dependency') {
        objectList[idx]['status_error'] = err
      }
    }
  })
}

// loads dependency deployment status
export async function getResources (d, successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy', 'dependency', d['metadata']['namespace'], d['metadata']['name'], 'resources'].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data['resources'])
  }, function (err) {
    errorFunc(err)
  })
}

// loads revision event logs
export async function getEventLogs (r, successFunc, errorFunc) {
  await makeDelay()
  // temporarily just return back revision itself as it now contains both logs in it
  successFunc(r)
}

/*
 * Helpers to synchronously fetch missing data. Those are required because our API cannot return data in
 * a single call. So we have to make multiple calls.
 */

// receives a bare entry with populated fields (namespace, kind, name, generation), loads the corresponding object
// from the database and populates the corresponding fields in obj
export function fetchObjectProperties (obj, successFunc = null, errorFunc = null) {
  const handler = ['policy', 'gen', '0', 'object', obj['namespace'], obj['kind'], obj['name']].join('/')
  callAPI(handler, sync, function (data) {
    // copy over retrieved object properties
    for (const key in data) {
      obj[key] = data[key]
    }
    // remove generation prior to getting YAML representation
    delete data['metadata']['generation']
    obj['yaml'] = yaml.safeDump(data, {'lineWidth': 160})
    if (successFunc != null) {
      successFunc(obj)
    }
  }, function (err) {
    obj['error'] = 'unable to fetch object properties: ' + err
    if (errorFunc != null) {
      errorFunc(err)
    }
  })
}

// fetches policy revisions for a given policy
export function fetchPolicyRevisions (generation, p) {
  const handler = ['revisions', 'policy', generation].join('/')
  callAPI(handler, sync, function (data) {
    p['revisions'] = data['data']
  }, function (err) {
    // can't fetch policy revisions
    p['revisions'] = [err]
  })
}

// loads the latest policy
export function fetchPolicy (generation, p) {
  const handler = ['policy', 'gen', generation].join('/')
  callAPI(handler, sync, function (data) {
    for (const idx in data) {
      p[idx] = data[idx]
    }
  }, function (err) {
    p['error'] = err
  })
}

/*
 * Utility/helper functions
 */

// sleeps for a given number of milliseconds
function makeDelay () {
  return new Promise(resolve => setTimeout(resolve, delayMs))
}

// makes an API call to Aptomi
function callAPI (handler, isAsync, successFunc, errorFunc, body = null, deleteFlag = false) {
  const path = basePath + handler
  const xhr = new XMLHttpRequest()
  xhr.onreadystatechange = function () {
    if (xhr.readyState === 4) {
      var msg = path + ': returned ' + xhr.status + ' ' + xhr.statusText
      if (xhr.status === 200) {
        // parse response as YAML
        try {
          const data = yaml.safeLoad(xhr.responseText)
          successFunc(data)
        } catch (err) {
          msg += '(error while parsing response: ' + err + ')'
          errorFunc(msg)
        }
      } else {
        // let's try to parse out error text, which was returned
        try {
          const data = yaml.safeLoad(xhr.responseText)
          if (data['kind'] === 'error') {
            msg = 'error: ' + data['error']
          } else {
            msg += '(' + JSON.stringify(data) + ')'
          }
        } catch (err) {
          msg += '(error while parsing response: ' + err + ')'
        }

        // return error
        errorFunc(msg)
      }
    }
  }
  if (body == null) {
    xhr.open('GET', path, isAsync)
    xhr.setRequestHeader('Authorization', 'Bearer ' + localStorage.token)
    xhr.setRequestHeader('Content-type', 'application/yaml')
    xhr.send()
  } else {
    if (deleteFlag) {
      xhr.open('DELETE', path, isAsync)
    } else {
      xhr.open('POST', path, isAsync)
    }
    xhr.setRequestHeader('Authorization', 'Bearer ' + localStorage.token)
    xhr.setRequestHeader('Content-type', 'application/yaml')
    xhr.send(yaml.safeDump(body))
  }
}
