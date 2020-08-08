<template>
  <v-app>
    <v-app-bar app color="primary" class>
      <v-row cols="12">
        <v-col cols="10">
          <router-link to="/list">
            <v-btn small fab absolute>
              <v-icon>mdi-format-list-bulleted</v-icon>
            </v-btn>
          </router-link>
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
    //
  })
})
</script>
