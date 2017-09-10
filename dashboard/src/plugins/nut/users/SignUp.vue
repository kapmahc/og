<template>
  <application-layout>
    <group :title="$t('nut.users.sign-up.title')">
      <x-input :title="$t('attributes.username')" v-model="name" :placeholder="$t('placeholders.username')" :min="2" :max="32" />
      <x-input :title="$t('attributes.email')" v-model="email" :placeholder="$t('placeholders.email')" is-type="email" />
      <x-input :title="$t('attributes.password')" type="password" :placeholder="$t('placeholders.password')" v-model="password" :min="6" :max="32" />
      <x-input :title="$t('attributes.passwordConfirmation')" v-model="passwordConfirmation" type="password" :placeholder="$t('placeholders.passwordConfirmation')" :equal-with="password" />
      <x-button
        type="primary"
        @click.native="onSubmit"
        :text="$t('buttons.submit')"
        />
    </group>
  </application-layout>
</template>

<script>
import { XInput, Group, XButton, Cell } from 'vux'

export default {
  name: 'nut-users-sign-up',
  components: {
    XInput,
    XButton,
    Group,
    Cell
  },
  data () {
    return {
      name: '',
      email: '',
      password: '',
      passwordConfirmation: ''
    }
  },
  methods: {
    onSubmit () {
      console.log('sign up')
    }
  }
}
</script>
