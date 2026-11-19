import { useState, useEffect, useCallback } from 'react'
import type { CSSProperties } from 'react';
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { ArrowLeftIcon, CheckCircleIcon, ArrowPathIcon } from '@heroicons/react/24/outline'
import toast from 'react-hot-toast'
import { API_BASE_URL } from '../config/api'
import { fetchCsrfToken, csrfHeader } from '../lib/csrf'
import { C, injectFonts } from '../theme'
import { LanguageSwitcher } from '../components/common/LanguageSwitcher'
import { resolveThemeVars } from '../components/verification/theme'
import { CompletionScreen } from '../components/verification/CompletionScreen'
import { Button, ColorPicker, Input, Upload } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import { Select } from 'antd'

// Theme-aware token alias: the editor shell follows the active theme;
// the device preview keeps using the page-builder config theme.
const T = {
  ...C,
  bg: 'var(--paper)',
  panel: 'var(--panel)',
  surface: 'var(--panel-2)',
  surfaceHover: 'var(--panel-3)',
  border: 'var(--rule)',
  borderStrong: 'var(--rule-strong)',
  text: 'var(--ink)',
  muted: 'var(--mid)',
  dim: 'var(--soft)',
  cyan: 'var(--accent)',
  cyanDim: 'var(--accent-soft)',
  cyanBorder: 'color-mix(in oklab, var(--accent) 30%, transparent)',
  accent: 'var(--accent)',
  accentInk: 'var(--accent-ink)',
  accentSoft: 'var(--accent-soft)',
  code: 'var(--accent-ink)',
  codeBg: 'var(--paper-2)',
}

// ─── Types ────────────────────────────────────────────────────

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
  steps: {
    front: { enabled: boolean; label: string }
    back: { enabled: boolean; label: string }
    liveness: { enabled: boolean; label: string }
  }
  completionTitle: string
  completionMessage: string
  showConfetti: boolean
}

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

const HEX_RE = /^#[0-9a-fA-F]{6}$/

// ─── Component ────────────────────────────────────────────────

export default function PageBuilderPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  useEffect(() => { injectFonts() }, [])

  const [authed, setAuthed] = useState<boolean | null>(null)
  const [config, setConfig] = useState<PageBuilderConfig>(DEFAULT_CONFIG)
  const [slug, setSlug] = useState('')
  const [slugAvailable, setSlugAvailable] = useState<boolean | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [savingSlug, setSavingSlug] = useState(false)
  const [previewDevice, setPreviewDevice] = useState<'mobile' | 'desktop'>('mobile')
  const [previewScreen, setPreviewScreen] = useState<'landing' | 'complete'>('landing')
  const [logo, setLogo] = useState<string | null>(null)
  const [uploadingLogo, setUploadingLogo] = useState(false)

  const mutationHeaders = useCallback(() => ({
    'Content-Type': 'application/json',
    ...csrfHeader(),
  }), [])

  // Step 1: Verify session + fetch CSRF token (same pattern as DeveloperPage)
  useEffect(() => {
    fetch(`${API_BASE_URL}/api/developer/profile`, { credentials: 'include' })
      .then(r => {
        if (!r.ok) { setAuthed(false); return }
        setAuthed(true)
        return fetchCsrfToken()
      })
      .catch(() => setAuthed(false))
  }, [])

  // Step 2: Once authed, load page builder config
  useEffect(() => {
    if (authed !== true) return
    fetch(`${API_BASE_URL}/api/developer/settings/page-builder`, {
      credentials: 'include',
    })
      .then(r => r.ok ? r.json() : null)
      .then(data => {
        if (!data) return
        if (data.config) setConfig({ ...DEFAULT_CONFIG, ...data.config })
        if (data.slug) setSlug(data.slug)
        if (data.logo_url) setLogo(data.logo_url)
      })
      .catch(() => toast.error('Failed to load page builder config'))
      .finally(() => setLoading(false))
  }, [authed])

  // Check slug availability (debounced)
  useEffect(() => {
    if (!slug || slug.length < 4) { setSlugAvailable(null); return }
    const timer = setTimeout(async () => {
      try {
        const r = await fetch(`${API_BASE_URL}/api/v2/verify/page-config/slug/${encodeURIComponent(slug)}`)
        const data = await r.json()
        setSlugAvailable(!data.page_builder_config)
      } catch {
        setSlugAvailable(null)
      }
    }, 500)
    return () => clearTimeout(timer)
  }, [slug])

  const saveConfig = async () => {
    setSaving(true)
    try {
      const r = await fetch(`${API_BASE_URL}/api/developer/settings/page-builder`, {
        method: 'PUT',
        credentials: 'include',
        headers: mutationHeaders(),
        body: JSON.stringify({ config }),
      })
      const data = await r.json()
      if (!r.ok) throw new Error(data.error || data.message || 'Save failed')
      toast.success('Page builder config saved')
    } catch (caught: unknown) {
      const e = caught as Error;
      toast.error(e.message || 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  const saveSlug = async () => {
    setSavingSlug(true)
    try {
      const r = await fetch(`${API_BASE_URL}/api/developer/settings/page-builder/slug`, {
        method: 'PUT',
        credentials: 'include',
        headers: mutationHeaders(),
        body: JSON.stringify({ slug: slug || null }),
      })
      const data = await r.json()
      if (!r.ok) throw new Error(data.error || data.message || 'Save failed')
      toast.success(slug ? `Slug saved: /v/${slug}` : 'Slug removed')
    } catch (caught: unknown) {
      const e = caught as Error;
      toast.error(e.message || 'Failed to save slug')
    } finally {
      setSavingSlug(false)
    }
  }

  const uploadLogo = async (file: File) => {
    if (!/\.(png|jpe?g)$/i.test(file.name)) { toast.error('Logo must be a PNG or JPEG'); return }
    if (file.size > 2 * 1024 * 1024) { toast.error('Logo must be under 2MB'); return }
    setUploadingLogo(true)
    try {
      const fd = new FormData()
      fd.append('file', file)
      const r = await fetch(`${API_BASE_URL}/api/developer/branding/logo`, {
        method: 'POST', credentials: 'include', headers: { ...csrfHeader() }, body: fd,
      })
      const data = await r.json()
      if (!r.ok) throw new Error(data.error || data.message || 'Upload failed')
      setLogo(data.data?.logo_url || null)
      toast.success('Logo uploaded')
    } catch (caught: unknown) {
      const e = caught as Error;
      toast.error(e.message || 'Logo upload failed')
    } finally {
      setUploadingLogo(false)
    }
  }

  const resetConfig = () => {
    setConfig(DEFAULT_CONFIG)
    toast.success('Reset to defaults (save to apply)')
  }

  const updateConfig = <K extends keyof PageBuilderConfig>(key: K, value: PageBuilderConfig[K]) => {
    setConfig(prev => ({ ...prev, [key]: value }))
  }

  const updateStep = (step: 'front' | 'back' | 'liveness', field: 'enabled' | 'label', value: any) => {
    setConfig(prev => ({
      ...prev,
      steps: {
        ...prev.steps,
        [step]: { ...prev.steps[step], [field]: value },
      },
    }))
  }

  const setTheme = (theme: 'dark' | 'light') => {
    if (theme === 'dark') {
      updateConfig('theme', 'dark')
      updateConfig('backgroundColor', '#080c14')
      updateConfig('cardBackgroundColor', '#0f1420')
      updateConfig('textColor', '#dde2ec')
      updateConfig('accentColor', '#00F0FF')
      updateConfig('mutedTextColor', '#8a8a90')
      updateConfig('borderColor', '#1e1e22')
    } else {
      updateConfig('theme', 'light')
      updateConfig('backgroundColor', '#f8fafc')
      updateConfig('cardBackgroundColor', '#ffffff')
      updateConfig('textColor', '#1e293b')
      updateConfig('accentColor', '#00d4d4')
      updateConfig('mutedTextColor', '#6b6b70')
      updateConfig('borderColor', '#e4e2dc')
    }
  }

  // ─── Auth guard ─────────────────────────────────────────────
  useEffect(() => {
    if (authed === false) navigate('/developer', { replace: true })
  }, [authed, navigate])
  if (authed === false) return null
  if (authed === null || loading) {
    return (
      <div style={{ background: T.bg, minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div style={{ color: T.muted, fontFamily: T.sans, fontSize: 14 }}>{t('pageBuilder.loading')}</div>
      </div>
    )
  }

  // ─── Shared styles ──────────────────────────────────────────

  const sectionTitle: CSSProperties = {
    fontFamily: T.mono, fontSize: 11, fontWeight: 600, color: T.cyan,
    letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: 12,
  }

  const label: CSSProperties = {
    fontFamily: T.sans, fontSize: 12, fontWeight: 500, color: T.muted,
    display: 'block', marginBottom: 4,
  }

  const input: CSSProperties = {
    width: '100%', background: T.surface, border: `1px solid ${T.border}`,
    color: T.text, borderRadius: 6, padding: '8px 10px', fontSize: 13,
    fontFamily: T.sans, outline: 'none', boxSizing: 'border-box',
  }

  const toggle = (on: boolean): CSSProperties => ({
    width: 36, height: 20, borderRadius: 10, cursor: 'pointer', border: 'none',
    background: on ? T.cyan : T.dim, position: 'relative', transition: 'background 0.2s',
    flexShrink: 0,
  })

  const toggleDot = (on: boolean): CSSProperties => ({
    width: 16, height: 16, borderRadius: '50%', background: '#fff',
    position: 'absolute', top: 2, left: on ? 18 : 2, transition: 'left 0.2s',
  })

  // ─── Preview derived values ─────────────────────────────────

  // Live preview theming: the real verification pages theme by spreading these
  // CSS-var overrides; the preview does the same so it matches production and
  // re-themes as the operator edits. resolveThemeVars keeps them in one place.
  const themeVars = resolveThemeVars(config) as CSSProperties
  const previewBranding = { logo_url: logo, accent_color: config.accentColor || null, company_name: null }
  const enabledSteps = (['front', 'back', 'liveness'] as const).filter(s => config.steps[s].enabled)

  const previewStepLabel = (key: 'front' | 'back' | 'liveness') => {
    const raw = config.steps[key].label
    if (raw && raw !== DEFAULT_CONFIG.steps[key].label) return raw
    return t(`pageBuilder.preview.${key === 'liveness' ? 'liveStep' : key === 'front' ? 'frontStep' : 'backStep'}`)
  }

  const previewConfig = {
    ...config,
    headerTitle: config.headerTitle === DEFAULT_CONFIG.headerTitle
      ? t('pageBuilder.preview.verifyIdentity')
      : config.headerTitle,
    headerSubtitle: config.headerSubtitle === DEFAULT_CONFIG.headerSubtitle
      ? t('pageBuilder.preview.completeStepsFull')
      : config.headerSubtitle,
    completionTitle: config.completionTitle === DEFAULT_CONFIG.completionTitle
      ? t('completion.successTitle')
      : config.completionTitle,
    completionMessage: config.completionMessage === DEFAULT_CONFIG.completionMessage
      ? t('completion.bodySuccess')
      : config.completionMessage,
  }

  // ─── Render ─────────────────────────────────────────────────

  return (
    <div style={{ background: T.bg, minHeight: '100vh', fontFamily: T.sans, color: T.text }}>
      {/* Top bar */}
      <div style={{
        background: T.panel,
        borderBottom: `1px solid ${T.border}`, padding: '12px 24px',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <Button onClick={() => navigate('/developer')} style={{
            background: 'none', border: 'none', color: T.muted, cursor: 'pointer', display: 'flex', alignItems: 'center',
          }}>
            <ArrowLeftIcon style={{ width: 16, height: 16 }} />
          </Button>
          <span style={{ fontFamily: T.mono, fontSize: 13, color: T.text, fontWeight: 600 }}>
            {t('pageBuilder.title')}
          </span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <LanguageSwitcher variant="toolbar" />
          <Button onClick={resetConfig} style={{
            background: 'none', border: `1px solid ${T.border}`, color: T.muted,
            borderRadius: 6, padding: '6px 14px', cursor: 'pointer', fontSize: 12, display: 'flex', alignItems: 'center', gap: 6,
          }}>
            <ArrowPathIcon style={{ width: 14, height: 14 }} /> {t('pageBuilder.reset')}
          </Button>
          <Button onClick={saveConfig} disabled={saving} style={{
            background: T.cyan, border: 'none', color: 'var(--accent-fg)', borderRadius: 6,
            padding: '6px 18px', cursor: saving ? 'not-allowed' : 'pointer', fontSize: 12, fontWeight: 600,
            opacity: saving ? 0.6 : 1, display: 'flex', alignItems: 'center', gap: 6,
          }}>
            <CheckCircleIcon style={{ width: 14, height: 14 }} /> {saving ? t('pageBuilder.saving') : t('pageBuilder.save')}
          </Button>
        </div>
      </div>

      {/* Two-panel layout */}
      <div style={{ display: 'flex', height: 'calc(100vh - 49px)' }}>

        {/* ── Left: Controls ── */}
        <div style={{
          width: '55%', overflowY: 'auto', padding: '24px 28px',
          background: T.panel,
          borderRight: `1px solid ${T.border}`,
        }}>

          {/* Header section */}
          <div style={{ marginBottom: 28 }}>
            <div style={sectionTitle}>{t('pageBuilder.sections.header')}</div>
            <div style={{ marginBottom: 10 }}>
              <label style={label}>{t('pageBuilder.titleLabel')}</label>
              <Input style={input} value={config.headerTitle}
                onChange={e => updateConfig('headerTitle', e.target.value)} maxLength={200} />
            </div>
            <div>
              <label style={label}>{t('pageBuilder.subtitle')}</label>
              <Input style={input} value={config.headerSubtitle}
                onChange={e => updateConfig('headerSubtitle', e.target.value)} maxLength={300} />
            </div>
          </div>

          {/* Logo section */}
          <div style={{ marginBottom: 28 }}>
            <div style={sectionTitle}>{t('pageBuilder.sections.logo')}</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
              <div style={{
                width: 56, height: 56, borderRadius: 10, flexShrink: 0,
                border: `1px solid ${T.border}`, background: T.surface,
                display: 'flex', alignItems: 'center', justifyContent: 'center', overflow: 'hidden',
              }}>
                {logo
                  ? <img src={logo} alt="Logo" style={{ maxWidth: '100%', maxHeight: '100%', objectFit: 'contain' }} />
                  : <span style={{ fontSize: 10, color: T.muted }}>{t('pageBuilder.none')}</span>}
              </div>
              <Upload
                accept="image/png,image/jpeg"
                showUploadList={false}
                disabled={uploadingLogo}
                beforeUpload={(file) => { void uploadLogo(file as File); return false }}
              >
                <Button icon={<UploadOutlined />} loading={uploadingLogo}>
                  {uploadingLogo ? t('pageBuilder.uploading') : logo ? t('pageBuilder.replaceLogo') : t('pageBuilder.uploadLogo')}
                </Button>
              </Upload>
            </div>
            <div style={{ fontSize: 11, color: T.muted, marginTop: 8 }}>
              {t('pageBuilder.logoHint')}
            </div>
          </div>

          {/* Theme section */}
          <div style={{ marginBottom: 28 }}>
            <div style={sectionTitle}>{t('pageBuilder.sections.theme')}</div>
            <div style={{ display: 'flex', gap: 8, marginBottom: 14 }}>
              {(['dark', 'light'] as const).map(mode => (
                <Button key={mode} onClick={() => setTheme(mode)} style={{
                  flex: 1, padding: '10px 0', borderRadius: 6, cursor: 'pointer', fontSize: 13, fontWeight: 500,
                  background: config.theme === mode ? T.cyan : T.surface,
                  color: config.theme === mode ? 'var(--accent-fg)' : T.muted,
                  border: `1px solid ${config.theme === mode ? T.cyan : T.border}`,
                }}>
                  {mode === 'dark' ? t('pageBuilder.dark') : t('pageBuilder.light')}
                </Button>
              ))}
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 10 }}>
              {([
                ['backgroundColor', t('pageBuilder.colors.background')],
                ['cardBackgroundColor', t('pageBuilder.colors.card')],
                ['textColor', t('pageBuilder.colors.text')],
              ] as const).map(([key, lbl]) => (
                <div key={key}>
                  <label style={label}>{lbl}</label>
                  <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                    <Input style={{ ...input, flex: 1 }} value={config[key]}
                      onChange={e => updateConfig(key, e.target.value)} maxLength={7} />
                    <ColorPicker
                      value={HEX_RE.test(config[key]) ? config[key] : '#000000'}
                      onChange={(color) => updateConfig(key, color.toHexString())}
                      showText={false}
                    />
                  </div>
                </div>
              ))}
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 10, marginTop: 10 }}>
              {([
                ['accentColor', t('pageBuilder.colors.accent')],
                ['mutedTextColor', t('pageBuilder.colors.muted')],
                ['borderColor', t('pageBuilder.colors.borders')],
              ] as const).map(([key, lbl]) => (
                <div key={key}>
                  <label style={label}>{lbl}</label>
                  <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                    <Input style={{ ...input, flex: 1 }} value={config[key]}
                      onChange={e => updateConfig(key, e.target.value)} maxLength={7} />
                    <ColorPicker
                      value={HEX_RE.test(config[key]) ? config[key] : '#000000'}
                      onChange={(color) => updateConfig(key, color.toHexString())}
                      showText={false}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Typography section */}
          <div style={{ marginBottom: 28 }}>
            <div style={sectionTitle}>{t('pageBuilder.sections.typography')}</div>
            <label style={label}>{t('pageBuilder.fontFamily')}</label>
            <Select style={{ ...input, cursor: 'pointer' }} value={config.fontFamily}
              onChange={e => updateConfig('fontFamily', e as any)} options={[{ value: "dm-sans", label: "DM Sans" }, { value: "inter", label: "Inter" }, { value: "system", label: t('pageBuilder.systemDefault') }]} />
          </div>

          {/* Steps section */}
          <div style={{ marginBottom: 28 }}>
            <div style={sectionTitle}>{t('pageBuilder.sections.steps')}</div>
            {(['front', 'back', 'liveness'] as const).map(step => (
              <div key={step} style={{
                display: 'flex', alignItems: 'center', gap: 12, marginBottom: 10,
                padding: '10px 12px', background: T.surface, borderRadius: 8, border: `1px solid ${T.border}`,
              }}>
                <Button style={toggle(config.steps[step].enabled)}
                  onClick={() => updateStep(step, 'enabled', !config.steps[step].enabled)}>
                  <div style={toggleDot(config.steps[step].enabled)} />
                </Button>
                <Input style={{ ...input, marginBottom: 0 }} value={config.steps[step].label}
                  onChange={e => updateStep(step, 'label', e.target.value)} maxLength={50} />
              </div>
            ))}
          </div>

          {/* Completion section */}
          <div style={{ marginBottom: 28 }}>
            <div style={sectionTitle}>{t('pageBuilder.sections.completion')}</div>
            <div style={{ marginBottom: 10 }}>
              <label style={label}>{t('pageBuilder.completionTitle')}</label>
              <Input style={input} value={config.completionTitle}
                onChange={e => updateConfig('completionTitle', e.target.value)} maxLength={200} />
            </div>
            <div style={{ marginBottom: 10 }}>
              <label style={label}>{t('pageBuilder.completionMessage')}</label>
              <Input style={input} value={config.completionMessage}
                onChange={e => updateConfig('completionMessage', e.target.value)} maxLength={500} />
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <Button style={toggle(config.showConfetti)}
                onClick={() => updateConfig('showConfetti', !config.showConfetti)}>
                <div style={toggleDot(config.showConfetti)} />
              </Button>
              <span style={{ fontSize: 13, color: T.muted }}>{t('pageBuilder.showConfetti')}</span>
            </div>
          </div>

          {/* Options section */}
          <div style={{ marginBottom: 28 }}>
            <div style={sectionTitle}>{t('pageBuilder.sections.options')}</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <Button style={toggle(config.showPoweredBy)}
                onClick={() => updateConfig('showPoweredBy', !config.showPoweredBy)}>
                <div style={toggleDot(config.showPoweredBy)} />
              </Button>
              <span style={{ fontSize: 13, color: T.muted }}>{t('pageBuilder.showPoweredBy')}</span>
            </div>
          </div>

          {/* Custom URL section */}
          <div style={{ marginBottom: 28 }}>
            <div style={sectionTitle}>{t('pageBuilder.sections.customUrl')}</div>
            <label style={label}>{t('pageBuilder.slugLabel')}</label>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <span style={{ fontSize: 13, color: T.dim, whiteSpace: 'nowrap' }}>/v/</span>
              <Input style={{ ...input, flex: 1 }} value={slug} placeholder={t('pageBuilder.slugPlaceholder')}
                onChange={e => setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))} maxLength={50} />
              <Button onClick={saveSlug} disabled={savingSlug} style={{
                background: T.surface, border: `1px solid ${T.border}`, color: T.muted,
                borderRadius: 6, padding: '8px 14px', cursor: savingSlug ? 'not-allowed' : 'pointer', fontSize: 12, whiteSpace: 'nowrap',
              }}>
                {savingSlug ? '...' : t('pageBuilder.saveSlug')}
              </Button>
            </div>
            {slug && slug.length >= 4 && slugAvailable !== null && (
              <div style={{ fontSize: 12, marginTop: 6, color: slugAvailable ? T.green : T.red }}>
                {slugAvailable ? 'Available' : 'Already taken'}
              </div>
            )}
          </div>
        </div>

        {/* ── Right: Live Preview (device-accurate, re-themes as you edit) ── */}
        <div style={{
          width: '45%', overflowY: 'auto', padding: 24,
          background: T.surface,
          display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 16,
        }}>
          {/* Device + screen toggles */}
          <div style={{ display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap', justifyContent: 'center' }}>
            {(['mobile', 'desktop'] as const).map(d => (
              <Button key={d} onClick={() => setPreviewDevice(d)} style={{
                padding: '6px 14px', borderRadius: 8, fontSize: 12, fontWeight: 600, cursor: 'pointer',
                border: `1px solid ${previewDevice === d ? T.cyan : T.border}`,
                background: previewDevice === d ? T.cyan : 'transparent',
                color: previewDevice === d ? 'var(--accent-fg)' : T.muted,
              }}>{d === 'mobile' ? t('pageBuilder.mobile') : t('pageBuilder.desktop')}</Button>
            ))}
            <span style={{ width: 1, height: 20, background: T.border, margin: '0 4px' }} />
            {(['landing', 'complete'] as const).map(s => (
              <Button key={s} onClick={() => setPreviewScreen(s)} style={{
                padding: '6px 14px', borderRadius: 8, fontSize: 12, fontWeight: 600, cursor: 'pointer',
                border: `1px solid ${previewScreen === s ? T.cyan : T.border}`,
                background: previewScreen === s ? T.cyan : 'transparent',
                color: previewScreen === s ? 'var(--accent-fg)' : T.muted,
              }}>{s === 'landing' ? t('pageBuilder.landing') : t('pageBuilder.completion')}</Button>
            ))}
          </div>

          {/* Device frame — the inner container spreads themeVars so children
              resolve var(--paper)/var(--ink)/var(--accent)/… to the draft config,
              exactly like the real verification pages. */}
          <div style={{
            background: '#000', borderRadius: previewDevice === 'mobile' ? 36 : 12,
            padding: previewDevice === 'mobile' ? 10 : 8,
            width: previewDevice === 'mobile' ? 300 : '100%',
            maxWidth: previewDevice === 'mobile' ? 300 : 460,
            boxShadow: '0 24px 60px rgba(0,0,0,0.5)',
          }}>
            <div style={{
              ...themeVars,
              background: 'var(--paper)', color: 'var(--ink)', fontFamily: 'var(--sans)',
              borderRadius: previewDevice === 'mobile' ? 28 : 6, overflow: 'hidden',
              minHeight: previewDevice === 'mobile' ? 540 : 420,
            }}>
              {previewScreen === 'complete' ? (
                <CompletionScreen
                  device={previewDevice}
                  config={previewConfig}
                  branding={previewBranding}
                  result={{ status: 'verified', confidence_score: 0.97, face_match_score: 0.96, liveness_score: 0.98 }}
                />
              ) : previewDevice === 'mobile' ? (
                /* Mobile landing — mirrors the /verify/mobile first screen */
                <div style={{ paddingBottom: 20 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', padding: '12px 18px 8px', fontFamily: 'var(--mono)', fontSize: 10, color: 'var(--mid)' }}>
                    <span>9:41</span>
                    <span style={{ color: 'var(--accent)', letterSpacing: '0.1em', textTransform: 'uppercase' }}>{t('pageBuilder.preview.secureSession')}</span>
                    <span>●●●</span>
                  </div>
                  {logo && <div style={{ padding: '6px 18px 2px' }}><img src={logo} alt="" style={{ height: 22, maxWidth: '60%', objectFit: 'contain', display: 'block', margin: '0 auto' }} /></div>}
                  <div style={{ display: 'flex', gap: 4, padding: '8px 18px' }}>
                    {[...enabledSteps.map(s => previewStepLabel(s)), t('pageBuilder.preview.complete')].map((lbl, i) => (
                      <div key={i} style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ height: 3, background: i === 0 ? 'var(--accent)' : 'var(--rule)', borderRadius: 2 }} />
                        <div style={{ fontSize: 8, marginTop: 4, fontFamily: 'var(--mono)', color: i === 0 ? 'var(--ink)' : 'var(--mid)', textTransform: 'uppercase', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{lbl}</div>
                      </div>
                    ))}
                  </div>
                  <div style={{ padding: '16px 18px 0' }}>
                    <div style={{ fontFamily: 'var(--mono)', fontSize: 9, color: 'var(--accent)', letterSpacing: '0.08em', marginBottom: 8 }}>{t('pageBuilder.preview.stepOf', { current: 1, total: enabledSteps.length + 1 })}</div>
                    <h2 style={{ fontSize: 22, fontWeight: 600, color: 'var(--ink)', margin: '0 0 6px', lineHeight: 1.15 }}>{previewConfig.headerTitle}</h2>
                    <p style={{ fontSize: 13, color: 'var(--mid)', margin: 0 }}>{previewConfig.headerSubtitle}</p>
                  </div>
                  <div style={{ margin: '16px 18px', padding: '30px 16px', background: 'var(--panel)', border: '2px solid var(--accent)', borderRadius: 12, textAlign: 'center' }}>
                    <div style={{ fontSize: 11, color: 'var(--mid)', fontFamily: 'var(--mono)', letterSpacing: '0.04em' }}>{t('pageBuilder.preview.cameraPreview')}</div>
                  </div>
                  <div style={{ padding: '0 18px' }}>
                    <div style={{ background: 'var(--accent)', color: 'var(--paper)', textAlign: 'center', padding: 14, borderRadius: 10, fontWeight: 600, fontSize: 13, fontFamily: 'var(--mono)', letterSpacing: '0.04em', textTransform: 'uppercase' }}>{t('pageBuilder.preview.takePhoto', { label: previewStepLabel('front') })}</div>
                  </div>
                  {config.showPoweredBy && <div style={{ textAlign: 'center', padding: '18px 0 0', fontSize: 10, color: 'var(--mid)' }}>{t('pageBuilder.preview.poweredBy')}</div>}
                </div>
              ) : (
                /* Desktop landing — mirrors the choice screen */
                <div style={{ padding: '36px 28px', textAlign: 'center' }}>
                  {logo && <img src={logo} alt="" style={{ height: 30, maxWidth: '55%', objectFit: 'contain', display: 'block', margin: '0 auto 18px' }} />}
                  <h2 style={{ fontSize: 24, fontWeight: 600, color: 'var(--ink)', margin: '0 0 8px' }}>{previewConfig.headerTitle}</h2>
                  <p style={{ fontSize: 14, color: 'var(--mid)', margin: '0 0 24px' }}>{config.headerSubtitle === DEFAULT_CONFIG.headerSubtitle ? t('pageBuilder.preview.chooseMethod') : previewConfig.headerSubtitle}</p>
                  <div style={{ display: 'flex', gap: 12 }}>
                    {[t('pageBuilder.preview.continueDevice'), t('pageBuilder.preview.scanQr')].map((label, i) => (
                      <div key={i} style={{ flex: 1, padding: '22px 16px', background: 'var(--panel)', border: `1.5px solid ${i === 0 ? 'var(--accent)' : 'var(--rule)'}`, borderRadius: 12 }}>
                        <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--ink)' }}>{label}</div>
                      </div>
                    ))}
                  </div>
                  {config.showPoweredBy && <div style={{ padding: '24px 0 0', fontSize: 11, color: 'var(--mid)' }}>{t('pageBuilder.preview.poweredBy')}</div>}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
