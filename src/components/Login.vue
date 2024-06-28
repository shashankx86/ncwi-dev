<template>
  <div class="login-container">
    <h1>Login</h1>
    <form @submit.prevent="login">
      <input v-model="username" type="text" placeholder="Username" required />
      <input v-model="password" type="password" placeholder="Password" required />
      <button type="submit">Login</button>
    </form>
    <p v-if="error">{{ error }}</p>
  </div>
</template>

<script>
export default {
  data() {
    return {
      username: '',
      password: '',
      error: ''
    };
  },
  methods: {
    async login() {
      const apiUrl = `https://napi.${this.username}.hackclub.app/login`;
      try {
        const response = await fetch(apiUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username: this.username, password: this.password })
        });
        const result = await response.json();
        if (response.ok) {
          // Save session ID to localStorage
          localStorage.setItem('sessionId', result.sessionId);

          // Optionally, store the username if needed
          localStorage.setItem('username', this.username);

          // Emit the logged-in event
          this.$emit('logged-in');

          // Redirect to the main application page or dashboard
          this.$router.push({ name: 'dashboard' });
        } else {
          this.error = result.message || 'Invalid username or password. Please try again.';
        }
      } catch (error) {
        this.error = 'An error occurred. Please try again.';
      }
    }
  },
  beforeDestroy() {
    // Remove session data when the component is destroyed
    localStorage.removeItem('sessionId');
    localStorage.removeItem('username');
  },
  beforeRouteLeave(to, from, next) {
    // Clean up session data on page unload
    localStorage.removeItem('sessionId');
    localStorage.removeItem('username');
    next();
  }
};
</script>
