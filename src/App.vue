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
      // Store session in session storage with expiration time
      const expirationTime = new Date().getTime() + 600000; // 10 minutes in milliseconds
      sessionStorage.setItem('session', JSON.stringify({ loggedIn: true, expiresAt: expirationTime }));
    },
    checkSessionExpiration() {
      const session = JSON.parse(sessionStorage.getItem('session'));
      if (session && session.loggedIn && session.expiresAt > new Date().getTime()) {
        this.loggedIn = true;
      } else {
        this.loggedIn = false;
        sessionStorage.removeItem('session');
      }
    },
  },
  mounted() {
    this.checkSessionExpiration();
    // Check session expiration periodically
    setInterval(this.checkSessionExpiration, 60000); // Check every minute
  },
};
</script>
