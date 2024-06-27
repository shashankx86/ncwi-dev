// src/components/Login.vue
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
          localStorage.setItem('session', result.sessionId);
          this.$emit('logged-in');
        } else {
          this.error = result.message || 'An error occurred. Please try again.';
        }
      } catch (error) {
        this.error = 'An error occurred. Please try again.';
      }
    }
  }
};
</script>
