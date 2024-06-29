<template>
    <div class="bg-neutral-800 flex flex-col items-center justify-center h-screen">
      <h1 class="text-white font-mono p-2 text-5xl">Login</h1>
      <form class="flex flex-col" @submit.prevent="login">
        <Input v-model="username" type="text" class="mb-2 p-2 font-mono text-lg bg-black" placeholder="Username" required />
        <Input v-model="password" type="password" class="mb-2 p-2 font-mono text-lg bg-black" placeholder="Password" required />
        <Button class="p-2 text-lg cursor-pointer bg-slate-500 bg-blue-800" type="submit">Login</Button>
      </form>
      <p class="bg-red-600" v-if="error">{{ error }}</p>
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