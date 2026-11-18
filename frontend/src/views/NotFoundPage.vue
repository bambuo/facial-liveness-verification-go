<template>
  <div class="page-narrow" style="text-align:center;padding-top:100px">
    <div class="eyebrow">404</div><h1 style="font-size:56px;margin:8px 0">Page not found</h1>
    <p class="muted">The page you requested does not exist or has moved.</p>
    <a-space><a-button type="primary" @click="$router.push('/')">Home</a-button><a-button @click="$router.push('/docs')">Documentation</a-button></a-space>
  </div>
</template>
