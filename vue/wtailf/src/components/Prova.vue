<template>
  <div class="hello">
    <h1>{{ msg }}</h1>
    <input v-model="filter">
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
import { Component, Prop } from 'vue-property-decorator'

@Component
export default class Prova extends Vue {
  @Prop() private msg!: string;
  private source: EventSource | null = null;
  private idx = 0;
  private list: {idx: number; message: string}[] = [];
  private filter = ''

  constructor () {
    super()
    console.log('ctor')
  }

  get filteredList (): object[] {
    return this.list.filter(x => x.message.indexOf(this.filter) !== -1)
  }

  mounted () {
    console.log('mounted')
    this.source = new EventSource('/events')
    this.source.addEventListener('log', (_event: Event) => {
      const event = _event as MessageEvent
      // console.log(event)
      this.idx++
      this.list.unshift({
        idx: this.idx,
        message: event.data
      })
      if (this.list.length > 500) {
        this.list = this.list.splice(0, 200)
      }
    })
  }

  destroyed () {
    console.log('destroyed')
    if (this.source) { this.source.close() }
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
