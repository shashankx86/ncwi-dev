<template>
    <div class="login-container">
      <h1>Login</h1>
      <form @submit.prevent="login">
        <Input v-model="username" type="text" placeholder="Username" required />
        <Input v-model="password" type="password" placeholder="Password" required />
        <Button type="submit">Login</Button>
      </form>
      <p v-if="error">{{ error }}</p>
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref } from 'vue';
  import { Input } from '@/components/ui/input';
  import { Button } from '@/components/ui/button';
  
  const username = ref('');
  const password = ref('');
  const error = ref('');
  
  const login = async () => {
    const apiUrl = `https://napi.${username.value}.hackclub.app/login`;
    try {
      const response = await fetch(apiUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: username.value, password: password.value }),
      });
      const result = await response.json();
      if (response.ok) {
        // Save access token and refresh token to localStorage
        localStorage.setItem('accessToken', result.access_token);
        localStorage.setItem('refreshToken', result.refresh_token);
  
        // Optionally, store the username if needed
        localStorage.setItem('username', username.value);
  
        // Emit the logged-in event
        this.$emit('logged-in');
  
        // Redirect to the main application page or dashboard
        this.$router.push({ name: 'dashboard' });
      } else {
        error.value = result.message || 'Invalid username or password. Please try again.';
      }
    } catch (err) {
      error.value = 'An error occurred. Please try again.';
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
  margin-bottom: 0.1rem;
  font-size: 300%;
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