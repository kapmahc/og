<template>
  <div>
    <top-header />
    <slot/>
    <bottom-footer />
  </div>
</template>

<script>
import Header from '@/components/Header'
import Footer from '@/components/Footer'

export default {
  name: 'application-layout',
  components: {
    'top-header': Header,
    'bottom-footer': Footer
  },
  data () {
    return {
    }
  }
}
</script>
