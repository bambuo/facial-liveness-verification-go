<template>
  <div class="page-narrow">
    <div class="eyebrow">Legal</div><h1>Privacy, terms and data handling</h1>
    <p class="muted">This self-hosted deployment processes identity documents for verification. Configure your own retention and privacy policy before production use.</p>
    <a-collapse :default-active-key="['privacy']" style="margin-top:24px">
      <a-collapse-item header="Privacy" key="privacy"><p>Biometric embeddings are processed in memory and are not persisted by default. Document images follow your configured storage and retention policy.</p></a-collapse-item>
      <a-collapse-item header="Terms of service" key="terms"><p>Use the service only with lawful consent and within applicable identity-verification regulations. You are responsible for your deployment configuration.</p></a-collapse-item>
      <a-collapse-item header="Data retention" key="retention"><p>Retention is controlled by the deployment operator. Configure retention windows and deletion workflows before onboarding users.</p></a-collapse-item>
      <a-collapse-item header="Security contact" key="security"><p>Report security issues to your deployment operator. Do not include identity documents or secrets in reports.</p></a-collapse-item>
    </a-collapse>
  </div>
</template>
