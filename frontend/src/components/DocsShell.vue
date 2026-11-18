<template>
  <div class="page">
    <div class="eyebrow">{{ eyebrow }}</div><h1>{{ title }}</h1>
    <p class="muted" style="max-width:760px;font-size:16px;line-height:1.7">{{ description }}</p>
    <a-tabs default-active-key="content" style="margin-top:18px">
      <a-tab-pane key="content" title="Content"><slot /></a-tab-pane>
    </a-tabs>
  </div>
</template>
<script setup lang="ts">
defineProps<{ eyebrow: string; title: string; description: string }>()
</script>
