<template>
  <tabbar style="position: fixed">
    <tabbar-item selected link="/component/demo">
      <img slot="icon" src="../assets/icons/home.png">
    </tabbar-item>
    <tabbar-item selected link="/component/demo">
      <img slot="icon" src="../assets/icons/donate.png">
    </tabbar-item>
    <tabbar-item selected link="/component/demo">
      <img slot="icon" src="../assets/icons/settings.png">
    </tabbar-item>
  </tabbar>
</template>

<script>
import { Tabbar, TabbarItem, Icon } from 'vux'

export default {
  name: 'bottom-footer',
  components: {
    Tabbar,
    TabbarItem,
    Icon
  },
  data () {
    return {}
  }
}
</script>
