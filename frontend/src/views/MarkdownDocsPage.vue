<template><DocsShell eyebrow="Content" title="Markdown documentation" description="Guides can be authored as Markdown and rendered through the docs pipeline.">
  <div class="panel"><h3>Example</h3><pre class="code-block"># Verify an identity

1. Initialize a verification
2. Upload the front document
3. Capture a selfie
4. Finalize

> Results are deterministic and auditable.</pre>
  <p class="muted" style="margin-top:14px">The production docs pipeline can fetch Markdown from your repository or object storage.</p></div>
</DocsShell>
</template>
<script setup lang="ts">
import DocsShell from '../components/DocsShell.vue'
</script>
