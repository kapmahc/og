import Vue from 'vue'
import Vuex from 'vuex'
import vuexI18n from 'vuex-i18n'

import enUS from '@/locales/en-US'
import zhHans from '@/locales/zh-Hans'
import zhHant from '@/locales/zh-Hant'
import {LOCALE} from '@/constants'

Vue.use(Vuex)
const store = new Vuex.Store()

Vue.use(vuexI18n.plugin, store)

Vue.i18n.add('en-US', enUS)
Vue.i18n.add('zh-Hans', zhHans)
Vue.i18n.add('zh-Hant', zhHant)

Vue.i18n.set(localStorage.getItem(LOCALE) || 'en-US')

export default store
