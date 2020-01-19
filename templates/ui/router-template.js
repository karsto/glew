import Vue from 'vue'
import Router from 'vue-router'
import qs from 'qs'
import merge from 'lodash.merge'
import isEqual from 'lodash.isequal'

import {
  newDecoder
} from '@/util'

// TODO: check network tab bundles imported twice on pages not sure why maybe https://stackoverflow.com/questions/37081559/all-my-code-runs-twice-when-compiled-by-webpack
const {{.PluralModelName}} = () => import( /* webpackChunkName: "{{.PluralModelName}}" */ '@/components/{{.PluralModelName}}')
const new{{.TitleCaseModelName}} = () => import( /* webpackChunkName: "new{{.TitleCaseModelName}}" */ '@/components/new{{.TitleCaseModelName}}')


Vue.use(Router)

let router = new Router({
  // mode: 'history', // TODO: removing # from url breaks dev proxy /api due to conflict with catch all route in vue router
  base: process.env.BASE_URL,
  parseQuery: (q) =>
    qs.parse(q, {
      decoder: newDecoder()
    }),
  stringifyQuery: (q) => {
    let qp = qs.stringify(q, {
      arrayFormat: 'repeat',
      skipNulls: true
    })
    if (!qp || qp.length <= 1) {
      return ''
    } else return `?${qp}`
  },
  routes: [
    //   {
    //   path: '/login',
    //   name: 'login',
    //   component: login,
    //   meta: {
    //     isPublic: true,
    //     allowedRoles: []
    //   }
    // },
    {
      path: '/{{.ResourceName}}/new',
      name: 'new{{.TitleCaseModelName}}',
      component: new{{.TitleCaseModelName}},
      props: (route) => ({})
    },
    {
      path: '/{{.ResourceName}}/:id',
      name: 'edit{{.TitleCaseModelName}}',
      component: new{{.TitleCaseModelName}},
      props: (route) => ({})
    },
    {
      path: '/{{.ResourceName}}',
      name: '{{.PluralModelName}}',
      component: {{.PluralModelName}},
      meta: {

      }
    },
    {
      // ROOT
      path: '/',
      // redirect: '/'
    },
    {
      // catch all
      path: '*',
      redirect: '/'
    }
  ]
})

router.beforeEach((to, from, next) => {
  let isLoggedIn = true // TODO: fetch
  let isPublic = to.matched.some(record => record.meta.isPublic)

  // auth check
  if (!isPublic && !isLoggedIn) {
    next({
      name: 'login',
      query: {
        fromUrl: to.fullPath
      }
    })
    return
  }

  let activeRole = '' // TODO: fetch
  let isRoleRequired = false // TODO: to.matched.some(record => record.meta.allowedRoles && !record.meta.allowedRoles.includes('*')
  let isRoleAllowed = (role) => {
    to.matched.some(
      record =>
        record.meta.allowedRoles.includes('*') ||
        record.meta.allowedRoles.includes(role)
    )
  }

  let defaultPage = '/' // TODO:
  // role check
  if (isRoleRequired && isRoleAllowed(activeRole)) {
    next(defaultPage)
    return
  }

  // if they nav to login page when they are authenticated
  if (isLoggedIn && (to.name === 'login' || to.name === 'changePassword')) {
    next(defaultPage)
    return
  }

  if (to.meta.defaults) {
    let nRoute = merge({}, to.meta.defaults, to)
    if (!isEqual(nRoute, to)) {
      next(nRoute)
      return
    }
  }

  next()
})

export default router
