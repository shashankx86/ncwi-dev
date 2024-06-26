<template>
  <div id="app">
    <Navbar v-if="loggedIn" @open-terminal="showTerminal = !showTerminal" />
    <div class="content">
      <Terminal v-if="loggedIn && showTerminal" />
      <Login v-if="!loggedIn" @logged-in="handleLogin" />
    </div>
  </div>
</template>

<script>
import Navbar from './components/Navbar.vue';
import Terminal from './components/Terminal.vue';
import Login from './components/Login.vue';
import './style.css';

export default {
  name: 'App',
  components: {
    Navbar,
    Terminal,
    Login,
  },
  data() {
    return {
      loggedIn: false,
      showTerminal: false,
    };
  },
  methods: {
    handleLogin() {
      this.loggedIn = true;
      // Store session in session storage
      sessionStorage.setItem('session', 'true');
    },
  },
  mounted() {
    // Check session storage for logged-in state
    const session = sessionStorage.getItem('session');
    if (session) {
      this.loggedIn = true;
    }
  },
};
</script>
