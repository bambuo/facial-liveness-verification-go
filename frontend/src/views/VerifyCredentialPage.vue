<template>
  <div class="vc-page">
    <div class="vc-inner">
      <!-- Header -->
      <header class="vc-head">
        <div class="vc-head-badge"><icon-check-circle /></div>
        <h1>Verify a credential</h1>
        <p class="vc-hint">
          Paste a JWT-VC or open a link with <code class="vc-mono">?jwt=</code>. Signatures are checked in your
          browser against the issuer's public key.
        </p>
      </header>

      <!-- Verifying -->
      <div v-if="verifying" class="vc-verifying">
        <a-spin :size="40" />
        <div class="vc-mono vc-stage">{{ stage }}</div>
        <p class="vc-hint">Verification happens client-side — the credential never leaves your browser.</p>
      </div>

      <!-- Input -->
      <template v-else-if="!result">
        <a-card class="vc-card">
          <div class="vc-label">HOW IT WORKS</div>
          <div class="vc-steps">
            <div v-for="step in howItWorks" :key="step.n" class="vc-step">
              <div class="vc-step-n">{{ step.n }}</div>
              <div>
                <div class="vc-step-title">{{ step.title }}</div>
                <div class="vc-hint">{{ step.desc }}</div>
              </div>
            </div>
          </div>
        </a-card>

        <a-textarea
          v-model="jwtInput"
          class="vc-textarea"
          :auto-size="{ minRows: 6, maxRows: 12 }"
          placeholder="eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9..."
          spellcheck="false"
        />

        <a-button type="primary" long size="large" class="vc-gap" :disabled="!jwtInput.trim()" @click="verifyFromInput">
          Verify credential
        </a-button>

        <p class="vc-hint vc-center vc-gap">
          No credential yet? <router-link to="/demo">Try the demo</router-link> to issue one.
        </p>
      </template>

      <!-- Results -->
      <template v-else>
        <div class="vc-result-head">
          <div class="vc-result-badge" :class="headerClass">{{ headerIcon }}</div>
          <h2>{{ headerLabel }}</h2>
          <p class="vc-hint">{{ headerDesc }}</p>
        </div>

        <a-card class="vc-card">
          <div class="vc-label">DETAILS</div>
          <a-descriptions :column="1" bordered size="small">
            <a-descriptions-item label="Signature">
              <a-tag :color="signatureOk ? 'green' : 'red'">{{ signatureOk ? 'Valid' : 'Invalid' }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="DID resolution">
              <a-tag :color="result.didResolved ? 'green' : 'red'">
                {{ result.didResolved ? 'Resolved' : 'Failed' }}
              </a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="Issuer">
              <span class="vc-mono">{{ result.issuer || 'N/A' }}</span>
            </a-descriptions-item>
            <a-descriptions-item label="Algorithm">
              <span class="vc-mono">{{ (result.header?.alg as string) || 'N/A' }}</span>
            </a-descriptions-item>
            <a-descriptions-item label="Expiration">
              <a-tag v-if="result.expiresAt" :color="result.expired ? 'red' : 'green'">
                {{ result.expired ? 'Expired' : 'Valid' }} — {{ formatDate(result.expiresAt) }}
              </a-tag>
              <span v-else class="vc-hint">No expiration</span>
            </a-descriptions-item>
          </a-descriptions>
        </a-card>

        <a-card v-if="claimRows.length" class="vc-card">
          <div class="vc-label">CLAIMS</div>
          <a-descriptions :data="claimRows" :column="1" bordered size="small" />
        </a-card>

        <a-card v-if="statusResult" class="vc-card">
          <div class="vc-label">REVOCATION STATUS</div>
          <a-space>
            <a-tag :color="statusResult.active ? 'green' : 'red'">
              {{ statusResult.active ? 'Active' : 'Inactive' }}
            </a-tag>
            <span v-if="statusResult.reason && !statusResult.active" class="vc-hint vc-mono">
              {{ statusResult.reason }}
            </span>
          </a-space>
        </a-card>

        <a-card v-if="result.header" class="vc-card">
          <div class="vc-label">JWT HEADER</div>
          <pre class="vc-json">{{ pretty(result.header) }}</pre>
        </a-card>

        <a-card v-if="result.payload" class="vc-card">
          <div class="vc-label">JWT PAYLOAD</div>
          <pre class="vc-json vc-json-tall">{{ pretty(result.payload) }}</pre>
        </a-card>

        <a-card class="vc-card">
          <div class="vc-label">RAW JWT</div>
          <div class="vc-raw">
            <span class="vc-raw-head">{{ jwtParts[0] }}</span>
            <span>.</span>
            <span class="vc-raw-body">{{ jwtParts[1] }}</span>
            <span>.</span>
            <span class="vc-raw-sig">{{ jwtParts[2] }}</span>
          </div>
        </a-card>

        <a-alert v-if="errorMessage" type="error" class="vc-gap">{{ errorMessage }}</a-alert>

        <a-space class="vc-actions">
          <a-button
            v-if="jti && !statusResult"
            type="primary"
            :loading="checkingStatus"
            @click="checkRevocation"
          >
            {{ checkingStatus ? 'Checking…' : 'Check revocation' }}
          </a-button>
          <a-button @click="reset">Verify another</a-button>
        </a-space>
      </template>

      <footer class="vc-footer">
        Powered by Identity Verification · W3C Verifiable Credentials · DID:web
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, apiMessage } from '../lib/api'

interface CredentialResult {
  valid: boolean
  header: Record<string, any> | null
  payload: Record<string, any> | null
  claims: Record<string, any> | null
  issuer: string | null
  expiresAt: Date | null
  expired: boolean
  error?: string
  didResolved: boolean
}

const route = useRoute()

const jwtInput = ref('')
const verifying = ref(false)
const stage = ref('')
const result = ref<CredentialResult | null>(null)
const statusResult = ref<{ active: boolean; reason?: string } | null>(null)
const checkingStatus = ref(false)
const errorMessage = ref('')

const howItWorks = [
  { n: '1', title: 'Decode', desc: 'The JWT header and payload are decoded in your browser.' },
  { n: '2', title: 'Resolve', desc: 'The issuer DID:web document is fetched to obtain the Ed25519 public key.' },
  { n: '3', title: 'Verify', desc: 'The signature is verified and the expiration is checked.' },
]

const signatureOk = computed(() => {
  const value = result.value
  if (!value) return false
  return value.valid || (value.expired && value.didResolved && !value.error?.includes('signature'))
})
const isExpiredButSigOk = computed(() => {
  const value = result.value
  return !!value && !value.valid && value.expired && value.didResolved
})
const headerClass = computed(() => (result.value?.valid ? 'vc-ok' : isExpiredButSigOk.value ? 'vc-warn' : 'vc-bad'))
const headerIcon = computed(() => (result.value?.valid ? '✓' : isExpiredButSigOk.value ? '⚠' : '✗'))
const headerLabel = computed(() => (result.value?.valid ? 'Valid credential' : isExpiredButSigOk.value ? 'Expired credential' : 'Invalid credential'))
const headerDesc = computed(() => {
  if (result.value?.valid) return "The Ed25519 signature was verified against the issuer's public key."
  if (isExpiredButSigOk.value) return 'The signature is valid but the credential has expired.'
  return result.value?.error || 'Signature verification failed.'
})

const jti = computed<string | null>(() => {
  const payload = result.value?.payload
  if (!payload) return null
  return (payload.jti as string) || ((payload.vc as Record<string, any> | undefined)?.id as string) || null
})

const claimRows = computed(() => (result.value?.claims ? Object.entries(result.value.claims).map(([key, value]) => ({ label: key, value: String(value) })) : []))
const jwtParts = computed(() => jwtInput.value.trim().split('.'))

function pretty(value: unknown) {
  return JSON.stringify(value, null, 2)
}

function formatDate(date: Date) {
  return date.toLocaleDateString(undefined, { day: '2-digit', month: 'short', year: 'numeric' })
}

function base64urlDecode(value: string): Uint8Array {
  const padded = value.replace(/-/g, '+').replace(/_/g, '/')
  const padding = (4 - (padded.length % 4)) % 4
  const binary = atob(padded + '='.repeat(padding))
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes
}

function decodeSegment(segment: string): Record<string, any> | null {
  try {
    return JSON.parse(new TextDecoder().decode(base64urlDecode(segment)))
  } catch {
    return null
  }
}

const didCache = new Map<string, { doc: any; fetchedAt: number }>()
const DID_CACHE_TTL = 5 * 60 * 1000

function didWebToUrl(did: string): string | null {
  if (!did.startsWith('did:web:')) return null
  const decoded = decodeURIComponent(did.slice('did:web:'.length))
  const parts = decoded.split(':')
  const host = parts[0]
  const pathSegments = parts.slice(1)
  return pathSegments.length === 0
    ? `https://${host}/.well-known/did.json`
    : `https://${host}/${pathSegments.join('/')}/did.json`
}

async function fetchDIDDocument(did: string) {
  const now = Date.now()
  for (const [key, entry] of didCache) {
    if (now - entry.fetchedAt >= DID_CACHE_TTL) didCache.delete(key)
  }
  const cached = didCache.get(did)
  if (cached) return cached.doc
  const url = didWebToUrl(did)
  if (!url) throw new Error(`Unsupported DID method in ${did}. Only did:web is supported.`)
  const response = await fetch(url)
  if (!response.ok) throw new Error(`Failed to resolve DID document: HTTP ${response.status}`)
  const doc = await response.json()
  didCache.set(did, { doc, fetchedAt: now })
  return doc
}

function extractPublicKey(doc: any, kid?: string): Uint8Array {
  const methods = doc?.verificationMethod
  if (!Array.isArray(methods) || methods.length === 0) throw new Error('No verification methods in DID document')
  let method
  if (kid) {
    method = methods.find((entry: any) => entry.id === kid)
    if (!method) throw new Error(`Verification method not found for kid: ${kid}`)
  } else {
    method = methods.find((entry: any) => entry.publicKeyJwk?.crv === 'Ed25519' || entry.type === 'JsonWebKey2020')
  }
  if (!method?.publicKeyJwk?.x) throw new Error('No Ed25519 public key found in DID document')
  return base64urlDecode(method.publicKeyJwk.x)
}

async function verifyCredential(jwt: string): Promise<CredentialResult> {
  const empty: CredentialResult = {
    valid: false, header: null, payload: null, claims: null, issuer: null,
    expiresAt: null, expired: false, didResolved: false,
  }
  const parts = jwt.split('.')
  if (parts.length !== 3) {
    return { ...empty, error: 'Invalid JWT format — expected 3 dot-separated segments' }
  }
  const [headerSegment, payloadSegment, signatureSegment] = parts
  stage.value = 'DECODING JWT…'

  const header = decodeSegment(headerSegment)
  if (!header) return { ...empty, error: 'Failed to decode JWT header' }
  if (header.alg !== 'EdDSA') {
    return { ...empty, header, error: `Unsupported algorithm: ${header.alg} (expected EdDSA)` }
  }
  const payload = decodeSegment(payloadSegment)
  if (!payload) return { ...empty, header, error: 'Failed to decode JWT payload' }

  const issuer = (payload.iss as string) || null
  const claims = ((payload.vc as Record<string, any> | undefined)?.credentialSubject as Record<string, any> | undefined) ?? null
  const expiresAt = payload.exp ? new Date(Number(payload.exp) * 1000) : null
  const expired = expiresAt ? expiresAt < new Date() : false
  if (!issuer) {
    return { ...empty, header, payload, claims, expiresAt, expired, error: 'No issuer (iss) claim in JWT' }
  }

  stage.value = 'RESOLVING DID DOCUMENT…'
  let publicKey: Uint8Array
  try {
    const doc = await fetchDIDDocument(issuer)
    publicKey = extractPublicKey(doc, header.kid as string | undefined)
  } catch (error) {
    return { ...empty, header, payload, claims, issuer, expiresAt, expired, error: (error as Error).message }
  }

  stage.value = 'VERIFYING SIGNATURE…'
  try {
    const algorithm = { name: 'Ed25519' }
    const key = await crypto.subtle.importKey('raw', publicKey as unknown as BufferSource, algorithm, false, ['verify'])
    const data = new TextEncoder().encode(`${headerSegment}.${payloadSegment}`)
    const signature = base64urlDecode(signatureSegment)
    const valid = await crypto.subtle.verify(algorithm, key, signature as unknown as BufferSource, data)
    return { valid, header, payload, claims, issuer, expiresAt, expired, didResolved: true }
  } catch (error) {
    return {
      ...empty, header, payload, claims, issuer, expiresAt, expired, didResolved: true,
      error: `This browser could not verify the signature: ${(error as Error).message}`,
    }
  }
}

async function verifyFromInput() {
  const jwt = jwtInput.value.trim()
  if (!jwt) return
  verifying.value = true
  result.value = null
  statusResult.value = null
  errorMessage.value = ''
  try {
    result.value = await verifyCredential(jwt)
  } finally {
    verifying.value = false
    stage.value = ''
  }
}

async function checkRevocation() {
  if (!jti.value) return
  checkingStatus.value = true
  try {
    const clean = jti.value.replace(/^urn:uuid:/, '')
    const { data } = await api.get(`/v2/credentials/${encodeURIComponent(clean)}/status`)
    statusResult.value = data
  } catch (error) {
    const status = (error as { response?: { status?: number } }).response?.status
    errorMessage.value = status === 404 ? apiMessage(error) : ''
    statusResult.value = { active: false, reason: 'error' }
  } finally {
    checkingStatus.value = false
  }
}

function reset() {
  result.value = null
  statusResult.value = null
  jwtInput.value = ''
  errorMessage.value = ''
}

onMounted(() => {
  const value = route.query.jwt
  if (typeof value === 'string' && value) {
    jwtInput.value = value
    void verifyFromInput()
  }
})
</script>

<style scoped>
.vc-page { min-height: 100vh; background: #f7f6f3; color: #0b0b0d; }
.vc-inner { max-width: 640px; margin: 0 auto; padding: 64px 24px 56px; }
.vc-mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.vc-hint { color: #6b6b70; font-size: 13px; line-height: 1.6; margin: 0; }
.vc-center { text-align: center; }
.vc-gap { margin-top: 14px; }
.vc-head { text-align: center; margin-bottom: 32px; }
.vc-head-badge { width: 56px; height: 56px; margin: 0 auto 16px; display: grid; place-items: center; border: 1px solid #2563eb; color: #2563eb; font-size: 22px; }
.vc-head h1 { font-family: ui-monospace, Menlo, monospace; font-size: 26px; font-weight: 500; letter-spacing: -.02em; margin: 0 0 8px; }
.vc-card { margin-bottom: 16px; border-radius: 4px; }
.vc-label { font-family: ui-monospace, Menlo, monospace; font-size: 11px; font-weight: 500; letter-spacing: .06em; color: #6b6b70; margin-bottom: 12px; }
.vc-steps { display: flex; flex-direction: column; gap: 14px; }
.vc-step { display: flex; gap: 12px; align-items: flex-start; }
.vc-step-n { width: 24px; height: 24px; flex-shrink: 0; display: grid; place-items: center; border: 1px solid #2563eb; color: #2563eb; font-family: ui-monospace, Menlo, monospace; font-size: 11px; }
.vc-step-title { font-size: 13px; font-weight: 600; margin-bottom: 2px; }
.vc-textarea :deep(.arco-textarea-wrapper) { background: #0f172a; border: 1px solid #1e293b; border-radius: 0; }
.vc-textarea :deep(textarea) { background: #0f172a; color: #e2e8f0; font-family: ui-monospace, Menlo, monospace; font-size: 12px; line-height: 1.6; }
.vc-textarea :deep(textarea::placeholder) { color: #475569; }
.vc-verifying { display: flex; flex-direction: column; align-items: center; gap: 12px; padding: 48px 0; text-align: center; }
.vc-stage { font-family: ui-monospace, Menlo, monospace; font-size: 12px; letter-spacing: .08em; color: #2563eb; }
.vc-result-head { text-align: center; margin-bottom: 24px; }
.vc-result-badge { width: 56px; height: 56px; margin: 0 auto 14px; display: grid; place-items: center; font-size: 24px; border: 1px solid; }
.vc-ok { color: #15803d; background: #ecfdf5; border-color: #15803d; }
.vc-warn { color: #b45309; background: #fdf6e3; border-color: #b45309; }
.vc-bad { color: #b91c1c; background: #fef2f2; border-color: #b91c1c; }
.vc-result-head h2 { font-family: ui-monospace, Menlo, monospace; font-size: 21px; font-weight: 500; margin: 0 0 6px; }
.vc-json { margin: 0; padding: 14px; background: #0f172a; color: #e2e8f0; border: 1px solid #1e293b; font-family: ui-monospace, Menlo, monospace; font-size: 11px; line-height: 1.5; overflow: auto; }
.vc-json-tall { max-height: 320px; }
.vc-raw { padding: 14px; background: #0f172a; border: 1px solid #1e293b; font-family: ui-monospace, Menlo, monospace; font-size: 9px; line-height: 1.6; word-break: break-all; color: #64748b; max-height: 120px; overflow-y: auto; }
.vc-raw-head { color: #38bdf8; }
.vc-raw-body { color: #34d399; }
.vc-raw-sig { color: #fbbf24; }
.vc-actions { display: flex; justify-content: center; margin-top: 8px; }
.vc-footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #e4e2dc; text-align: center; font-family: ui-monospace, Menlo, monospace; font-size: 11px; letter-spacing: .02em; color: #9a978f; }
</style>
