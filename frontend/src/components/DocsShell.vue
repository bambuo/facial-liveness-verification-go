<template>
  <div class="docs-shell">
    <!-- Header bar -->
    <div class="docs-header">
      <div class="docs-brand">
        <span class="docs-brand-accent">identity</span>
        <span class="docs-brand-dim"> / </span>
        <span>{{ slug }}</span>
      </div>

      <div ref="searchRef" class="docs-search">
        <a-input
          ref="searchInputRef"
          v-model="searchQuery"
          placeholder="Search docs..."
          class="docs-search-input"
          @focus="onSearchFocus"
          @input="onSearchInput"
        >
          <template #prefix><icon-search /></template>
          <template #suffix><span class="docs-kbd">{{ shortcutLabel }}</span></template>
        </a-input>

        <div v-if="searchOpen && searchResults.length" class="docs-search-panel">
          <div v-for="group in groupedResults" :key="group.path">
            <div class="docs-search-group">{{ group.page }}</div>
            <button
              v-for="item in group.items"
              :key="item.path + '#' + item.id"
              type="button"
              class="docs-search-item"
              @click="selectResult(item)"
            >{{ item.section }}</button>
          </div>
        </div>
      </div>

      <div class="docs-header-actions">
        <router-link to="/docs/markdown" class="docs-chip">.md</router-link>
        <a-tag size="small" color="green">v1.7.0</a-tag>
        <a-tag size="small">March 2026</a-tag>
      </div>
    </div>

    <!-- Two-column layout -->
    <div class="docs-body">
      <aside class="docs-aside">
        <div class="docs-aside-label">Pages</div>
        <a-menu
          :selected-keys="[route.path]"
          class="docs-page-menu"
          @menu-item-click="goToPage"
        >
          <a-menu-item v-for="page in DOC_PAGES" :key="page.path">{{ page.label }}</a-menu-item>
        </a-menu>

        <template v-if="nav.length">
          <div class="docs-aside-rule" />
          <div class="docs-aside-label">On this page</div>
          <a-anchor :change-hash="false" :affix="false" :line-less="true" :boundary="80" :target-offset="120" class="docs-toc">
            <a-anchor-link
              v-for="item in nav"
              :key="item.id"
              :href="'#' + item.id"
              :title="item.label"
              :class="item.depth === 1 ? 'docs-toc-sub' : 'docs-toc-top'"
            />
          </a-anchor>
        </template>
      </aside>

      <main class="docs-main">
        <slot />
        <div class="docs-footer">&copy; 2026 Identity Verification &mdash; Open source under MIT License</div>
      </main>
    </div>
  </div>
</template>

<script lang="ts">
/**
 * Shared documentation atoms, ported 1:1 from the reference frontend
 * (components/docs/shared.tsx + components/docs/DocLayout.tsx) onto Arco
 * components. Exported from this file so every /docs/* page can reuse them
 * without adding new files.
 */
import { computed, defineComponent, h, ref, type PropType } from 'vue'
import {
  Alert,
  Button,
  Card,
  Divider,
  Message,
  Step,
  Steps,
  TabPane,
  Tabs,
  Tag,
  TypographyParagraph,
  TypographyTitle,
} from '@arco-design/web-vue'

export interface DocsNavItem {
  id: string
  label: string
  depth: number
}

export interface DocsPageLink {
  path: string
  label: string
}

/** Page-level navigation shown at the top of every docs sidebar. */
export const DOC_PAGES: DocsPageLink[] = [
  { path: '/docs', label: 'Getting Started' },
  { path: '/docs/guides', label: 'Guides' },
  { path: '/docs/sdk', label: 'SDK & Embed' },
  { path: '/docs/features', label: 'Features' },
  { path: '/docs/reference', label: 'Reference' },
  { path: '/docs/review', label: 'Review Dashboard' },
]

/** Search index (all sections across all docs pages). */
export interface DocsSearchItem {
  page: string
  path: string
  section: string
  id: string
}

export const SEARCH_INDEX: DocsSearchItem[] = [
  { page: 'Getting Started', path: '/docs', section: 'Quick Start', id: 'quick-start' },
  { page: 'Getting Started', path: '/docs', section: 'Authentication', id: 'auth' },
  { page: 'Getting Started', path: '/docs', section: 'Verification Flow', id: 'flow' },
  { page: 'Getting Started', path: '/docs', section: '1 - Start Session', id: 'step-1' },
  { page: 'Getting Started', path: '/docs', section: '2 - Upload Front', id: 'step-2' },
  { page: 'Getting Started', path: '/docs', section: '3 - Upload Back', id: 'step-3' },
  { page: 'Getting Started', path: '/docs', section: '4 - Live Capture', id: 'step-4' },
  { page: 'Getting Started', path: '/docs', section: '5 - Get Results', id: 'step-5' },
  { page: 'Getting Started', path: '/docs', section: 'Cross-Validation', id: 'selfie' },
  { page: 'Getting Started', path: '/docs', section: 'Get Selfie URL', id: 'selfie-url' },
  { page: 'Guides', path: '/docs/guides', section: 'Integration Options', id: 'integration' },
  { page: 'Guides', path: '/docs/guides', section: 'Guides & Tutorials', id: 'guides' },
  { page: 'Guides', path: '/docs/guides', section: 'End-to-End Tutorial', id: 'guide-e2e' },
  { page: 'Guides', path: '/docs/guides', section: 'Mobile Handoff', id: 'guide-mobile' },
  { page: 'Guides', path: '/docs/guides', section: 'Building Custom UI', id: 'guide-custom-ui' },
  { page: 'Guides', path: '/docs/guides', section: 'Self-Hosting', id: 'self-hosting' },
  { page: 'Guides', path: '/docs/guides', section: 'Prerequisites', id: 'sh-prerequisites' },
  { page: 'Guides', path: '/docs/guides', section: 'Install', id: 'sh-install' },
  { page: 'Guides', path: '/docs/guides', section: 'External Database', id: 'sh-external-db' },
  { page: 'Guides', path: '/docs/guides', section: 'Useful Commands', id: 'sh-commands' },
  { page: 'SDK & Embed', path: '/docs/sdk', section: 'JavaScript SDK', id: 'sdk' },
  { page: 'SDK & Embed', path: '/docs/sdk', section: 'Embed Component', id: 'embed' },
  { page: 'Features', path: '/docs/features', section: 'Analysis Engine', id: 'analysis' },
  { page: 'Features', path: '/docs/features', section: 'Statuses', id: 'statuses' },
  { page: 'Features', path: '/docs/features', section: 'Batch API', id: 'batch' },
  { page: 'Features', path: '/docs/features', section: 'Address Verification', id: 'address' },
  { page: 'Features', path: '/docs/features', section: 'AML / Sanctions', id: 'aml' },
  { page: 'Features', path: '/docs/features', section: 'Age Estimation', id: 'age-estimation' },
  { page: 'Features', path: '/docs/features', section: 'Velocity Checks', id: 'velocity' },
  { page: 'Features', path: '/docs/features', section: 'IP Geolocation', id: 'ip-geolocation' },
  { page: 'Features', path: '/docs/features', section: 'Voice Auth', id: 'voice-auth' },
  { page: 'Features', path: '/docs/features', section: 'Compliance Rules', id: 'compliance' },
  { page: 'Features', path: '/docs/features', section: 'Monitoring', id: 'monitoring' },
  { page: 'Reference', path: '/docs/reference', section: 'Rate Limits', id: 'rate-limits' },
  { page: 'Reference', path: '/docs/reference', section: 'Changelog', id: 'changelog' },
  { page: 'Reference', path: '/docs/reference', section: 'Support', id: 'support' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Overview', id: 'overview' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Getting Access', id: 'access' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Roles', id: 'roles' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Invite Team Members', id: 'reviewers' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'OTP Login', id: 'reviewer-login' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'The Dashboard', id: 'dashboard' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Stats Bar', id: 'stats' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Verification Table', id: 'table' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Filters & Search', id: 'filters' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Detail Panel', id: 'detail' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Review Actions', id: 'actions' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Approve', id: 'approve' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Reject', id: 'reject' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Override', id: 'override' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Status Reference', id: 'statuses' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Webhooks', id: 'webhooks' },
]

/** Documentation API/site origin (self-hosted deployments serve both on one origin). */
export const docsApiUrl = (): string =>
  typeof window !== 'undefined' ? window.location.origin : 'https://api.example.com'

export const docsSiteUrl = (): string =>
  typeof window !== 'undefined' ? window.location.origin : 'https://example.com'

/** Section anchor target (matches reference scrollMarginTop: 88). */
export const DocsSectionAnchor = defineComponent({
  name: 'DocsSectionAnchor',
  props: { id: { type: String, required: true } },
  setup(props) {
    return () => h('div', { id: props.id, class: 'docs-anchor' })
  },
})

export const DocsH2 = defineComponent({
  name: 'DocsH2',
  props: { index: { type: String, default: '' } },
  setup(props, { slots }) {
    return () =>
      h('div', { class: 'docs-h2' }, [
        props.index ? h('span', { class: 'docs-h2-index' }, props.index) : null,
        h(TypographyTitle, { heading: 2, style: 'margin: 0; font-size: 22px' }, () => slots.default?.()),
      ])
  },
})

export const DocsH3 = defineComponent({
  name: 'DocsH3',
  setup(_props, { slots }) {
    return () => h('h3', { class: 'docs-h3' }, ['-- ', slots.default?.()])
  },
})

export const DocsH4 = defineComponent({
  name: 'DocsH4',
  setup(_props, { slots }) {
    return () => h(TypographyTitle, { heading: 4, class: 'docs-h4' }, () => slots.default?.())
  },
})

export const DocsLead = defineComponent({
  name: 'DocsLead',
  setup(_props, { slots }) {
    return () => h(TypographyParagraph, { class: 'docs-lead' }, () => slots.default?.())
  },
})

export const DocsDivider = defineComponent({
  name: 'DocsDivider',
  setup() {
    return () => h(Divider, { class: 'docs-divider' })
  },
})

export const DocsPill = defineComponent({
  name: 'DocsPill',
  props: {
    color: { type: String, default: 'gray' },
    label: { type: String, default: '' },
  },
  setup(props, { slots }) {
    return () => h(Tag, { size: 'small', color: props.color }, () => slots.default?.() ?? props.label)
  },
})

const STATUS_COLORS: Record<string, string> = {
  pending: 'orange',
  processing: 'arcoblue',
  verified: 'green',
  failed: 'red',
  manual_review: 'orangered',
  AWAITING_FRONT: 'orange',
  AWAITING_BACK: 'orange',
  CROSS_VALIDATING: 'arcoblue',
  AWAITING_LIVE: 'orange',
  FACE_MATCHING: 'arcoblue',
  COMPLETE: 'green',
  HARD_REJECTED: 'red',
}

export const DocsStatusPill = defineComponent({
  name: 'DocsStatusPill',
  props: { status: { type: String, required: true } },
  setup(props) {
    return () => h(Tag, { size: 'small', color: STATUS_COLORS[props.status] ?? 'gray' }, () => props.status)
  },
})

const CALLOUT_META: Record<string, { type: 'info' | 'warning' | 'error' | 'success'; title: string }> = {
  note: { type: 'info', title: 'Note' },
  warning: { type: 'warning', title: 'Important' },
  danger: { type: 'error', title: 'Warning' },
  tip: { type: 'success', title: 'Tip' },
}

export const DocsCallout = defineComponent({
  name: 'DocsCallout',
  props: {
    type: { type: String as PropType<'note' | 'warning' | 'danger' | 'tip'>, default: 'note' },
  },
  setup(props, { slots }) {
    return () => {
      const meta = CALLOUT_META[props.type] ?? CALLOUT_META.note
      return h(
        Alert,
        { type: meta.type, title: meta.title, class: 'docs-callout' },
        () => h('div', { class: 'docs-callout-body' }, slots.default?.()),
      )
    }
  },
})

/** Code panel with a label bar and a Copy action. */
export const DocsCode = defineComponent({
  name: 'DocsCode',
  props: {
    code: { type: String, required: true },
    label: { type: String, default: '' },
  },
  setup(props) {
    const copy = async () => {
      try {
        await navigator.clipboard.writeText(props.code)
        Message.success('Copied to clipboard')
      } catch {
        Message.error('Copy failed')
      }
    }
    return () =>
      h('div', { class: 'docs-code' }, [
        h('div', { class: 'docs-code-head' }, [
          h('span', { class: 'docs-code-label' }, props.label || 'Code'),
          h(
            Button,
            { size: 'mini', type: 'text', class: 'docs-code-copy', onClick: copy },
            () => 'Copy',
          ),
        ]),
        h('pre', { class: 'docs-code-pre' }, props.code),
      ])
  },
})

const TAB_TITLES: Record<string, string> = { curl: 'cURL', js: 'JavaScript', python: 'Python' }

/** Tabbed code block. `v-model` keeps every block on the page in sync. */
export const DocsCodeTabs = defineComponent({
  name: 'DocsCodeTabs',
  props: {
    modelValue: { type: String, default: 'curl' },
    curl: { type: String, default: '' },
    js: { type: String, default: '' },
    python: { type: String, default: '' },
  },
  emits: { 'update:modelValue': (_key: string) => true },
  setup(props, { emit }) {
    const tabs = computed(() => (props.curl ? ['curl', 'js', 'python'] : ['js', 'python']))
    return () => {
      const active = tabs.value.includes(props.modelValue) ? props.modelValue : tabs.value[0]
      return h(
        Tabs,
        {
          activeKey: active,
          type: 'line',
          size: 'small',
          class: 'docs-code-tabs',
          onChange: (key: string | number) => emit('update:modelValue', String(key)),
        },
        () =>
          tabs.value.map((key) =>
            h(TabPane, { key, title: TAB_TITLES[key] }, () =>
              h(DocsCode, { code: (props as unknown as Record<string, string>)[key], label: TAB_TITLES[key] }),
            ),
          ),
      )
    }
  },
})

/** Field row used in endpoint / option tables. */
export const DocsFieldRow = defineComponent({
  name: 'DocsFieldRow',
  props: {
    name: { type: String, required: true },
    type: { type: String, default: '' },
    req: { type: Boolean, default: false },
    desc: { type: String, default: '' },
  },
  setup(props) {
    return () =>
      h('div', { class: 'docs-field' }, [
        h('code', { class: 'docs-field-name' }, props.name),
        h('code', { class: 'docs-field-type' }, props.type),
        h(Tag, { size: 'small', color: props.req ? 'red' : 'gray' }, () => (props.req ? 'required' : 'optional')),
        h('span', { class: 'docs-field-desc' }, props.desc),
      ])
  },
})

/** Endpoint card: step badge, method, path, title, optional badge. */
export const DocsEndpointCard = defineComponent({
  name: 'DocsEndpointCard',
  props: {
    step: { type: Number, default: 0 },
    method: { type: String, default: 'POST' },
    path: { type: String, default: '' },
    title: { type: String, default: '' },
    badge: { type: String, default: '' },
    badgeColor: { type: String, default: 'gray' },
  },
  setup(props, { slots }) {
    return () =>
      h(
        Card,
        { class: 'docs-endpoint', bordered: true },
        {
          title: () =>
            h('div', { class: 'docs-endpoint-head' }, [
              props.step ? h('span', { class: 'docs-step' }, String(props.step)) : null,
              h(Tag, { size: 'small', color: props.method === 'GET' ? 'green' : 'arcoblue' }, () => props.method),
              h('code', { class: 'docs-endpoint-path' }, props.path),
              props.badge ? h(Tag, { size: 'small', color: props.badgeColor }, () => props.badge) : null,
              h('span', { class: 'docs-endpoint-title' }, props.title),
            ]),
          default: () => h('div', { class: 'docs-endpoint-body' }, slots.default?.()),
        },
      )
  },
})

/** Five-step verification pipeline diagram. */
export const DocsPipeline = defineComponent({
  name: 'DocsPipeline',
  setup() {
    const steps = [
      { title: 'Start Session', description: 'initialize' },
      { title: 'Front Doc', description: 'OCR' },
      { title: 'Back-of-ID', description: 'cross-validation' },
      { title: 'Live Capture', description: 'liveness + match' },
      { title: 'Results', description: 'status' },
    ]
    return () =>
      h(Card, { class: 'docs-pipeline' }, () => [
        h('div', { class: 'docs-eyebrow' }, 'Verification Pipeline'),
        h('div', { class: 'docs-pipeline-steps' }, [
          h(Steps, { current: steps.length, size: 'small' }, () =>
            steps.map((step) =>
              h(Step, { key: step.title, title: step.title, description: step.description, status: 'wait' as const }),
            ),
          ),
        ]),
        h(
          Alert,
          { type: 'warning', class: 'docs-pipeline-gate', title: 'Cross-validation gate:' },
          () =>
            'If Step 3 (back-of-ID) fails cross-validation, status is set to failed and Step 4 is blocked.',
        ),
      ])
  },
})
</script>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const props = defineProps<{
  slug: string
  nav: DocsNavItem[]
}>()

const route = useRoute()
const router = useRouter()

const searchRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<HTMLElement | null>(null)
const searchQuery = ref('')
const searchOpen = ref(false)

const shortcutLabel =
  typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.userAgent) ? '⌘K' : 'Ctrl+K'

const searchResults = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return []
  return SEARCH_INDEX.filter(
    (item) => item.section.toLowerCase().includes(query) || item.page.toLowerCase().includes(query),
  ).slice(0, 12)
})

const groupedResults = computed(() => {
  const groups: { page: string; path: string; items: DocsSearchItem[] }[] = []
  for (const item of searchResults.value) {
    let group = groups.find((entry) => entry.path === item.path)
    if (!group) {
      group = { page: item.page, path: item.path, items: [] }
      groups.push(group)
    }
    group.items.push(item)
  }
  return groups
})

function onSearchInput() {
  searchOpen.value = true
}

function onSearchFocus() {
  if (searchQuery.value.trim()) searchOpen.value = true
}

function selectResult(item: DocsSearchItem) {
  searchQuery.value = ''
  searchOpen.value = false
  if (route.path === item.path) {
    document.getElementById(item.id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  } else {
    router.push({ path: item.path, hash: `#${item.id}` })
  }
}

function goToPage(key: string) {
  router.push(key)
}

function scrollToHash() {
  const hash = route.hash.replace(/^#/, '')
  if (!hash) return
  window.setTimeout(() => {
    document.getElementById(hash)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, 150)
}

function onKeydown(event: KeyboardEvent) {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    const input = (searchInputRef.value as unknown as { focus?: () => void } | null) ?? null
    input?.focus?.()
    searchOpen.value = true
  }
  if (event.key === 'Escape') {
    searchOpen.value = false
  }
}

function onPointerDown(event: MouseEvent) {
  if (!searchOpen.value) return
  const target = event.target as Node | null
  if (searchRef.value && target && !searchRef.value.contains(target)) searchOpen.value = false
}

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
  document.addEventListener('mousedown', onPointerDown)
  scrollToHash()
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
  document.removeEventListener('mousedown', onPointerDown)
})

watch(() => route.hash, scrollToHash)
watch(
  () => props.slug,
  () => {
    searchQuery.value = ''
    searchOpen.value = false
  },
)
</script>

<style>
/* Docs shell + shared docs atoms (prefix .docs- only, owned by this file). */
.docs-shell {
  background: var(--paper, #f7f6f3);
  color: var(--ink, #0b0b0d);
  min-height: 100vh;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif;
}
.docs-header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 32px;
  background: var(--paper, #f7f6f3);
  border-bottom: 1px solid var(--rule, #e4e2dc);
}
.docs-brand {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 15px;
  font-weight: 600;
  white-space: nowrap;
}
.docs-brand-accent { color: #0e7490; }
.docs-brand-dim { color: #9a9a9f; }
.docs-search { position: relative; flex: 0 1 280px; }
.docs-search-input .arco-input-prefix { color: #9a9a9f; }
.docs-kbd {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10px;
  color: #9a9a9f;
  pointer-events: none;
}
.docs-search-panel {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 4px;
  max-height: 320px;
  overflow-y: auto;
  background: #fff;
  border: 1px solid var(--rule, #e4e2dc);
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.14);
  z-index: 50;
}
.docs-search-group {
  padding: 8px 12px 4px;
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: #9a9a9f;
}
.docs-search-item {
  display: block;
  width: 100%;
  padding: 6px 12px 6px 20px;
  border: none;
  background: transparent;
  text-align: left;
  cursor: pointer;
  font-size: 13px;
  color: #475569;
}
.docs-search-item:hover { background: #f1f5f9; color: #0e7490; }
.docs-header-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
}
.docs-chip {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 11.5px;
  color: #64748b;
  text-decoration: none;
  padding: 3px 10px;
  border: 1px solid var(--rule, #e4e2dc);
  background: #fff;
}
.docs-chip:hover { color: #0e7490; border-color: #0e7490; }
.docs-body { display: flex; max-width: 1440px; margin: 0 auto; align-items: flex-start; }
.docs-aside {
  width: 230px;
  flex-shrink: 0;
  position: sticky;
  top: 57px;
  height: calc(100vh - 57px);
  overflow-y: auto;
  border-right: 1px solid var(--rule, #e4e2dc);
  padding: 18px 0 28px;
  background: #fff;
}
.docs-aside-label {
  padding: 0 16px 6px;
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: #9a9a9f;
}
.docs-aside-rule { height: 1px; background: var(--rule, #e4e2dc); margin: 12px 16px; }
.docs-page-menu { background: transparent; }
.docs-page-menu .arco-menu-inner { padding: 0; }
.docs-page-menu .arco-menu-item {
  height: 30px;
  line-height: 30px;
  margin: 0;
  border-radius: 0;
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
}
.docs-page-menu .arco-menu-selected {
  color: #0e7490;
  background: #ecfeff !important;
  box-shadow: inset 2px 0 0 #0e7490;
}
.docs-toc .arco-anchor-line,
.docs-toc .arco-anchor-line-less > .arco-anchor-line { display: none; }
.docs-toc .arco-anchor-link {
  display: block;
  padding: 3px 0;
  font-size: 12.5px;
  line-height: 1.5;
  color: #64748b;
  white-space: normal;
  overflow: visible;
  text-overflow: clip;
  word-break: break-word;
}
.docs-toc .arco-anchor-link:hover { color: #0e7490; background: transparent; }
.docs-toc .arco-anchor-link-active {
  color: #0e7490;
  font-weight: 600;
  background: #ecfeff;
  box-shadow: inset 2px 0 0 #0e7490;
}
.docs-toc .arco-anchor-link.docs-toc-sub { padding-left: 12px; }
.docs-toc .arco-anchor-link.docs-toc-top { font-weight: 500; }
.docs-main { flex: 1; min-width: 0; max-width: 980px; padding: 40px 52px 72px; }
.docs-footer {
  margin-top: 64px;
  padding-top: 20px;
  border-top: 1px solid var(--rule, #e4e2dc);
  font-family: ui-monospace, Menlo, monospace;
  font-size: 11px;
  color: #9a9a9f;
}
.docs-anchor { scroll-margin-top: 88px; }
.docs-h2 { display: flex; align-items: baseline; gap: 16px; padding: 28px 0 8px; }
.docs-h2-index {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 11px;
  letter-spacing: 0.1em;
  color: #94a3b8;
  text-transform: uppercase;
}
.docs-h3 { display: flex; align-items: center; gap: 8px; margin: 30px 0 8px; font-size: 16px; font-weight: 700; }
.docs-h3 > :last-child { font-size: 16px; }
.docs-h4 { margin: 22px 0 10px; }
.docs-lead { font-size: 14.5px; color: #475569; line-height: 1.75; max-width: 680px; }
.docs-divider { margin: 40px 0; }
.docs-eyebrow {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10.5px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: #64748b;
  margin-bottom: 16px;
}
.docs-callout { margin: 14px 0; }
.docs-callout-body { font-size: 13px; line-height: 1.7; }
.docs-callout-body code {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 12px;
  background: #f1f5f9;
  padding: 1px 4px;
}
.docs-code { border: 1px solid #1e293b; overflow: hidden; margin: 14px 0; background: #0f172a; }
.docs-code-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 5px 14px;
  background: #1e293b;
  border-bottom: 1px solid #334155;
}
.docs-code-label {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10.5px;
  letter-spacing: 0.07em;
  text-transform: uppercase;
  color: #cbd5e1;
}
.docs-code-copy.arco-btn-text { color: #cbd5e1; }
.docs-code-copy.arco-btn-text:hover { color: #fff; background: #334155; }
.docs-code-pre {
  margin: 0;
  padding: 14px 16px;
  overflow: auto;
  background: #0f172a;
  color: #e2e8f0;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12.5px;
  line-height: 1.7;
}
.docs-code-tabs { margin: 14px 0; }
.docs-field {
  display: flex;
  gap: 12px;
  padding: 9px 0;
  border-top: 1px solid #eef2f7;
  flex-wrap: wrap;
  align-items: flex-start;
}
.docs-field-name {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 12.5px;
  color: #0e7490;
  width: 170px;
  flex-shrink: 0;
  overflow-wrap: anywhere;
}
.docs-field-type {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 11.5px;
  color: #64748b;
  width: 110px;
  flex-shrink: 0;
}
.docs-field-desc { font-size: 13px; color: #334155; flex: 1; line-height: 1.6; }
.docs-endpoint { margin: 20px 0; }
.docs-endpoint .arco-card-header { padding: 12px 20px; }
.docs-endpoint-body { padding: 4px 0 0; }
.docs-endpoint-head { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.docs-endpoint-path { font-family: ui-monospace, Menlo, monospace; font-size: 12.5px; color: #0e7490; }
.docs-endpoint-title { font-size: 13.5px; font-weight: 600; }
.docs-step {
  width: 24px;
  height: 24px;
  display: grid;
  place-items: center;
  border-radius: 6px;
  background: #ecfeff;
  border: 1px solid #bae6fd;
  color: #0e7490;
  font-family: ui-monospace, Menlo, monospace;
  font-size: 11px;
  font-weight: 700;
}
.docs-pipeline { margin: 20px 0; }
.docs-pipeline-gate { margin-top: 16px; }
.docs-pipeline-steps { overflow-x: auto; padding-bottom: 4px; }
.docs-pipeline-steps .arco-steps { min-width: 600px; }
.docs-muted { color: #64748b; }
.docs-body-text { font-size: 13.5px; color: #475569; line-height: 1.7; margin-bottom: 14px; }
.docs-body-narrow { max-width: 680px; }
.docs-card-note { font-size: 12px; color: #64748b; }

/* Shared docs content atoms used across several /docs pages. */
.docs-inline-code {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 12px;
  color: #0e7490;
  background: #ecfeff;
  padding: 1px 5px;
}
.docs-inline-link { color: #0e7490; font-family: ui-monospace, Menlo, monospace; font-size: 12.5px; }
.docs-success { color: #15803d; }
.docs-danger { color: #b91c1c; }
.docs-number { color: #0e7490; font-family: ui-monospace, Menlo, monospace; font-size: 13px; font-weight: 400; }
.docs-info-card { margin: 14px 0; background: #fbfcfe; }
.docs-info-card .arco-card-body { padding: 14px 18px; }
.docs-info-label {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10.5px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #64748b;
  margin-bottom: 8px;
}
.docs-info-grid { display: flex; gap: 28px; flex-wrap: wrap; }
.docs-info-value { font-family: ui-monospace, Menlo, monospace; font-size: 13px; color: #0e7490; }
.docs-field-list { margin-bottom: 8px; }
.docs-check-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 6px 24px; }
.docs-check-item { display: flex; gap: 8px; align-items: center; font-size: 12.5px; color: #475569; }
.docs-check-mark { color: #15803d; font-size: 11px; }
.docs-table-card { margin: 14px 0; background: #fbfcfe; }
.docs-table-card .arco-card-body { padding: 14px 18px; }
.docs-table-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #0e7490; white-space: nowrap; }
.docs-table-return { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #64748b; white-space: nowrap; }
.docs-table-payload { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #64748b; white-space: normal; }
.docs-table-next { font-size: 12px; color: #15803d; }
.docs-mode-row { margin: 16px 0 18px; }
.docs-mode-card { height: 100%; }
.docs-mode-card .arco-card-body { display: flex; flex-direction: column; gap: 10px; padding: 18px 20px; }
.docs-mode-head { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.docs-mode-title { font-family: ui-monospace, Menlo, monospace; font-size: 13px; font-weight: 700; }
.docs-mode-desc { margin: 0; font-size: 12.5px; color: #475569; line-height: 1.6; }
.docs-mode-signals { display: flex; flex-direction: column; gap: 3px; }
.docs-mode-signal { display: flex; gap: 8px; font-size: 12px; color: #64748b; }
.docs-mode-foot { margin-top: auto; padding-top: 8px; border-top: 1px solid #eef2f7; }
.docs-mode-strong { font-size: 12.5px; font-weight: 600; }
.docs-mode-best { font-size: 12px; color: #64748b; }
.docs-client-grid { display: grid; grid-template-columns: 90px 1fr 1fr; gap: 8px 16px; }
.docs-client-cell { font-size: 12.5px; color: #475569; line-height: 1.5; }
.docs-resource { height: 100%; }
.docs-resource .arco-card-body { display: flex; flex-direction: column; gap: 6px; padding: 18px; }
.docs-resource-icon { font-size: 22px; line-height: 1.2; }
.docs-resource-title { font-size: 14px; font-weight: 600; }
.docs-resource-desc { font-size: 12px; color: #64748b; line-height: 1.55; flex: 1; }
.docs-resource-cta { font-size: 12.5px; font-weight: 600; }
.docs-numbered-list { display: grid; grid-template-columns: auto 1fr; gap: 6px 16px; }
.docs-list-number { font-family: ui-monospace, Menlo, monospace; font-size: 12px; font-weight: 600; color: #0e7490; }
.docs-list-text { font-size: 12.5px; color: #475569; line-height: 1.6; }
.docs-ticks { display: flex; flex-direction: column; gap: 6px; }
.docs-tick { display: flex; gap: 8px; align-items: flex-start; font-size: 12.5px; color: #475569; line-height: 1.6; }
.docs-diamond { color: #0e7490; font-size: 10px; margin-top: 3px; }
.docs-option-row { margin-bottom: 20px; }
.docs-option-emoji { font-size: 20px; }
.docs-option-title { font-size: 14px; font-weight: 700; }
.docs-strategy-card .arco-card-body { padding: 16px 18px; display: flex; flex-direction: column; gap: 4px; }
.docs-strategy-title { font-family: ui-monospace, Menlo, monospace; font-size: 12.5px; font-weight: 600; margin-bottom: 4px; }
.docs-plus { color: #15803d; }
.docs-minus { color: #b91c1c; }
.docs-strategy-code { margin-top: 6px; font-family: ui-monospace, Menlo, monospace; font-size: 11px; color: #94a3b8; }
.docs-progress-row { display: flex; gap: 12px; align-items: center; padding: 8px 0; border-top: 1px solid #eef2f7; }
.docs-progress-row:first-child { border-top: none; }
.docs-progress-row .docs-table-code { width: 170px; flex-shrink: 0; }
.docs-progress-label { flex: 1; font-size: 13px; color: #0f172a; }
.docs-progress-step { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #94a3b8; }
.docs-error-row { display: flex; gap: 12px; align-items: flex-start; padding: 10px 0; border-top: 1px solid #eef2f7; }
.docs-error-row:first-child { border-top: none; }
.docs-error-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; width: 180px; flex-shrink: 0; }
.docs-issue-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; width: 190px; flex-shrink: 0; color: #b45309; }
.docs-command-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; width: 380px; flex-shrink: 0; color: #0e7490; }
.docs-error-action { font-size: 12.5px; color: #475569; line-height: 1.6; flex: 1; }
.docs-dep-row { display: flex; gap: 12px; align-items: center; padding: 8px 0; border-top: 1px solid #eef2f7; }
.docs-dep-row:first-child { border-top: none; }
.docs-dep-name { font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #0e7490; width: 150px; flex-shrink: 0; }
.docs-dep-desc { font-size: 12.5px; color: #475569; }
.docs-hosted-card { margin-bottom: 24px; border-left: 3px solid #0e7490 !important; }
.docs-hosted-card .arco-card-body { padding: 16px 20px; }
.docs-hosted-title { font-size: 13.5px; font-weight: 600; margin-bottom: 6px; }
.docs-hosted-desc { margin: 0 0 8px; font-size: 12.5px; color: #64748b; line-height: 1.65; }
.docs-hosted-url {
  display: block;
  padding: 8px 12px;
  background: #f8fafc;
  border: 1px solid var(--rule, #e4e2dc);
  font-family: ui-monospace, Menlo, monospace;
  font-size: 12.5px;
  color: #0e7490;
}
.docs-callback-desc { margin: 0 0 10px; font-size: 12.5px; color: #64748b; line-height: 1.65; }
.docs-analysis-card .arco-card-body { padding: 16px 18px; }
.docs-analysis-title { font-family: ui-monospace, Menlo, monospace; font-size: 12.5px; font-weight: 600; margin-bottom: 10px; }
.docs-field-item { display: flex; gap: 8px; padding: 3px 0; font-size: 12px; color: #64748b; }
.docs-field-key { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #0f172a; min-width: 120px; }
.docs-risk-tile { background: #f8fafc; padding: 12px; height: 100%; }
.docs-risk-level { font-family: ui-monospace, Menlo, monospace; font-size: 12px; font-weight: 600; margin-bottom: 4px; }
.docs-risk-score { font-family: ui-monospace, Menlo, monospace; font-size: 11px; color: #94a3b8; margin-bottom: 6px; }
.docs-risk-outcome { font-size: 12px; color: #475569; line-height: 1.5; }
.docs-operator-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 4px 24px; }
.docs-operator-row { display: flex; gap: 8px; padding: 4px 0; }
.docs-operator-code { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #0e7490; min-width: 130px; flex-shrink: 0; }
.docs-operator-desc { font-size: 12px; color: #64748b; }
.docs-pattern-label { font-family: ui-monospace, Menlo, monospace; font-size: 10.5px; color: #94a3b8; margin: 4px 0 2px; }
.docs-pattern-code {
  margin: 0 0 4px;
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10.5px;
  color: #64748b;
  white-space: pre-wrap;
  word-break: break-all;
}
.docs-endpoint-row { display: flex; align-items: center; gap: 10px; padding: 5px 0; }
.docs-method { font-family: ui-monospace, Menlo, monospace; font-size: 11px; font-weight: 600; width: 52px; text-align: right; flex-shrink: 0; }
.docs-endpoint-row-path { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #0f172a; flex: 1; }
.docs-endpoint-row-desc { font-size: 12px; color: #94a3b8; }
@media (max-width: 767px) {
  .docs-check-grid,
  .docs-client-grid,
  .docs-operator-grid { grid-template-columns: 1fr; }
  .docs-error-row,
  .docs-dep-row { flex-direction: column; gap: 4px; }
  .docs-error-code,
  .docs-issue-code,
  .docs-command-code,
  .docs-dep-name { width: auto; }
}
@media (max-width: 1023px) {
  .docs-aside { display: none; }
  .docs-header { padding: 12px 18px; flex-wrap: wrap; }
  .docs-main { padding: 26px 18px 56px; }
}
</style>
