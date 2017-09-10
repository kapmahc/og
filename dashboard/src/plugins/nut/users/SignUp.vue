<template>
  <div>
    <group :title="$t('nut.users.sign-up.title')">
      <x-input
        :title="$t('attributes.username')"
        v-model="form.name"
        :placeholder="$t('placeholders.username')"
        placeholder-align="right"
        :min="2" :max="32" />
      <x-input
        :title="$t('attributes.email')"
        v-model="form.email"
        :placeholder="$t('placeholders.email')"
        placeholder-align="right"
        is-type="email" />
      <x-input
        :title="$t('attributes.password')"
        type="password"
        :placeholder="$t('placeholders.password')"
        placeholder-align="right"
        v-model="form.password"
        :min="6" :max="32" />
      <x-input
        :title="$t('attributes.passwordConfirmation')"
        v-model="form.passwordConfirmation"
        type="password"
        :placeholder="$t('placeholders.passwordConfirmation')"
        placeholder-align="right"
        :equal-with="form.password" />
      <x-button
        type="primary"
        @click.native="onSubmit"
        :text="$t('buttons.submit')"
        />
    </group>
  </div>
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
      form: {
        name: '',
        email: '',
        password: '',
        passwordConfirmation: ''
      }
    }
  },
  methods: {
    onSubmit () {
      console.log(this.form)
    }
  }
}
</script>
