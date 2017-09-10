<template>
  <div style="width:100%;position:absolute;left:0;top:0;z-index:100;">
    <x-header
      :right-options="{showMore: true}"
      @on-click-more="showMenus = true">
      with more menu
    </x-header>
    <div v-transfer-dom>
      <actionsheet
        :menus="languages.map((l)=>$t(`languages.${l}`))"
        @on-click-menu="switchLanguage"
        v-model="showMenus"
        show-cancel />
    </div>
  </div>
</template>

<script>
import { cookie, XHeader, Actionsheet, TransferDom, ButtonTab, ButtonTabItem } from 'vux'

import {LOCALE} from '@/constants'

export default {
  name: 'top-header',
  directives: {
    TransferDom
  },
  components: {
    XHeader,
    Actionsheet,
    ButtonTab,
    ButtonTabItem
  },
  data () {
    return {
      languages: ['en-US', 'zh-Hans', 'zh-Hant'],
      showMenus: false
    }
  },
  methods: {
    switchLanguage (i) {
      if (i < this.languages.length) {
        var lang = this.languages[i]
        localStorage.setItem(LOCALE, lang)
        cookie.set(LOCALE, lang, {path: '/', expires: 1024})
        this.$i18n.set(lang)
      }
    }
  }
}
</script>
