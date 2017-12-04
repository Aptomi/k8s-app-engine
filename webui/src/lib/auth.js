/* globals localStorage */
import { authenticateUser } from 'lib/api'

export default {
  login (username, password, cb) {
    cb = arguments[arguments.length - 1]
    if (localStorage.token) {
      // eslint-disable-next-line
      if (cb) cb(true)
      this.onChange(true)
      return
    }
    authenticate(username, password, (res) => {
      if (res.authenticated) {
        localStorage.token = res.token
        localStorage.username = username
        // eslint-disable-next-line
        if (cb) cb(true)
        this.onChange(true)
      } else {
        // eslint-disable-next-line
        if (cb) cb(false)
        this.onChange(false)
      }
    })
  },

  getToken () {
    return localStorage.token
  },

  getUsername () {
    return localStorage.username
  },

  logout (cb) {
    delete localStorage.token
    delete localStorage.username
    if (cb) cb()
    this.onChange(false)
  },

  loggedIn () {
    return !!localStorage.token
  },

  onChange () {}
}

function authenticate (username, password, cb) {
  setTimeout(() => {
    const fetchSuccess = $.proxy(function (data) {
      // eslint-disable-next-line
      cb({
        authenticated: true,
        token: Math.random().toString(36).substring(7)
      })
    }, this)

    const fetchError = $.proxy(function () {
      // eslint-disable-next-line
      cb({ authenticated: false })
    }, this)

    authenticateUser(username, password, fetchSuccess, fetchError)
  }, 0)
}
