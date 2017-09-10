<template>
  <application-layout>
    <group>
      <x-input title="邮箱" name="email" placeholder="请输入邮箱地址" is-type="email"></x-input>
    </group>
  </application-layout>
</template>

<script>
import { XInput, Group, XButton, Cell } from 'vux'

export default {
  name: 'nut-users-sign-in',
  components: {
    XInput,
    XButton,
    Group,
    Cell
  },
  data () {
    return {}
  }
}
</script>
