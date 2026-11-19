<template>
  <div class="docs-shell md-shell">
    <!-- Header bar -->
    <div class="docs-header">
      <div class="docs-brand">
        <span class="docs-brand-accent">identity</span>
        <span class="docs-brand-dim"> / </span>
        <span>api-docs</span>
        <span class="docs-brand-dim"> / </span>
        <span class="md-brand-muted">markdown</span>
      </div>

      <div class="docs-header-actions">
        <router-link to="/docs" class="docs-chip md-chip">
          <icon-left /> Back to Docs
        </router-link>
        <a-link :href="markdownUrl" target="_blank" class="docs-chip md-chip">
          <icon-download /> Download .md
        </a-link>
        <a-tag size="small" color="green">v1.7.0</a-tag>
      </div>
    </div>

    <div class="md-container">
      <div v-if="loading" class="md-center">
        <a-spin dot :size="28" />
      </div>

      <div v-else-if="error" class="md-error">
        <p>Failed to load documentation: {{ error }}</p>
        <a-button type="primary" @click="load">Retry</a-button>
      </div>

      <div v-else-if="!blocks.length" class="md-empty">
        <a-empty description="No markdown content yet" />
      </div>

      <div v-else class="md-body">
        <template v-for="(block, index) in blocks" :key="index">
          <h1 v-if="block.type === 'h1'" class="md-h1"><MdInline :tokens="block.inline || []" /></h1>
          <h2 v-else-if="block.type === 'h2'" class="md-h2"><MdInline :tokens="block.inline || []" /></h2>
          <h3 v-else-if="block.type === 'h3'" class="md-h3"><MdInline :tokens="block.inline || []" /></h3>
          <h4 v-else-if="block.type === 'h4'" class="md-h4"><MdInline :tokens="block.inline || []" /></h4>
          <p v-else-if="block.type === 'p'" class="md-p"><MdInline :tokens="block.inline || []" /></p>
          <div v-else-if="block.type === 'code'" class="docs-code">
            <div class="docs-code-head">
              <span class="docs-code-label">{{ block.lang || 'code' }}</span>
              <a-button size="mini" type="text" class="docs-code-copy" @click="copy(block.code || '')">Copy</a-button>
            </div>
            <pre class="docs-code-pre">{{ block.code }}</pre>
          </div>
          <div v-else-if="block.type === 'table'" class="md-table-wrap">
            <table class="md-table">
              <thead>
                <tr><th v-for="(cell, cellIndex) in block.head || []" :key="cellIndex">{{ cell }}</th></tr>
              </thead>
              <tbody>
                <tr v-for="(row, rowIndex) in block.rows || []" :key="rowIndex">
                  <td v-for="(cell, cellIndex) in row" :key="cellIndex">{{ cell }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <ul v-else-if="block.type === 'ul'" class="md-list">
            <li v-for="(item, itemIndex) in block.items || []" :key="itemIndex"><MdInline :tokens="item" /></li>
          </ul>
          <ol v-else-if="block.type === 'ol'" class="md-list">
            <li v-for="(item, itemIndex) in block.items || []" :key="itemIndex"><MdInline :tokens="item" /></li>
          </ol>
          <blockquote v-else-if="block.type === 'quote'" class="md-quote">
            <MdInline :tokens="block.inline || []" />
          </blockquote>
          <hr v-else-if="block.type === 'hr'" class="md-hr" />
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineComponent, h, onMounted, ref, type PropType } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
// Pulls in the shared docs shell styles (header bar, chips, code panel) for this standalone page.
import '../components/DocsShell.vue'

interface MdToken {
  type: 'text' | 'code' | 'strong' | 'em' | 'link'
  text: string
  href?: string
}

interface MdBlock {
  type: 'h1' | 'h2' | 'h3' | 'h4' | 'p' | 'code' | 'table' | 'ul' | 'ol' | 'quote' | 'hr'
  inline?: MdToken[]
  lang?: string
  code?: string
  head?: string[]
  rows?: string[][]
  items?: MdToken[][]
}

const apiUrl = window.location.origin
const markdownUrl = `${apiUrl}/api/docs/markdown`

const markdown = ref('')
const loading = ref(true)
const error = ref<string | null>(null)
const blocks = ref<MdBlock[]>([])

const MdInline = defineComponent({
  name: 'MdInline',
  props: { tokens: { type: Array as PropType<MdToken[]>, default: () => [] } },
  setup(props) {
    return () =>
      props.tokens.map((token) => {
        if (token.type === 'code') return h('code', { class: 'md-code' }, token.text)
        if (token.type === 'strong') return h('strong', null, token.text)
        if (token.type === 'em') return h('em', { class: 'md-em' }, token.text)
        if (token.type === 'link')
          return h('a', { href: token.href, target: '_blank', rel: 'noopener noreferrer', class: 'md-link' }, token.text)
        return token.text
      })
  },
})

function splitRow(line: string): string[] {
  return line
    .trim()
    .replace(/^\|/, '')
    .replace(/\|$/, '')
    .split('|')
    .map((cell) => cell.trim())
}

function parseInline(text: string): MdToken[] {
  const tokens: MdToken[] = []
  const pattern = /(`[^`]+`|\*\*[^*]+\*\*|\*[^*]+\*|\[[^\]]+\]\([^)]+\))/
  let rest = text
  while (rest.length) {
    const match = rest.match(pattern)
    if (!match || match.index === undefined) {
      tokens.push({ type: 'text', text: rest })
      break
    }
    if (match.index > 0) tokens.push({ type: 'text', text: rest.slice(0, match.index) })
    const raw = match[0]
    if (raw.startsWith('`')) tokens.push({ type: 'code', text: raw.slice(1, -1) })
    else if (raw.startsWith('**')) tokens.push({ type: 'strong', text: raw.slice(2, -2) })
    else if (raw.startsWith('*')) tokens.push({ type: 'em', text: raw.slice(1, -1) })
    else {
      const link = raw.match(/^\[([^\]]+)\]\(([^)]+)\)$/)
      if (link) tokens.push({ type: 'link', text: link[1], href: link[2] })
      else tokens.push({ type: 'text', text: raw })
    }
    rest = rest.slice(match.index + raw.length)
  }
  return tokens
}

function parseMarkdown(source: string): MdBlock[] {
  const lines = source.replace(/\r\n/g, '\n').split('\n')
  const result: MdBlock[] = []
  let index = 0

  while (index < lines.length) {
    const line = lines[index]
    if (!line.trim()) {
      index += 1
      continue
    }

    const fence = line.match(/^```(\S*)/)
    if (fence) {
      const lang = fence[1] || ''
      const buffer: string[] = []
      index += 1
      while (index < lines.length && !lines[index].startsWith('```')) {
        buffer.push(lines[index])
        index += 1
      }
      index += 1
      result.push({ type: 'code', lang, code: buffer.join('\n') })
      continue
    }

    if (/^(\*\*\*|---|___)\s*$/.test(line)) {
      result.push({ type: 'hr' })
      index += 1
      continue
    }

    const heading = line.match(/^(#{1,6})\s+(.*)$/)
    if (heading) {
      const level = heading[1].length
      const type = level === 1 ? 'h1' : level === 2 ? 'h2' : level === 3 ? 'h3' : 'h4'
      result.push({ type, inline: parseInline(heading[2]) })
      index += 1
      continue
    }

    if (line.includes('|') && lines[index + 1] && /^\s*\|?[\s:|-]+\|/.test(lines[index + 1])) {
      const head = splitRow(line)
      index += 2
      const rows: string[][] = []
      while (index < lines.length && lines[index].includes('|') && lines[index].trim()) {
        rows.push(splitRow(lines[index]))
        index += 1
      }
      result.push({ type: 'table', head, rows })
      continue
    }

    if (/^>\s?/.test(line)) {
      const buffer: string[] = []
      while (index < lines.length && /^>\s?/.test(lines[index])) {
        buffer.push(lines[index].replace(/^>\s?/, ''))
        index += 1
      }
      result.push({ type: 'quote', inline: parseInline(buffer.join(' ')) })
      continue
    }

    if (/^\s*[-*+]\s+/.test(line) || /^\s*\d+\.\s+/.test(line)) {
      const ordered = /^\s*\d+\.\s+/.test(line)
      const items: MdToken[][] = []
      while (index < lines.length && (/^\s*[-*+]\s+/.test(lines[index]) || /^\s*\d+\.\s+/.test(lines[index]))) {
        items.push(parseInline(lines[index].replace(/^\s*(?:[-*+]|\d+\.)\s+/, '')))
        index += 1
      }
      result.push({ type: ordered ? 'ol' : 'ul', items })
      continue
    }

    const buffer: string[] = [line]
    index += 1
    while (
      index < lines.length &&
      lines[index].trim() &&
      !/^(#{1,6}\s|```|>|\s*[-*+]\s|\s*\d+\.\s)/.test(lines[index]) &&
      !/^(\*\*\*|---|___)\s*$/.test(lines[index])
    ) {
      buffer.push(lines[index])
      index += 1
    }
    result.push({ type: 'p', inline: parseInline(buffer.join(' ')) })
  }

  return result
}

async function load() {
  loading.value = true
  error.value = null
  try {
    const response = await api.get('/docs/markdown', { responseType: 'text' })
    markdown.value = typeof response.data === 'string' ? response.data : String(response.data ?? '')
    blocks.value = parseMarkdown(markdown.value)
  } catch (err) {
    const status = (err as { response?: { status?: number } }).response?.status
    error.value = status ? `HTTP ${status}` : apiMessage(err)
    markdown.value = ''
    blocks.value = []
  } finally {
    loading.value = false
  }
}

async function copy(code: string) {
  try {
    await navigator.clipboard.writeText(code)
    Message.success('Copied to clipboard')
  } catch {
    Message.error('Copy failed')
  }
}

onMounted(load)
</script>

<style>
.md-shell { margin: 0; }
.md-brand-muted { color: #64748b; }
.md-chip { display: inline-flex; align-items: center; gap: 4px; font-size: 11.5px; }
.md-container { max-width: 860px; margin: 0 auto; padding: 36px 32px 80px; }
.md-center { display: flex; justify-content: center; padding: 80px 0; }
.md-error { text-align: center; padding: 80px 0; color: #64748b; }
.md-error p { font-size: 13.5px; margin-bottom: 16px; }
.md-empty { padding: 60px 0; }
.md-p { font-size: 14px; line-height: 1.7; color: #0f172a; margin: 0 0 16px; }
.md-h1 {
  font-size: 30px;
  font-weight: 700;
  color: #0e7490;
  margin: 44px 0 16px;
  line-height: 1.3;
  border-bottom: 1px solid var(--rule, #e4e2dc);
  padding-bottom: 12px;
}
.md-h2 {
  font-size: 22px;
  font-weight: 600;
  margin: 36px 0 12px;
  line-height: 1.3;
  border-bottom: 1px solid var(--rule, #e4e2dc);
  padding-bottom: 8px;
}
.md-h3 { font-size: 17px; font-weight: 600; margin: 28px 0 8px; }
.md-h4 { font-size: 14.5px; font-weight: 600; color: #64748b; margin: 22px 0 8px; }
.md-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #0e7490; background: #f1f5f9; padding: 1px 5px; }
.md-em { color: #64748b; }
.md-link { color: #0e7490; text-decoration: none; border-bottom: 1px solid #bae6fd; }
.md-link:hover { border-bottom-color: #0e7490; }
.md-list { padding-left: 22px; margin: 0 0 16px; line-height: 1.7; font-size: 14px; }
.md-list li { margin: 4px 0; }
.md-quote { border-left: 3px solid #0e7490; margin: 0 0 16px; padding: 12px 20px; background: #ecfeff; color: #475569; }
.md-hr { border: none; border-top: 1px solid var(--rule, #e4e2dc); margin: 28px 0; }
.md-table-wrap { overflow-x: auto; margin: 0 0 16px; }
.md-table { width: 100%; border-collapse: collapse; font-size: 13px; border: 1px solid var(--rule, #e4e2dc); }
.md-table thead { background: #f8fafc; }
.md-table th {
  text-align: left;
  padding: 9px 12px;
  font-weight: 600;
  border-bottom: 1px solid var(--rule, #e4e2dc);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.md-table td { padding: 9px 12px; border-bottom: 1px solid #eef2f7; }
@media (max-width: 767px) {
  .md-container { padding: 24px 18px 56px; }
}
</style>
