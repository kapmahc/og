<template>
  <group>
    <cell-box is-link :link="{name: 'users.sign-in'}">
      {{$t('nut.users.sign-in.title')}}
    </cell-box>
    <cell-box is-link :link="{name: 'users.sign-up'}">
      {{$t('nut.users.sign-up.title')}}
    </cell-box>
  </group>

</template>

<script>
import {CellBox, Group} from 'vux'

export default {
  name: 'nut-users-shared-links',
  components: {
    CellBox,
    Group
  },
  data () {
    return {

    }
  }
}
</script>
