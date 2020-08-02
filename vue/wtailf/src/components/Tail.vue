<template>
  <div class="hello">
    <v-progress-linear v-if="running"
      indeterminate
      color="yellow darken-2"
    ></v-progress-linear>
    <!-- input v-model="filter" -->
    <v-text-field v-model="filter" :rules="rules" hint="Filter" ></v-text-field>
    <div>
      <div v-for="a in filteredList" v-bind:key="a.idx">
        <div>
          [{{a.idx}}] {{a.message}}
          <hr>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component, Prop, Watch } from 'vue-property-decorator'
import { Route } from 'vue-router'

@Component
export default class Tail extends Vue {
  @Prop()
  private msg!: string;

  private source: EventSource | null = null;
  private idx = 0;
  private list: {idx: number; message: string}[] = [];
  private filter = ''
  private rules = [
    (value: string) => !!value || 'Required.',
    (value: string) => (value || '').length <= 30 || 'Max 20 characters'
  ]

  private running = 0

  @Watch('$route', { immediate: true, deep: true })
  onUrlChange (newVal: Route) {
    // Some action
    const q = newVal.query
    console.log(q)
    this.unsubscribe()
    this.substribe((q.source || '') as string)
  }

  constructor () {
    super()
    console.log('Tail: ctor')
  }

  get filteredList (): object[] {
    return this.list.filter(x => x.message.indexOf(this.filter) !== -1)
  }

  mounted () {
    console.log('mounted')
  }

  substribe (subject: string) {
    this.running++
    this.source = new EventSource('/events?source=' + encodeURIComponent(subject))
    this.source.addEventListener('log', (_event: Event) => {
      const event = _event as MessageEvent
      // console.log(event)
      if (this.idx === 0) {
        this.running--
      }
      this.idx++
      this.list.unshift({
        idx: this.idx,
        message: event.data
      })
      if (this.list.length > 1200) {
        this.list = this.list.splice(0, 1000)
      }
    })
  }

  unsubscribe () {
    if (this.source) { this.source.close() }
    this.source = null
  }

  destroyed () {
    console.log('destroyed')
    this.unsubscribe()
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
