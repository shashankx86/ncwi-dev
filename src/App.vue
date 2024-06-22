<template>
  <div id="app">
    <Navbar />
    <div class="content">
      <h1>NCWI</h1>
    </div>
  </div>
</template>

<script>
import Navbar from './components/Navbar.vue';
import './style.css';

export default {
  name: 'App',
  components: {
    Navbar
  }
}
</script>

