<template>
<v-container>
  <v-row>
    <div v-for="a in items" v-bind:key="a">
      <router-link :to="{ path: 'tail', query: { source: a }}"><v-btn>{{a}}</v-btn></router-link>
    </div>

  </v-row>
</v-container>
</template>
<script lang="ts">
import Vue from 'vue'
import { Component, Prop, Watch } from 'vue-property-decorator'

@Component
export default class List extends Vue {
  private items: string[] = []
  constructor () {
    super()
    console.log('List: ctor')
  }

  mounted () {
    console.log('List: mounted')
    fetch('/sources').then(stream => stream.json())
      .then((data: string[]) => {
        this.items = data
        console.log(data)
      })
      .catch(error => console.error(error))
  }
}
</script>
