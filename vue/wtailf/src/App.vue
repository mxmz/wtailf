<template>
  <v-app>
    <v-app-bar app color="primary" class>
      <v-row cols="12">
        <v-col cols="10">

            <v-btn small fab to="/list" :disabled="$route.path === '/list' || $route.path === '/'">
              <v-icon>mdi-format-list-checkbox</v-icon>
            </v-btn>

          <v-divider class="mx-4" vertical inset></v-divider>

            <v-btn small fab to="/peerlist" :disabled="$route.path === '/peerlist'">
              <v-icon>mdi-lan</v-icon>
            </v-btn>

          <v-divider class="mx-4" vertical inset></v-divider>
          <b style="font-family: monospace">{{title}}</b>

        </v-col>

        <v-col cols="2">
          <v-btn href="https://github.com/mxmz/wtailf" target="_blank" text>
            <span class="mr-2">wtailf</span>
            <v-icon>mdi-github</v-icon>
          </v-btn>
        </v-col>
      </v-row>
    </v-app-bar>

    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>
<style>
* {
  text-transform: none !important;
}
</style>
<script lang="ts">
import Vue from 'vue'

export default Vue.extend({
  name: 'App',

  components: {
    //    HelloWorld
  },

  data: () => ({
    title: document.location.hostname + ' : ' + document.location.port
  })
})
</script>
