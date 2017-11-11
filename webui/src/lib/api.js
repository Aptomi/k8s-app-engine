import 'js-yaml'

const basePath = 'http://127.0.0.1:27866/api/v1/'

export function callAPI (handler, successFunc, errorFunc) {
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
