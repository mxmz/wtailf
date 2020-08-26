<template>
  <v-container>
    <v-text-field v-model="filter" hint="Filter" placeholder="Filter list" dense></v-text-field>
    <v-list dense>
      <template v-for="a in filteredItems">
        <v-list-item two-lines  v-bind:key="a.id" >
        <v-list-item-content>
          <v-list-item-title>
            <v-icon x-small>mdi-network-outline</v-icon>
          <a :href='a.endpoint'>

             {{a.id}}
          </a>
          <v-divider class="mx-2" vertical inset></v-divider>
          <a fab x-small :href='a.endpoint' :target="'_blank'+a.endpoint"><v-icon x-small>mdi-open-in-new</v-icon></a>
          </v-list-item-title>
          {{a.hostname}}
          <v-divider class="mx-2" vertical inset></v-divider>
           {{a.endpoint}}
           <v-divider class="mx-2" vertical inset></v-divider>
           {{a.when | formatDate}}

        </v-list-item-content>

      </v-list-item>
      <v-divider class="mx-2" inset v-bind:key="a.hostname" ></v-divider>

      </template>
      <v-divider class="mx-2" inset></v-divider>
    </v-list>
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
      .catch((error) => {
        console.error(error)
        setTimeout(() => {
          window.document.location.reload(true)
        }, 3000)
      })
  }
}
</script>
