<template>
  <v-container>
    <v-text-field v-model="filter" hint="Filter" placeholder="Filter list" dense></v-text-field>
    <v-list dense>
      <template v-for="a in filteredItems">
        <v-list-item two-lines v-bind:key="a">
          <v-list-item-content>
            <v-list-item-title>
              <v-icon small>mdi-file-outline</v-icon>
              <router-link :to="{ path: 'tail', query: { source: a }}">{{a}}</router-link>

              <v-divider class="mx-2" vertical inset></v-divider>
              <router-link x-small :to="{ path: 'tail', query: { source: a }}" :target="'_blank'+a">
                <v-icon x-small>mdi-open-in-new</v-icon>
              </router-link>
            </v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </template>
    </v-list>
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
    fetch('/api/sources')
      .then((stream) => stream.json())
      .then((data: string[]) => {
        this.items = data.sort()
        console.log(data)
      })
      .catch((error) => {
        console.error(error)
        setTimeout(() => {
          window.document.location.replace('/?_=' + encodeURIComponent(new Date().toISOString()))
        }, 3000)
      })
  }
}
</script>
