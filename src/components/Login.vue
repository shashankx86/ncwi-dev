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
      try {
        const response = await fetch('http://localhost:5499/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username: this.username, password: this.password })
        });
        const result = await response.json();
        if (response.ok) {
          localStorage.setItem('session', result.sessionId);
          this.$emit('logged-in');
        } else {
          this.error = result.message;
        }
      } catch (error) {
        this.error = 'An error occurred. Please try again.';
      }
    }
  }
};
</script>

<style scoped>
.login-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
}

.login-container h1 {
  margin-bottom: 1rem;
}

.login-container form {
  display: flex;
  flex-direction: column;
}

.login-container input {
  margin-bottom: 0.5rem;
  padding: 0.5rem;
  font-size: 1rem;
}

.login-container button {
  padding: 0.5rem;
  font-size: 1rem;
  cursor: pointer;
}

.login-container p {
  color: red;
}
</style>
