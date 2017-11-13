const async = true
const sync = false

const yaml = require('js-yaml')
const delayMs = 1000
const basePath = process.env.API_BASEPATH

/*
 * Exported functions, which can be used in pages/components
 */

// returns the list of namespaces, given all policy objects
export function getNamespaces (policyObjects) {
  const namespaces = []
  for (const ns in policyObjects) {
    namespaces.push(ns)
  }
  return namespaces
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

// loads the policy diagram
export async function getPolicyDiagram (generation, successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy', 'diagram', 'gen', generation].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data['data'])
  }, function (err) {
    errorFunc(err)
  })
}

// loads the policy diagram, comparing two policies
export async function getPolicyDiagramCompare (generation, generationBase, successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy', 'diagram', 'compare', 'gen', generation, 'genBase', generationBase].join('/')
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
  const handler = ['user', 'authenticate'].join('/')
  let formData = new FormData()
  formData.append('username', username)
  formData.append('password', password)
  callAPI(handler, async, function (data) {
    successFunc(data)
  }, function (err) {
    errorFunc(err)
  }, formData)
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

// loads all policies and returns revision information for each and every of them
export async function getAllPolicies (successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy'].join('/')
  callAPI(handler, async, function (data) {
    // here we retrieved just the latest policy and we know its generation, so let's retrieve everything else
    const policies = []
    let lastGen = getPolicyGeneration(data)
    for (let i = 1; i <= lastGen; i++) {
      // policy = getPolicy
      policies.push(data)
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

// returns all policy objects (map[namespace][kind][name] -> generation), given the loaded policy
export function getPolicyObjects (policy) {
  return policy['objects']
}

// loads all dependencies
export async function getDependencies (successFunc, errorFunc) {
  await makeDelay()
  const handler = ['policy'].join('/')
  callAPI(handler, async, function (data) {
    const dependencies = filterObjects(data['objects'], null, 'dependency')
    for (const idx in dependencies) {
      fetchObjectProperties(dependencies[idx])
      fetchDependency(dependencies[idx])
    }
    successFunc(dependencies)
  }, function (err) {
    errorFunc(err)
  })
}

// loads all endpoints
export async function getEndpoints (successFunc, errorFunc) {
  await makeDelay()
  const handler = ['endpoints'].join('/')
  callAPI(handler, async, function (data) {
    successFunc(data['endpoints'])
  }, function (err) {
    errorFunc(err)
  })
}

/*
 * Helpers to synchronously fetch missing data. Those are required because our API cannot return data in
 * a single call. So we have to make multiple calls.
 */

// receives a bare entry with populated fields (namespace, kind, name, generation), loads the corresponding object
// from the database and populates the corresponding fields in obj
export function fetchObjectProperties (obj, successFunc = null, errorFunc = null) {
  const handler = ['policy', 'gen', obj['generation'], 'object', obj['namespace'], obj['kind'], obj['name']].join('/')
  callAPI(handler, sync, function (data) {
    for (const key in data) {
      obj[key] = data[key]
    }
    obj['yaml'] = yaml.safeDump(data)
    if (successFunc != null) {
      successFunc(obj)
    }
  }, function (err) {
    // can't fetch object properties
    obj['error'] = 'unable to fetch object properties: ' + err
    if (errorFunc != null) {
      errorFunc(err)
    }
  })
}

// fetches data for a single dependency
function fetchDependency (d) {
  const handler = ['dependency_status'].join('/')
  callAPI(handler, sync, function (data) {
    d['status'] = 'Deployed'
  }, function (err) {
    // can't fetch dependency properties
    d['status_error'] = 'unable to fetch dependency status: ' + err
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
function callAPI (handler, isAsync, successFunc, errorFunc, formData = null) {
  const path = basePath + handler
  const xhr = new XMLHttpRequest()
  xhr.onreadystatechange = function () {
    if (xhr.readyState === 4) {
      if (xhr.status === 200) {
        try {
          successFunc(yaml.safeLoad(xhr.responseText))
        } catch (err) {
          errorFunc('exception occurred: ' + err)
        }
      } else {
        if (xhr.statusText) {
          errorFunc(xhr.status + ' ' + xhr.statusText)
        } else {
          errorFunc('unable to load data from ' + path)
        }
      }
    }
  }
  if (formData == null) {
    xhr.open('GET', path, isAsync)
    xhr.send()
  } else {
    xhr.open('POST', path, isAsync)
    xhr.send(formData)
  }
}

