
<script setup lang="ts">
</script>

<template>
  <div class="toplevel">
    <h1>App</h1>
    <h3>
      Welcome to a app.
    </h3>
  </div>
</template>

<style scoped>
h1 {
  font-weight: 500;
  font-size: 2.6rem;
  position: relative;

  top: -10px;
}

h3 {
  font-size: 1.2rem;
}

.toplevel h1,
.toplevel h3 {
  text-align: center;
}

@media (min-width: 1024px) {
  .toplevel h1,
  .toplevel h3 {
    text-align: left;
  }
}
</style>
