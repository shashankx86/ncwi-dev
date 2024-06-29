<template>
    <div>
      <Navbar />
      <div class="content">
        <h1>Welcome to the Home Page</h1>
      </div>
    </div>
  </template>
  
  <script setup lang="ts">
  import Navbar from '@/components/Navbar.vue';
  </script>
  
  <style scoped>
  .content {
    padding-top: 60px;
  }
  </style>
  