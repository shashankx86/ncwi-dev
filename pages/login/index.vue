<template>
    <div class="bg-neutral-800 flex flex-col items-center justify-center h-screen">
      <h1 class="text-white font-mono p-2 text-5xl">Login</h1>
      <form class="flex flex-col" @submit.prevent="login">
        <Input v-model="username" type="text" class="mb-2 p-2 font-mono text-lg bg-black text-white" placeholder="Username" required />
        <Input v-model="password" type="password" class="mb-2 p-2 font-mono text-lg bg-black text-white" placeholder="Password" required />
        <Button class="p-2 text-lg cursor-pointer bg-slate-500 bg-blue-800" type="submit">Login</Button>
      </form>
      <p class="bg-red-600" v-if="error">{{ error }}</p>
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref } from 'vue';
  import { useRouter } from 'vue-router';
  import { Input } from '@/components/ui/input';
  import { Button } from '@/components/ui/button';
  
  const username = ref('');
  const password = ref('');
  const error = ref('');
  const router = useRouter();
  
  const login = async () => {
    const apiUrl = `https://napi.${username.value}.hackclub.app/login`;
    try {
      const response = await fetch(apiUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: username.value, password: password.value })
      });
  
      // Debug: Log the raw response
      console.log(response);
  
      const result = await response.json();
  
      // Debug: Log the result
      console.log(result);
  
      if (response.ok) {
        localStorage.setItem('accessToken', result.access_token);
        localStorage.setItem('refreshToken', result.refresh_token);
        localStorage.setItem('username', username.value);
  
        // Redirect to the home page
        router.push('/home');
      } else {
        // Display the error message from the response
        error.value = result.message || 'Invalid username or password. Please try again.';
      }
    } catch (err) {
      // Debug: Log the error
      console.log(err);
  
      error.value = 'An error occurred. Please try again.';
    }
  };
  </script>