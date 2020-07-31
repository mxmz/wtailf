<template>
  <v-app>
    <v-app-bar
      app
      color="primary"
      dark
    >
      <div class="d-flex align-center">

        </div>

      <v-spacer></v-spacer>

      <v-btn
        href="https://github.com/mxmz/wtailf"
        target="_blank"
        text
      >
        <span class="mr-2">wtailf</span>
        <v-icon>mdi-</v-icon>
      </v-btn>
    </v-app-bar>

    <v-main>
      <router-view/>

    </v-main>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue'

export default Vue.extend({
  name: 'App',

  components: {
    //    HelloWorld
  },

  data: () => ({
    //
  })
})
</script>
