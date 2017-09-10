import Vue from 'vue'

import Application from '@/layouts/Application'
import nut from './nut'

Vue.component('application-layout', Application)

export default {
  routes: [].concat(nut.routes)
}
