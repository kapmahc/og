<template>
  <div>
    <group :title="$t('nut.users.sign-up.title')">
      <x-input
        :title="$t('attributes.email')"
        v-model="form.email"
        is-type="email" />
      <x-input
        :title="$t('attributes.password')"
        type="password"
        v-model="form.password"
        :min="6" />
      <x-button
        type="primary"
        @click.native="onSubmit"
        :text="$t('buttons.submit')"
        />
    </group>
    <users-shared-links />
  </div>
</template>

<script>
import { XInput, Group, XButton, Cell } from 'vux'

import SharedLinks from './SharedLinks'

export default {
  name: 'nut-users-sign-in',
  components: {
    XInput,
    XButton,
    Group,
    Cell,
    'users-shared-links': SharedLinks
  },
  data () {
    return {
      form: {
        email: '',
        password: ''
      }
    }
  },
  methods: {
    onSubmit () {

    }
  }
}
</script>
