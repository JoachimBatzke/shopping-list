<template>
  <router-view />
</template>

<style>
/* Font imports */
@font-face {
  font-family: 'Heavitas';
  src: url('@/assets/fonts/Heavitas.ttf') format('truetype');
  font-weight: normal;
  font-style: normal;
  font-display: swap;
}

/* Global styles */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: system-ui, -apple-system, sans-serif;
  background: #f5f5f5;
  min-height: 100vh;
}

/* Utility class for Heavitas headings */
.font-heavitas {
  font-family: 'Heavitas', sans-serif;
}
</style>
