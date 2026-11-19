<template>
  <div class="builder">
    <div class="builder-bar">
      <div class="builder-title">
        <router-link to="/developer" class="back-link">←</router-link>
        <div>
          <div class="eyebrow">Hosted verification page</div>
          <h2>Page builder</h2>
        </div>
        <a-tag v-if="dirty" color="orange">Unsaved changes</a-tag>
      </div>
      <a-space>
        <a-button :loading="loading" @click="loadAll">Reload</a-button>
        <a-button @click="resetDefaults">Defaults</a-button>
        <a-button type="primary" :loading="saving" @click="saveConfig">
          {{ saving ? 'Saving…' : 'Save' }}
        </a-button>
      </a-space>
    </div>

    <a-alert v-if="!auth.isDeveloper" type="warning" class="builder-alert">
      Connect an API key or sign in to a developer account to load and save this page.
      <template #action><a-button size="mini" @click="$router.push('/developer')">Open portal</a-button></template>
    </a-alert>
    <a-alert v-if="saveError" type="error" class="builder-alert">
      {{ saveError }}
      <template #action><a-button size="mini" :loading="saving" @click="saveConfig">Retry</a-button></template>
    </a-alert>
    <a-alert v-if="savedAt" type="success" class="builder-alert">
      Configuration saved at {{ savedAt }}.
    </a-alert>

    <div class="builder-body">
      <!-- ── Controls ──────────────────────────────────────── -->
      <a-card class="panel-controls" :bordered="true">
        <a-form :model="config" layout="vertical">
          <div class="section">
            <div class="section-title">Header</div>
            <a-form-item label="Page title">
              <a-input v-model="config.headerTitle" :max-length="200" show-word-limit />
            </a-form-item>
            <a-form-item label="Subtitle">
              <a-input v-model="config.headerSubtitle" :max-length="300" show-word-limit />
            </a-form-item>
          </div>

          <div class="section">
            <div class="section-title">Logo</div>
            <div class="logo-row">
              <div class="logo-box">
                <img v-if="renderableLogo" :src="renderableLogo" alt="Brand logo" />
                <span v-else class="muted">None</span>
              </div>
              <div class="logo-actions">
                <a-upload
                  accept="image/png,image/jpeg"
                  :auto-upload="false"
                  :show-file-list="false"
                  :disabled="uploadingLogo || !auth.isDeveloper"
                  @change="onLogoChange"
                >
                  <template #upload-button>
                    <a-button :loading="uploadingLogo">
                      {{ renderableLogo ? 'Replace logo' : 'Upload logo' }}
                    </a-button>
                  </template>
                </a-upload>
                <small class="muted">PNG or JPEG, up to 2 MB. Uploaded to developer branding.</small>
                <small v-if="logoReference && !renderableLogo" class="muted">
                  Stored reference: <span class="mono">{{ logoReference }}</span> (re-upload to preview)
                </small>
              </div>
            </div>
          </div>

          <div class="section">
            <div class="section-title">Theme</div>
            <a-radio-group v-model="config.theme" type="button" class="theme-group" @change="applyTheme">
              <a-radio value="dark">Dark</a-radio>
              <a-radio value="light">Light</a-radio>
            </a-radio-group>

            <div class="color-grid">
              <a-form-item v-for="color in COLOR_FIELDS" :key="color.key" :label="color.label" class="color-item">
                <a-space>
                  <a-input
                    :model-value="config[color.key]"
                    :max-length="7"
                    class="color-input"
                    @update:model-value="(value: string) => setColor(color.key, value)"
                  />
                  <input
                    type="color"
                    class="color-swatch"
                    :value="isHex(config[color.key]) ? config[color.key] : '#000000'"
                    @input="setColor(color.key, ($event.target as HTMLInputElement).value)"
                  />
                </a-space>
              </a-form-item>
            </div>
          </div>

          <div class="section">
            <div class="section-title">Typography</div>
            <a-form-item label="Font family">
              <a-select v-model="config.fontFamily">
                <a-option value="dm-sans">DM Sans</a-option>
                <a-option value="inter">Inter</a-option>
                <a-option value="system">System default</a-option>
              </a-select>
            </a-form-item>
          </div>

          <div class="section">
            <div class="section-title">Steps</div>
            <div v-for="step in STEP_FIELDS" :key="step.key" class="step-row">
              <a-switch
                :model-value="config.steps[step.key].enabled"
                @update:model-value="(value: string | number | boolean) => setStepEnabled(step.key, Boolean(value))"
              />
              <a-input
                :model-value="config.steps[step.key].label"
                :max-length="50"
                class="step-label-input"
                @update:model-value="(value: string) => setStepLabel(step.key, value)"
              />
            </div>
          </div>

          <div class="section">
            <div class="section-title">Completion</div>
            <a-form-item label="Completion title">
              <a-input v-model="config.completionTitle" :max-length="200" />
            </a-form-item>
            <a-form-item label="Completion message">
              <a-input v-model="config.completionMessage" :max-length="500" />
            </a-form-item>
            <div class="switch-row">
              <a-switch v-model="config.showConfetti" />
              <span class="muted">Show confetti on success</span>
            </div>
          </div>

          <div class="section">
            <div class="section-title">Options</div>
            <div class="switch-row">
              <a-switch v-model="config.showPoweredBy" />
              <span class="muted">Show “powered by” footer</span>
            </div>
          </div>

          <div class="section">
            <div class="section-title">Custom URL</div>
            <a-form-item label="Public slug">
              <div class="slug-row">
                <span class="muted mono">/v/</span>
                <a-input
                  :model-value="slug"
                  class="slug-input"
                  placeholder="acme-verify"
                  :max-length="50"
                  @update:model-value="onSlugInput"
                />
                <a-button :loading="savingSlug" @click="saveSlug">{{ savingSlug ? '…' : 'Save slug' }}</a-button>
              </div>
            </a-form-item>
            <small class="muted">
              Min 4 characters, lowercase letters, digits and dashes. Availability is validated when you save.
            </small>
            <div v-if="slugError" class="slug-error">{{ slugError }}</div>
            <div v-else-if="savedSlug" class="slug-ok">
              Live at <router-link :to="`/v/${savedSlug}`">/v/{{ savedSlug }}</router-link>
            </div>
          </div>
        </a-form>
      </a-card>

      <!-- ── Live preview ──────────────────────────────────── -->
      <div class="panel-preview">
        <a-space wrap class="preview-toggles">
          <a-radio-group v-model="previewDevice" type="button" size="small">
            <a-radio value="mobile">Mobile</a-radio>
            <a-radio value="desktop">Desktop</a-radio>
          </a-radio-group>
          <span class="divider" />
          <a-radio-group v-model="previewScreen" type="button" size="small">
            <a-radio value="landing">Landing</a-radio>
            <a-radio value="complete">Completion</a-radio>
          </a-radio-group>
        </a-space>

        <div class="device-frame" :class="previewDevice">
          <div class="device-screen" :style="previewThemeVars">
            <!-- Completion -->
            <div v-if="previewScreen === 'complete'" class="screen-complete">
              <div class="complete-mark">✓</div>
              <h3>{{ previewConfig.completionTitle }}</h3>
              <p>{{ previewConfig.completionMessage }}</p>
              <div class="complete-card">
                <div class="complete-row"><span>Status</span><strong>Verified</strong></div>
                <div class="complete-row"><span>Face match</span><strong>96%</strong></div>
                <div class="complete-row"><span>Liveness</span><strong>98%</strong></div>
              </div>
              <div v-if="config.showConfetti" class="complete-note">Confetti enabled</div>
            </div>

            <!-- Mobile landing -->
            <div v-else-if="previewDevice === 'mobile'" class="screen-mobile">
              <div class="mobile-status">
                <span>9:41</span>
                <span class="accent">Secure session</span>
                <span>●●●</span>
              </div>
              <img v-if="renderableLogo" :src="renderableLogo" alt="" class="mobile-logo" />
              <div class="mobile-steps">
                <div v-for="(label, index) in previewStepLabels" :key="label" class="mobile-step">
                  <i :class="{ active: index === 0 }" />
                  <small :class="{ active: index === 0 }">{{ label }}</small>
                </div>
              </div>
              <div class="mobile-copy">
                <div class="step-of">Step 1 of {{ previewStepLabels.length }}</div>
                <h3>{{ previewConfig.headerTitle }}</h3>
                <p>{{ previewConfig.headerSubtitle }}</p>
              </div>
              <div class="camera-placeholder">Camera preview</div>
              <div class="cta">{{ previewStepLabels[0] }} — start</div>
              <div v-if="config.showPoweredBy" class="powered-by">Powered by Identity Verification</div>
            </div>

            <!-- Desktop landing -->
            <div v-else class="screen-desktop">
              <img v-if="renderableLogo" :src="renderableLogo" alt="" class="desktop-logo" />
              <h3>{{ previewConfig.headerTitle }}</h3>
              <p>{{ config.headerSubtitle === DEFAULT_CONFIG.headerSubtitle ? 'Choose how to continue' : previewConfig.headerSubtitle }}</p>
              <div class="desktop-methods">
                <div class="method primary">Continue on this device</div>
                <div class="method">Scan a QR code</div>
              </div>
              <div v-if="config.showPoweredBy" class="powered-by">Powered by Identity Verification</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch, type CSSProperties } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

interface PageBuilderStep {
  enabled: boolean
  label: string
}

interface PageBuilderConfig {
  headerTitle: string
  headerSubtitle: string
  showPoweredBy: boolean
  theme: 'dark' | 'light'
  backgroundColor: string
  cardBackgroundColor: string
  textColor: string
  accentColor: string
  mutedTextColor: string
  borderColor: string
  fontFamily: 'dm-sans' | 'inter' | 'system'
  steps: { front: PageBuilderStep; back: PageBuilderStep; liveness: PageBuilderStep }
  completionTitle: string
  completionMessage: string
  showConfetti: boolean
}

type ColorKey =
  | 'backgroundColor'
  | 'cardBackgroundColor'
  | 'textColor'
  | 'accentColor'
  | 'mutedTextColor'
  | 'borderColor'

type StepKey = 'front' | 'back' | 'liveness'

const DEFAULT_CONFIG: PageBuilderConfig = {
  headerTitle: 'Verify Your Identity',
  headerSubtitle: 'Complete the steps below to verify your identity',
  showPoweredBy: true,
  theme: 'dark',
  backgroundColor: '#080c14',
  cardBackgroundColor: '#0f1420',
  textColor: '#dde2ec',
  accentColor: '#00F0FF',
  mutedTextColor: '#8a8a90',
  borderColor: '#1e1e22',
  fontFamily: 'dm-sans',
  steps: {
    front: { enabled: true, label: 'Front of ID' },
    back: { enabled: true, label: 'Back of ID' },
    liveness: { enabled: true, label: 'Live Capture' },
  },
  completionTitle: 'Verification Complete',
  completionMessage: 'Your identity has been successfully verified.',
  showConfetti: false,
}

const DARK_PALETTE: Pick<PageBuilderConfig, ColorKey> = {
  backgroundColor: '#080c14',
  cardBackgroundColor: '#0f1420',
  textColor: '#dde2ec',
  accentColor: '#00F0FF',
  mutedTextColor: '#8a8a90',
  borderColor: '#1e1e22',
}

const LIGHT_PALETTE: Pick<PageBuilderConfig, ColorKey> = {
  backgroundColor: '#f8fafc',
  cardBackgroundColor: '#ffffff',
  textColor: '#1e293b',
  accentColor: '#00d4d4',
  mutedTextColor: '#6b6b70',
  borderColor: '#e4e2dc',
}

const COLOR_FIELDS: Array<{ key: ColorKey; label: string }> = [
  { key: 'backgroundColor', label: 'Background' },
  { key: 'cardBackgroundColor', label: 'Card' },
  { key: 'textColor', label: 'Text' },
  { key: 'accentColor', label: 'Accent' },
  { key: 'mutedTextColor', label: 'Muted text' },
  { key: 'borderColor', label: 'Borders' },
]

const STEP_FIELDS: Array<{ key: StepKey; label: string }> = [
  { key: 'front', label: 'Front document' },
  { key: 'back', label: 'Back document' },
  { key: 'liveness', label: 'Live capture' },
]

const HEX_RE = /^#[0-9a-fA-F]{6}$/

const auth = useAuthStore()

const config = reactive<PageBuilderConfig>(cloneDefaults())
const slug = ref('')
const savedSlug = ref('')
const slugError = ref('')
const loading = ref(false)
const saving = ref(false)
const savingSlug = ref(false)
const dirty = ref(false)
const savedAt = ref('')
const saveError = ref('')
const previewDevice = ref<'mobile' | 'desktop'>('mobile')
const previewScreen = ref<'landing' | 'complete'>('landing')
const logo = ref('')
const logoReference = ref('')
const uploadingLogo = ref(false)
let hydrating = false

function cloneDefaults(): PageBuilderConfig {
  return JSON.parse(JSON.stringify(DEFAULT_CONFIG)) as PageBuilderConfig
}

function isHex(value: string) {
  return HEX_RE.test(value)
}

function isRenderable(value: string) {
  return /^(https?:|data:|blob:|\/)/i.test(value)
}

function applyConfig(raw: unknown) {
  const source = raw && typeof raw === 'object' ? raw as Record<string, any> : {}
  const next = cloneDefaults()
  for (const key of Object.keys(DEFAULT_CONFIG) as Array<keyof PageBuilderConfig>) {
    if (key === 'steps') continue
    if (typeof source[key] === typeof DEFAULT_CONFIG[key]) {
      ;(next as unknown as Record<string, unknown>)[key] = source[key]
    }
  }
  const steps = source.steps && typeof source.steps === 'object' ? source.steps as Record<string, any> : {}
  for (const key of ['front', 'back', 'liveness'] as StepKey[]) {
    const incoming = steps[key]
    if (incoming && typeof incoming === 'object') {
      if (typeof incoming.enabled === 'boolean') next.steps[key].enabled = incoming.enabled
      if (typeof incoming.label === 'string') next.steps[key].label = incoming.label
    }
  }
  if (next.theme !== 'dark' && next.theme !== 'light') next.theme = 'dark'
  if (!['dm-sans', 'inter', 'system'].includes(next.fontFamily)) next.fontFamily = 'dm-sans'
  hydrating = true
  Object.assign(config, next)
  nextTick(() => { hydrating = false; dirty.value = false })
}

async function loadAll() {
  if (!auth.isDeveloper) {
    loading.value = false
    return
  }
  loading.value = true
  saveError.value = ''
  slugError.value = ''
  try {
    const [builder, branding] = await Promise.all([
      api.get('/developer/settings/page-builder'),
      api.get('/developer/settings/branding').catch(() => null),
    ])
    const payload = builder.data?.data || {}
    let rawConfig: unknown = payload.config
    if (typeof rawConfig === 'string') {
      try { rawConfig = JSON.parse(rawConfig || '{}') } catch { rawConfig = {} }
    }
    applyConfig(rawConfig)
    slug.value = typeof payload.slug === 'string' ? payload.slug : ''
    savedSlug.value = slug.value
    const stored = branding?.data?.data?.logo_url
    logoReference.value = typeof stored === 'string' ? stored : ''
    logo.value = isRenderable(logoReference.value) ? logoReference.value : ''
  } catch (error) {
    saveError.value = apiMessage(error)
  } finally {
    loading.value = false
  }
}

watch(config, () => {
  if (!hydrating) dirty.value = true
}, { deep: true })

async function saveConfig() {
  if (!auth.isDeveloper) {
    saveError.value = 'Connect an API key or sign in before saving.'
    return
  }
  saving.value = true
  saveError.value = ''
  try {
    await api.put('/developer/settings/page-builder', { config: { ...config } })
    dirty.value = false
    savedAt.value = new Date().toLocaleTimeString()
    Message.success('Page configuration saved')
  } catch (error) {
    saveError.value = apiMessage(error)
    Message.error(saveError.value)
  } finally {
    saving.value = false
  }
}

async function saveSlug() {
  if (!auth.isDeveloper) return
  slugError.value = ''
  savingSlug.value = true
  try {
    const payload = slug.value ? { slug: slug.value } : { slug: null }
    const { data } = await api.put('/developer/settings/page-builder/slug', payload)
    savedSlug.value = data?.slug || slug.value
    slug.value = savedSlug.value
    Message.success(savedSlug.value ? `Slug saved: /v/${savedSlug.value}` : 'Slug removed')
  } catch (error) {
    slugError.value = apiMessage(error)
    Message.error(slugError.value)
  } finally {
    savingSlug.value = false
  }
}

function onSlugInput(value: string) {
  slug.value = value.toLowerCase().replace(/[^a-z0-9-]/g, '')
  slugError.value = ''
}

async function onLogoChange(_fileList: unknown, item: { file?: File }) {
  const file = item?.file
  if (!file) return
  if (!/\.(png|jpe?g)$/i.test(file.name)) {
    Message.error('Logo must be a PNG or JPEG')
    return
  }
  if (file.size > 2 * 1024 * 1024) {
    Message.error('Logo must be under 2 MB')
    return
  }
  uploadingLogo.value = true
  saveError.value = ''
  try {
    const form = new FormData()
    form.append('file', file)
    const { data } = await api.post('/developer/branding/logo', form)
    const url = data?.data?.url || data?.data?.logo_url || data?.url
    logo.value = typeof url === 'string' ? url : URL.createObjectURL(file)
    logoReference.value = logo.value
    Message.success('Logo uploaded')
  } catch (error) {
    saveError.value = apiMessage(error)
    Message.error(saveError.value)
  } finally {
    uploadingLogo.value = false
  }
}

function resetDefaults() {
  applyConfig({})
  slug.value = savedSlug.value
  slugError.value = ''
  Message.success('Defaults restored — save to apply')
}

function applyTheme(value: string | number | boolean) {
  const palette = value === 'light' ? LIGHT_PALETTE : DARK_PALETTE
  config.theme = value === 'light' ? 'light' : 'dark'
  Object.assign(config, palette)
}

function setColor(key: ColorKey, value: string) {
  config[key] = value
}

function setStepEnabled(key: StepKey, value: boolean) {
  config.steps[key].enabled = value
}

function setStepLabel(key: StepKey, value: string) {
  config.steps[key].label = value
}

onMounted(loadAll)

const renderableLogo = computed(() => (logo.value && isRenderable(logo.value) ? logo.value : ''))

const previewConfig = computed(() => ({
  headerTitle: config.headerTitle,
  headerSubtitle: config.headerSubtitle,
  completionTitle: config.completionTitle,
  completionMessage: config.completionMessage,
}))

const previewStepLabels = computed(() => {
  const labels = (['front', 'back', 'liveness'] as StepKey[])
    .filter((key) => config.steps[key].enabled)
    .map((key) => config.steps[key].label)
  return [...labels, 'Complete']
})

function withAlpha(value: string, alpha: number) {
  const match = /^#([0-9a-f]{6})$/i.exec(value)
  if (!match) return `rgba(0, 240, 255, ${alpha})`
  const int = parseInt(match[1], 16)
  const r = (int >> 16) & 255
  const g = (int >> 8) & 255
  const b = int & 255
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

function fontStack(family: PageBuilderConfig['fontFamily']) {
  if (family === 'inter') return 'Inter, ui-sans-serif, system-ui, sans-serif'
  if (family === 'system') return "ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif"
  return "'DM Sans', Inter, ui-sans-serif, system-ui, sans-serif"
}

const previewThemeVars = computed<CSSProperties>(() => ({
  '--paper': config.backgroundColor,
  '--panel': config.cardBackgroundColor,
  '--ink': config.textColor,
  '--mid': config.mutedTextColor,
  '--rule': config.borderColor,
  '--accent': config.accentColor,
  '--accent-soft': withAlpha(config.accentColor, 0.14),
  '--sans': fontStack(config.fontFamily),
}))
</script>

<style scoped>
.builder { padding: 0; }
.builder-bar { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; max-width: 1320px; margin: 0 auto; padding: 20px 28px 14px; }
.builder-title { display: flex; align-items: center; gap: 12px; }
.builder-title h2 { margin: 2px 0 0; font-size: 20px; }
.back-link { font-family: ui-monospace, Menlo, monospace; color: #6b6b70; text-decoration: none; font-size: 16px; }
.builder-alert { max-width: 1264px; margin: 0 28px 12px; }
.builder-body { display: grid; grid-template-columns: minmax(0, 55fr) minmax(0, 45fr); gap: 18px; max-width: 1320px; margin: 0 auto; padding: 0 28px 48px; align-items: start; }
.panel-controls :deep(.arco-card-body) { padding: 6px 18px 18px; }
.section { padding: 14px 0; border-bottom: 1px solid #eef0f3; }
.section:last-child { border-bottom: none; }
.section-title { font-family: ui-monospace, Menlo, monospace; font-size: 11px; font-weight: 600; letter-spacing: .08em; text-transform: uppercase; color: #2563eb; margin-bottom: 10px; }
.logo-row { display: flex; gap: 14px; align-items: flex-start; }
.logo-box { width: 56px; height: 56px; flex-shrink: 0; border: 1px solid #e4e2dc; background: #fff; display: flex; align-items: center; justify-content: center; overflow: hidden; font-size: 11px; }
.logo-box img { max-width: 100%; max-height: 100%; object-fit: contain; }
.logo-actions { display: flex; flex-direction: column; gap: 6px; }
.logo-actions small { font-size: 11px; line-height: 1.5; }
.theme-group { margin-bottom: 12px; }
.color-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px 10px; }
.color-item { margin-bottom: 0; }
.color-item :deep(.arco-form-item-label) { font-size: 11px; color: #6b6b70; }
.color-input { width: 96px; }
.color-swatch { width: 28px; height: 28px; border: 1px solid #e4e2dc; padding: 0; background: transparent; cursor: pointer; }
.step-row { display: flex; align-items: center; gap: 12px; padding: 8px 10px; border: 1px solid #e4e2dc; background: #fafbfc; margin-bottom: 8px; }
.step-label-input { flex: 1; min-width: 0; }
.switch-row { display: flex; align-items: center; gap: 10px; font-size: 13px; }
.slug-row { display: flex; align-items: center; gap: 8px; width: 100%; }
.slug-row > .arco-input-outer { flex: 1; min-width: 0; }
.mono { font-family: ui-monospace, Menlo, monospace; }
.slug-error { margin-top: 6px; font-size: 12px; color: #cb2634; }
.slug-ok { margin-top: 6px; font-size: 12px; color: #00a35c; }
.panel-preview { background: #fff; border: 1px solid #e4e2dc; padding: 18px; display: flex; flex-direction: column; align-items: center; gap: 16px; position: sticky; top: 88px; }
.preview-toggles { justify-content: center; }
.preview-toggles .divider { width: 1px; height: 20px; background: #e4e2dc; display: inline-block; }
.device-frame { background: #000; padding: 10px; width: 300px; max-width: 100%; border-radius: 36px; box-shadow: 0 22px 50px rgba(15, 23, 42, .28); transition: width .2s; }
.device-frame.desktop { border-radius: 12px; padding: 8px; width: 100%; max-width: 440px; }
.device-screen { border-radius: 28px; overflow: hidden; min-height: 520px; background: var(--paper); color: var(--ink); font-family: var(--sans); }
.device-frame.desktop .device-screen { border-radius: 6px; min-height: 400px; }
.screen-mobile { padding-bottom: 18px; }
.mobile-status { display: flex; justify-content: space-between; padding: 12px 18px 8px; font-family: ui-monospace, Menlo, monospace; font-size: 10px; color: var(--mid); }
.mobile-status .accent { color: var(--accent); letter-spacing: .1em; text-transform: uppercase; }
.mobile-logo { height: 22px; max-width: 60%; object-fit: contain; display: block; margin: 8px auto 0; }
.mobile-steps { display: flex; gap: 4px; padding: 10px 18px 4px; }
.mobile-step { flex: 1; min-width: 0; }
.mobile-step i { display: block; height: 3px; border-radius: 2px; background: var(--rule); }
.mobile-step i.active { background: var(--accent); }
.mobile-step small { display: block; margin-top: 4px; font-family: ui-monospace, Menlo, monospace; font-size: 8px; text-transform: uppercase; color: var(--mid); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.mobile-step small.active { color: var(--ink); }
.mobile-copy { padding: 14px 18px 0; }
.step-of { font-family: ui-monospace, Menlo, monospace; font-size: 9px; letter-spacing: .08em; color: var(--accent); margin-bottom: 8px; }
.mobile-copy h3 { font-size: 21px; font-weight: 600; margin: 0 0 6px; line-height: 1.15; color: var(--ink); }
.mobile-copy p { font-size: 12px; margin: 0; color: var(--mid); }
.camera-placeholder { margin: 16px 18px; padding: 28px 16px; text-align: center; background: var(--panel); border: 2px solid var(--accent); border-radius: 12px; font-size: 11px; color: var(--mid); font-family: ui-monospace, Menlo, monospace; }
.cta { margin: 0 18px; padding: 13px; text-align: center; background: var(--accent); color: var(--paper); border-radius: 10px; font-family: ui-monospace, Menlo, monospace; font-size: 12px; font-weight: 600; letter-spacing: .04em; text-transform: uppercase; }
.powered-by { text-align: center; padding-top: 16px; font-size: 10px; color: var(--mid); }
.screen-desktop { padding: 34px 26px; text-align: center; }
.desktop-logo { height: 30px; max-width: 55%; object-fit: contain; display: block; margin: 0 auto 18px; }
.screen-desktop h3 { font-size: 23px; font-weight: 600; margin: 0 0 8px; color: var(--ink); }
.screen-desktop p { font-size: 13px; margin: 0 0 22px; color: var(--mid); }
.desktop-methods { display: flex; gap: 12px; }
.method { flex: 1; padding: 20px 14px; font-size: 13px; font-weight: 600; color: var(--ink); background: var(--panel); border: 1.5px solid var(--rule); border-radius: 12px; }
.method.primary { border-color: var(--accent); }
.screen-complete { padding: 40px 24px; text-align: center; }
.complete-mark { width: 54px; height: 54px; margin: 0 auto 14px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 24px; background: var(--accent-soft); color: var(--accent); border: 1px solid var(--accent); }
.screen-complete h3 { font-size: 21px; margin: 0 0 6px; color: var(--ink); }
.screen-complete p { font-size: 13px; color: var(--mid); margin: 0 0 20px; }
.complete-card { text-align: left; background: var(--panel); border: 1px solid var(--rule); border-radius: 10px; padding: 14px 16px; }
.complete-row { display: flex; justify-content: space-between; font-size: 12px; padding: 4px 0; color: var(--mid); }
.complete-row strong { color: var(--ink); }
.complete-note { margin-top: 14px; font-size: 10px; font-family: ui-monospace, Menlo, monospace; color: var(--mid); }
@media (max-width: 1080px) {
  .builder-body { grid-template-columns: 1fr; }
  .panel-preview { position: static; }
  .color-grid { grid-template-columns: 1fr 1fr; }
  .builder-alert { margin: 0 16px 12px; }
  .builder-bar, .builder-body { padding-left: 16px; padding-right: 16px; }
}
</style>
