<template>
  <div :class="{'dark-mode': isDarkMode}" class="layout-wrapper min-h-screen">
    <router-view></router-view>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
const isDarkMode = ref(false);
</script>

<style>
.layout-wrapper {
  transition: background-color 0.3s ease;
}
</style>
