<template>
  <div class="legal-page">
    <div class="legal-inner">
      <!-- Header -->
      <div class="legal-eyebrow eyebrow">identity / legal</div>
      <a-typography-title :heading="2" class="legal-title" style="margin:0 0 8px">Legal</a-typography-title>
      <a-typography-paragraph class="legal-p">Last updated: {{ lastUpdated }}</a-typography-paragraph>

      <!-- Tab navigation -->
      <nav class="legal-tabs">
        <a
          v-for="section in sections"
          :key="section.id"
          :href="`#${section.id}`"
          :class="['legal-tab', { active: activeSection === section.id }]"
          @click.prevent="goTo(section.id)"
        >
          {{ section.label }}
        </a>
      </nav>

      <!-- Privacy Policy -->
      <h2 id="privacy" class="section-heading">Privacy Policy</h2>

      <a-typography-paragraph class="legal-p">
        Identity Verification ("we", "our", "us") is committed to protecting the privacy of developers
        and end users who interact with our identity verification platform. This policy
        describes how we collect, use, store, and protect personal information.
      </a-typography-paragraph>

      <h3 class="sub-heading">Information We Collect</h3>
      <ul class="legal-ul">
        <li v-for="item in privacyCollect" :key="item">{{ item }}</li>
      </ul>

      <h3 class="sub-heading">How We Use Your Information</h3>
      <ul class="legal-ul">
        <li v-for="item in privacyUse" :key="item">{{ item }}</li>
      </ul>

      <h3 class="sub-heading">Data Storage & Security</h3>
      <a-typography-paragraph class="legal-p">
        On Identity Verification Cloud (example.com), identity documents and biometric data are stored with
        AES-256 server-side encryption at rest. Application-layer AES-256-GCM encryption additionally
        protects stored third-party provider credentials (e.g. LLM and SMS API keys) and tokenized
        identity records held in our Identity Vault. Customer API keys are stored as HMAC-SHA256
        hashes — the raw key is never persisted. Data is transmitted exclusively over HTTPS/TLS 1.3.
        We retain verification records for the duration required by your configured retention policy
        (default: 90 days), after which they are permanently deleted.
      </a-typography-paragraph>
      <a-typography-paragraph class="legal-p">
        Self-hosted deployments (community edition) ship with local filesystem storage by default;
        operators are responsible for configuring encryption at rest, either via filesystem-level
        encryption (LUKS, dm-crypt, EBS volume encryption) or by switching to S3-compatible storage
        with server-side encryption enabled.
      </a-typography-paragraph>

      <h3 class="sub-heading">Third-Party Sharing</h3>
      <a-typography-paragraph class="legal-p">
        We do not sell personal data. We share information only with: (1) the developer/organization
        that initiated the verification request, (2) infrastructure providers necessary for service
        operation (hosting, database), and (3) law enforcement when required by law.
      </a-typography-paragraph>

      <h3 class="sub-heading">Your Rights</h3>
      <a-typography-paragraph class="legal-p">
        You may request access to, correction of, or deletion of your personal data at any time by
        contacting us at <a href="mailto:privacy@example.com" class="legal-link">privacy@example.com</a>.
        For GDPR-specific rights, see the GDPR Compliance section below.
      </a-typography-paragraph>

      <a-divider class="legal-divider" />

      <!-- Terms of Service -->
      <h2 id="terms" class="section-heading">Terms of Service</h2>

      <a-typography-paragraph class="legal-p">
        By accessing or using the Identity Verification API, dashboard, or verification services, you agree
        to be bound by these Terms of Service. If you do not agree, do not use our services.
      </a-typography-paragraph>

      <h3 class="sub-heading">1. Account & API Access</h3>
      <a-typography-paragraph class="legal-p">
        You must register for a developer account to access the Identity Verification API. You are responsible for
        maintaining the confidentiality of your API keys and for all activity that occurs under your
        account. You must notify us immediately of any unauthorized use.
      </a-typography-paragraph>

      <h3 class="sub-heading">2. Acceptable Use</h3>
      <a-typography-paragraph class="legal-p">You agree not to:</a-typography-paragraph>
      <ul class="legal-ul">
        <li v-for="item in termsAcceptable" :key="item">{{ item }}</li>
      </ul>

      <h3 class="sub-heading">3. Service Level & Availability</h3>
      <a-typography-paragraph class="legal-p">
        We strive for high availability but do not guarantee uninterrupted service. Scheduled
        maintenance windows will be communicated in advance. Our system status is available at
        <a href="/status" class="legal-link">example.com/status</a>.
      </a-typography-paragraph>

      <h3 class="sub-heading">4. Verification Accuracy</h3>
      <a-typography-paragraph class="legal-p">
        While we target &gt;90% document validation accuracy, verification results are probabilistic
        and should not be the sole basis for critical decisions. We recommend implementing manual
        review workflows for edge cases. Identity Verification is not liable for decisions made solely based on
        automated verification results.
      </a-typography-paragraph>

      <h3 class="sub-heading">5. Data Retention & Deletion</h3>
      <a-typography-paragraph class="legal-p">
        Verification data is retained according to your organization's configured retention period
        (default: 30 days). You may request early deletion of specific records via the API or admin
        dashboard. Upon account termination, all associated data is deleted within 30 days.
      </a-typography-paragraph>

      <h3 class="sub-heading">6. Limitation of Liability</h3>
      <a-typography-paragraph class="legal-p">
        To the maximum extent permitted by law, Identity Verification shall not be liable for any indirect,
        incidental, special, consequential, or punitive damages arising from your use of the service,
        including but not limited to loss of profits, data, or business opportunities.
      </a-typography-paragraph>

      <h3 class="sub-heading">7. Termination</h3>
      <a-typography-paragraph class="legal-p">
        We may suspend or terminate your access if you violate these Terms, engage in abusive behavior,
        or fail to pay applicable fees. You may terminate your account at any time through the developer
        dashboard. Termination does not relieve you of obligations incurred prior to termination.
      </a-typography-paragraph>

      <h3 class="sub-heading">8. Changes to Terms</h3>
      <a-typography-paragraph class="legal-p">
        We may update these Terms from time to time. Material changes will be communicated via email
        to registered developers at least 30 days before taking effect. Continued use after changes
        constitutes acceptance.
      </a-typography-paragraph>

      <a-divider class="legal-divider" />

      <!-- GDPR Compliance -->
      <h2 id="gdpr" class="section-heading">GDPR Compliance</h2>

      <a-typography-paragraph class="legal-p">
        Identity Verification is committed to compliance with the General Data Protection Regulation (EU) 2016/679.
        This section outlines how we handle personal data for users in the European Economic Area (EEA).
      </a-typography-paragraph>

      <h3 class="sub-heading">Legal Basis for Processing</h3>
      <a-typography-paragraph class="legal-p">We process personal data under the following legal bases:</a-typography-paragraph>
      <ul class="legal-ul">
        <li v-for="item in gdprBasis" :key="item">{{ item }}</li>
      </ul>

      <h3 class="sub-heading">Data Processing Roles</h3>
      <a-typography-paragraph class="legal-p">
        When you (the developer/organization) use Identity Verification to verify your users' identities, you act
        as the <strong class="legal-strong">Data Controller</strong> and Identity Verification acts as the
        <strong class="legal-strong">Data Processor</strong>. We process data solely on your
        instructions and in accordance with our Data Processing Agreement (DPA).
      </a-typography-paragraph>

      <h3 class="sub-heading">Data Subject Rights</h3>
      <a-typography-paragraph class="legal-p">Under GDPR, individuals whose data we process have the right to:</a-typography-paragraph>
      <ul class="legal-ul">
        <li v-for="item in gdprRights" :key="item">{{ item }}</li>
      </ul>
      <a-typography-paragraph class="legal-p">
        To exercise these rights, contact us at
        <a href="mailto:gdpr@example.com" class="legal-link">gdpr@example.com</a>.
        We respond to all requests within 30 days.
      </a-typography-paragraph>

      <h3 class="sub-heading">Data Retention</h3>
      <a-typography-paragraph class="legal-p">
        Identity documents and biometric data are retained only for the verification session duration
        plus your configured retention period (default: 30 days). After this period, all personal data
        is permanently and irreversibly deleted from our systems, including backups.
      </a-typography-paragraph>

      <h3 class="sub-heading">International Transfers</h3>
      <a-typography-paragraph class="legal-p">
        When personal data is transferred outside the EEA, we ensure appropriate safeguards are in
        place, including Standard Contractual Clauses (SCCs) approved by the European Commission.
      </a-typography-paragraph>

      <h3 class="sub-heading">Data Processing Agreement</h3>
      <a-typography-paragraph class="legal-p">
        We offer a pre-signed DPA to all customers processing EEA personal data through our platform.
        To request a copy, contact
        <a href="mailto:legal@example.com" class="legal-link">legal@example.com</a>.
      </a-typography-paragraph>

      <h3 class="sub-heading">Data Protection Officer</h3>
      <a-typography-paragraph class="legal-p">
        For GDPR-related inquiries, you may contact our Data Protection team at
        <a href="mailto:dpo@example.com" class="legal-link">dpo@example.com</a>.
      </a-typography-paragraph>

      <h3 class="sub-heading">CCPA (California)</h3>
      <a-typography-paragraph class="legal-p">
        California residents have additional rights under the California Consumer Privacy Act (CCPA),
        including the right to know what personal information is collected, the right to delete, and
        the right to opt out of the sale of personal information. We do not sell personal information.
        To exercise your CCPA rights, contact
        <a href="mailto:privacy@example.com" class="legal-link">privacy@example.com</a>.
      </a-typography-paragraph>

      <!-- Bottom spacer + contact -->
      <a-divider class="legal-divider" />
      <a-card class="legal-contact">
        <h3 class="contact-title">Questions?</h3>
        <a-typography-paragraph class="legal-p" style="margin-bottom:0">
          If you have questions about this legal documentation, contact us at
          <a href="mailto:legal@example.com" class="legal-link">legal@example.com</a>.
        </a-typography-paragraph>
      </a-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const lastUpdated = 'March 22, 2026'

const sections = [
  { id: 'privacy', label: 'Privacy Policy' },
  { id: 'terms', label: 'Terms of Service' },
  { id: 'gdpr', label: 'GDPR Compliance' },
] as const

const activeSection = computed(() => route.hash.slice(1) || 'privacy')

const privacyCollect = [
  'Account information: email address, name, and company name when you register as a developer.',
  "Identity documents: government-issued IDs uploaded for verification (passport, driver's license, national ID).",
  'Biometric data: facial images captured during live capture for face matching.',
  'Usage data: API call logs, verification request metadata, timestamps, and IP addresses.',
  'Device information: browser type, operating system, and device identifiers during verification sessions.',
]

const privacyUse = [
  'To process identity verification requests submitted through our API.',
  'To detect fraud and prevent abuse of our verification services.',
  'To improve our OCR, document recognition, and face matching accuracy.',
  'To communicate important service updates, security alerts, and account information.',
  'To generate anonymized, aggregate analytics for platform monitoring.',
]

const termsAcceptable = [
  'Use the service for any unlawful purpose or to facilitate identity fraud.',
  'Attempt to reverse-engineer, decompile, or extract source code from our verification pipeline.',
  'Submit fabricated or manipulated documents to test or exploit our verification system.',
  "Exceed rate limits or use automated tools to abuse the API beyond your plan's allocation.",
  'Resell or redistribute API access without a written agreement.',
]

const gdprBasis = [
  'Contractual necessity: processing verification requests as part of the service you signed up for.',
  'Legitimate interest: fraud prevention, service security, and platform improvement.',
  'Consent: where explicitly provided, such as optional biometric face matching.',
  'Legal obligation: compliance with applicable laws and regulations.',
]

const gdprRights = [
  'Access: request a copy of personal data we hold about them.',
  'Rectification: request correction of inaccurate personal data.',
  'Erasure ("right to be forgotten"): request deletion of personal data.',
  'Restriction: request that we limit processing of their data.',
  'Portability: receive personal data in a structured, machine-readable format.',
  'Object: object to processing based on legitimate interest.',
]

function goTo(id: string) {
  router.push({ path: '/legal', hash: `#${id}` })
}

function scrollToHash() {
  if (!route.hash) {
    window.scrollTo(0, 0)
    return
  }
  nextTick(() => {
    setTimeout(() => {
      document.getElementById(route.hash.slice(1))?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }, 100)
  })
}

onMounted(scrollToHash)
watch(() => route.hash, scrollToHash)
</script>

<style scoped>
.legal-page { background: #f7f6f3; min-height: 100vh; }
.legal-inner { max-width: 768px; margin: 0 auto; padding: 48px 24px 80px; }
.legal-eyebrow { color: #007a7a; margin-bottom: 12px; }
.legal-title { font-family: ui-monospace, Menlo, monospace; font-weight: 500; color: #0b0b0d; letter-spacing: -.02em; }
.legal-p { font-size: .9rem; line-height: 1.75; color: #6b6b70; margin-bottom: 14px; }
.legal-link { color: #007a7a; text-decoration: none; }
.legal-link:hover { text-decoration: underline; }
.legal-strong { color: #0b0b0d; }
.legal-tabs { display: flex; border: 1px solid #e4e2dc; margin-bottom: 40px; }
.legal-tab {
  flex: 1;
  text-align: center;
  padding: 10px 16px;
  font-family: ui-monospace, Menlo, monospace;
  font-size: .75rem;
  font-weight: 500;
  letter-spacing: .04em;
  text-decoration: none;
  background: transparent;
  color: #6b6b70;
  border-bottom: 2px solid transparent;
  border-right: 1px solid #e4e2dc;
  transition: border-color .15s;
}
.legal-tab:last-child { border-right: none; }
.legal-tab.active { background: #fff; color: #007a7a; border-bottom-color: #0b0b0d; }
.section-heading {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 1.35rem;
  font-weight: 600;
  color: #007a7a;
  letter-spacing: -.01em;
  margin: 0 0 8px;
  scroll-margin-top: 90px;
}
.sub-heading {
  font-family: ui-monospace, Menlo, monospace;
  font-size: .95rem;
  font-weight: 500;
  color: #0b0b0d;
  margin: 28px 0 8px;
}
.legal-ul { padding-left: 20px; margin: 0 0 14px; }
.legal-ul li { font-size: .9rem; line-height: 1.75; color: #6b6b70; margin-bottom: 4px; list-style-type: disc; }
.legal-divider { margin: 48px 0; border-color: #e4e2dc; }
.legal-contact :deep(.arco-card-body) { padding: 24px 28px; }
.contact-title { font-family: ui-monospace, Menlo, monospace; font-size: .85rem; font-weight: 500; color: #0b0b0d; margin: 0 0 8px; }
</style>
