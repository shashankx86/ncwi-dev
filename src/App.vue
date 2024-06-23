<template>
  <div id="app">
    <Navbar @open-terminal="showTerminal = !showTerminal" />
    <div class="content">
      <h1>NCWI</h1>
      <Terminal v-if="showTerminal" />
    </div>
  </div>
</template>

<script>
import Navbar from './components/Navbar.vue';
import Terminal from './components/Terminal.vue';
import './style.css';

export default {
  name: 'App',
  components: {
    Navbar,
    Terminal,
  },
  data() {
    return {
      showTerminal: false,
    };
  },
};
</script>
