<template>
    <nav class="navbar">
      <ul>
        <li><NuxtLink to="/">Home</NuxtLink></li>
        <li><NuxtLink to="/about">About</NuxtLink></li>
        <li><NuxtLink to="/contact">Contact</NuxtLink></li>
        <li><NuxtLink to="/profile">Profile</NuxtLink></li>
      </ul>
    </nav>
  </template>
  
  <script setup lang="ts">
  </script>
  
  <style scoped>
  .navbar {
    position: fixed;
    top: 0;
    width: 100%;
    background: #333;
    color: white;
    padding: 1em;
  }
  
  .navbar ul {
    list-style: none;
    display: flex;
    justify-content: space-around;
  }
  
  .navbar a {
    color: white;
    text-decoration: none;
  }
  </style>
  