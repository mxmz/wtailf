<template>
  <div class="home">
    <hr>
    <List /> </div>
</template>

<script>
// @ is an alias to /src
import List from '@/components/Peers.vue'

export default {
  name: 'PeersPage',
  components: {
    List
  }
}
</script>
