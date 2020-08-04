<template>
  <v-container>
    <v-text-field v-model="filter" hint="Filter" outlined></v-text-field>
    <v-col>
      <v-row v-for="a in filteredItems" v-bind:key="a">
        <router-link :to="{ path: 'tail', query: { source: a }}">
          <v-btn>{{a}}</v-btn>
        </router-link>
      </v-row>
    </v-col>
  </v-container>
</template>
<script lang="ts">
import Vue from 'vue'
import { Component, Prop, Watch } from 'vue-property-decorator'

@Component
export default class List extends Vue {
  private items: string[] = [];
  private filter = '';
  constructor () {
    super()
    console.log('List: ctor')
  }

  get filteredItems (): string[] {
    return this.items.filter((x) => x.indexOf(this.filter) !== -1)
  }

  mounted () {
    console.log('List: mounted')
    fetch('/sources')
      .then((stream) => stream.json())
      .then((data: string[]) => {
        this.items = data.sort()
        console.log(data)
      })
      .catch((error) => console.error(error))
  }
}
</script>
