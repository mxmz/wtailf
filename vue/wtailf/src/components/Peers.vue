<template>
  <v-container>
    <v-text-field v-model="filter" hint="Filter" outlined></v-text-field>
    <v-col>
      <v-row v-for="a in filteredItems" v-bind:key="a.hostname">

          <v-btn :href='a.endpoint'>
            <v-icon left>mdi-link</v-icon>
            {{a.hostname}} ({{a.id}})
          </v-btn>

      </v-row>
    </v-col>
  </v-container>
</template>
<script lang="ts">
import Vue from 'vue'
import { Component, Prop, Watch } from 'vue-property-decorator'

interface Peer {
  id: string;
  endpoint: string;
  hostname: string;
  when: string;
}

@Component
export default class Peers extends Vue {
  private items: Peer[] = [];
  private filter = '';
  constructor () {
    super()
    console.log('List: ctor')
  }

  get filteredItems (): Peer[] {
    return this.items.filter((x) => x.hostname.indexOf(this.filter) !== -1 || x.endpoint.indexOf(this.filter) !== -1)
  }

  mounted () {
    console.log('List: mounted')
    fetch('/peers')
      .then((stream) => stream.json())
      .then((data: Peer[]) => {
        this.items = data.sort((a, b) => a.hostname.localeCompare(b.hostname))
        console.log(data)
      })
      .catch((error) => console.error(error))
  }
}
</script>
