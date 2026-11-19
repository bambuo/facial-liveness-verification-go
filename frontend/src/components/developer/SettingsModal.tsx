import { useState, useEffect, type CSSProperties } from 'react'
import type { ComponentType } from 'react';
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import toast from 'react-hot-toast'
import { API_BASE_URL } from '../../config/api'
import { csrfHeader, clearCsrfToken } from '../../lib/csrf'
import { C } from '../../theme'
import { formatDate } from '../../i18n/formatters'
import {
  Cog6ToothIcon,
  UserCircleIcon,
  UsersIcon,
  XMarkIcon,
  CodeBracketIcon,
  DevicePhoneMobileIcon,
  PaintBrushIcon,
  ExclamationTriangleIcon,
  FingerPrintIcon,
  ShieldExclamationIcon,
  MicrophoneIcon,
  ServerIcon,
} from '@heroicons/react/24/outline'
import { labelStyle } from './types'
import { Alert, Button, ColorPicker, Input, Modal, Select, Spin, Switch, Tabs, Tag, Upload } from 'antd'

type SettingsTab = 'profile' | 'team' | 'integrations' | 'branding' | 'system' | 'account'

// Theme-aware token alias: structural colors follow the active light/dark theme.
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
  accent: 'var(--accent)',
  accentInk: 'var(--accent-ink)',
  accentSoft: 'var(--accent-soft)',
  code: 'var(--accent-ink)',
  codeBg: 'var(--paper-2)',
}

const tabs: Array<{ id: SettingsTab; label: string; icon: ComponentType<{ style?: CSSProperties }>; bottom?: boolean }> = [
  { id: 'profile', label: 'settings.tabs.profile', icon: UserCircleIcon },
  { id: 'team', label: 'settings.tabs.team', icon: UsersIcon },
  { id: 'integrations', label: 'settings.tabs.integrations', icon: CodeBracketIcon },
  { id: 'branding', label: 'settings.tabs.branding', icon: PaintBrushIcon },
  { id: 'system', label: 'settings.tabs.system', icon: ServerIcon },
  { id: 'account', label: 'settings.tabs.account', icon: ExclamationTriangleIcon, bottom: true },
]

interface SettingsModalProps {
  token: string
  onClose: () => void
  onAccountDeleted: () => void
  // Accepted so DeveloperPage can pass it (the private SettingsModal uses it
  // for operator-scoped UI); the community override doesn't render that UI.
  isOperator?: boolean
}

export function SettingsModal({ token, onClose, onAccountDeleted }: SettingsModalProps) {
  const { t } = useTranslation()
  const authHeaders = (token === 'session' ? {} : { Authorization: `Bearer ${token}` }) as Record<string, string>
  // Profile settings
  const [profileName, setProfileName] = useState('')
  const [profileCompany, setProfileCompany] = useState('')
  const [profileEmail, setProfileEmail] = useState('')
  const [profileAvatarUrl, setProfileAvatarUrl] = useState('')
  const [profileLoading, setProfileLoading] = useState(false)
  const [profileSaving, setProfileSaving] = useState(false)

  // LLM Enhancement settings
  const [llmProvider, setLlmProvider] = useState<string>('')
  const [llmApiKey, setLlmApiKey] = useState<string>('')
  const [llmEndpointUrl, setLlmEndpointUrl] = useState<string>('')
  const [llmModel, setLlmModel] = useState<string>('')
  const [llmKeyPreview, setLlmKeyPreview] = useState<string>('')
  const [llmConfigured, setLlmConfigured] = useState(false)
  const [llmSaving, setLlmSaving] = useState(false)
  const [llmLoading, setLlmLoading] = useState(false)
  const [showLlmKey, setShowLlmKey] = useState(false)

  // SMS Provider settings
  const [smsProvider, setSmsProvider] = useState<string>('')
  const [smsApiKey, setSmsApiKey] = useState<string>('')
  const [smsApiSecret, setSmsApiSecret] = useState<string>('')
  const [smsPhoneNumber, setSmsPhoneNumber] = useState<string>('')
  const [smsKeyPreview, setSmsKeyPreview] = useState<string>('')
  const [smsConfigured, setSmsConfigured] = useState(false)
  const [smsSaving, setSmsSaving] = useState(false)
  const [smsLoading, setSmsLoading] = useState(false)
  const [showSmsKey, setShowSmsKey] = useState(false)

  // AML / Sanctions screening settings
  const [amlEnabled, setAmlEnabled] = useState(true)
  const [amlLoading, setAmlLoading] = useState(false)
  const [amlSaving, setAmlSaving] = useState(false)

  // Duplicate detection settings
  const [dedupEnabled, setDedupEnabled] = useState(false)
  const [dedupAction, setDedupAction] = useState<string>('review')
  const [dedupLoading, setDedupLoading] = useState(false)
  const [dedupSaving, setDedupSaving] = useState(false)

  // Voice authentication settings
  const [voiceAuthEnabled, setVoiceAuthEnabled] = useState(false)
  const [voiceAuthLoading, setVoiceAuthLoading] = useState(false)
  const [voiceAuthSaving, setVoiceAuthSaving] = useState(false)

  // Page branding settings
  const [brandLogoUrl, setBrandLogoUrl] = useState<string>('')
  const [brandAccentColor, setBrandAccentColor] = useState<string>('')
  const [brandCompanyName, setBrandCompanyName] = useState<string>('')
  const [brandConfigured, setBrandConfigured] = useState(false)
  const [brandSaving, setBrandSaving] = useState(false)
  const [brandLoading, setBrandLoading] = useState(false)

  // Reviewer management
  const [reviewers, setReviewers] = useState<Array<{ id: string; email: string; name?: string; role?: string; status: string; invited_at: string; last_login_at?: string }>>([])
  const [reviewerEmail, setReviewerEmail] = useState('')
  const [reviewerName, setReviewerName] = useState('')
  const [reviewerRole, setReviewerRole] = useState<'reviewer' | 'admin'>('reviewer')
  const [reviewerInviting, setReviewerInviting] = useState(false)

  // Danger zone
  const [showDeleteAccount, setShowDeleteAccount] = useState(false)
  const [deleteAccountEmail, setDeleteAccountEmail] = useState('')
  const [deleteAccountLoading, setDeleteAccountLoading] = useState(false)

  // System / Version info
  const [versionInfo, setVersionInfo] = useState<{ current_version: string; latest_version: string | null; update_available: boolean; release_url: string | null; watchtower?: { configured: boolean; running: boolean; containers_scanned?: number | null; containers_updated?: number | null; containers_failed?: number | null } } | null>(null)
  const [versionLoading, setVersionLoading] = useState(false)
  const [versionFetched, setVersionFetched] = useState(false)
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)

  // Active settings tab
  const [activeTab, setActiveTab] = useState<SettingsTab>('profile')

  // Fetch data on mount
  useEffect(() => {
    if (!token) return
    fetchProfile()
    fetchLLMSettings()
    fetchSMSSettings()
    fetchAmlSettings()
    fetchDedupSettings()
    fetchVoiceAuthSettings()
    fetchBrandingSettings()
    fetchReviewers()
  }, [token])

  const fetchVersionInfo = async () => {
    if (versionFetched) return
    setVersionLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/system/version`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const data = await res.json()
        setVersionInfo(data)
      }
    } catch { /* network error */ }
    setVersionLoading(false)
    setVersionFetched(true)
  }

  // Lazy fetch version info when System tab is selected
  useEffect(() => {
    if (activeTab === 'system' && !versionFetched) {
      fetchVersionInfo()
    }
  }, [activeTab])

  const copyCommand = (cmd: string, label: string) => {
    navigator.clipboard.writeText(cmd).then(() => {
      setCopiedCommand(label)
      toast.success(t('developer.common.copied'))
      setTimeout(() => setCopiedCommand(null), 2000)
    }).catch(() => {
      toast.error(t('developer.common.copyFailed'))
    })
  }

  const fetchProfile = async () => {
    setProfileLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/profile`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const { data } = await res.json()
        setProfileName(data.name || '')
        setProfileCompany(data.company || '')
        setProfileEmail(data.email || '')
        setProfileAvatarUrl(data.avatar_url || '')
      }
    } catch { /* network error */ }
    setProfileLoading(false)
  }

  const saveProfile = async () => {
    if (!token) return
    setProfileSaving(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/profile`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() }, credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ name: profileName, company: profileCompany || null }),
      })
      if (res.ok) {
        toast.success(t('settings.toast.profileUpdated'))
      } else {
        const err = await res.json().catch(() => ({}))
        toast.error(err.message || t('settings.toast.profileFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setProfileSaving(false)
  }

  const uploadAvatar = async (file: File) => {
    if (!token) return
    const formData = new FormData()
    formData.append('file', file)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/avatar`, {
        method: 'POST',
        headers: { ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
        body: formData,
      })
      if (res.ok) {
        const { data } = await res.json()
        setProfileAvatarUrl(data.avatar_url)
        toast.success(t('settings.toast.avatarUpdated'))
      } else {
        const err = await res.json().catch(() => ({}))
        toast.error(err.message || t('settings.toast.avatarFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
  }

  const fetchLLMSettings = async () => {
    setLlmLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/llm`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const data = await res.json()
        setLlmConfigured(data.configured)
        setLlmProvider(data.provider || '')
        setLlmKeyPreview(data.api_key_preview || '')
        setLlmEndpointUrl(data.endpoint_url || '')
        setLlmModel(data.model || '')
        setLlmApiKey('')
        setShowLlmKey(false)
      }
    } catch { /* network error */ }
    setLlmLoading(false)
  }

  const saveLLMSettings = async () => {
    if (!token) return
    setLlmSaving(true)
    try {
      const body: Record<string, string | null> = { provider: llmProvider || null }
      if (llmApiKey) body.api_key = llmApiKey
      if (llmProvider === 'custom') body.endpoint_url = llmEndpointUrl || null
      body.model = llmModel || null
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/llm`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() }, credentials: 'include' as RequestCredentials,
        body: JSON.stringify(body),
      })
      if (res.ok) {
        toast.success(llmProvider ? t('settings.toast.llmSaved') : t('settings.toast.llmCleared'))
        fetchLLMSettings()
      } else {
        const err = await res.json().catch(() => ({ error: 'Failed to save' }))
        toast.error(err.error || t('settings.toast.llmSaveFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setLlmSaving(false)
  }

  const clearLLMSettings = async () => {
    if (!token) return
    setLlmSaving(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/llm`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() }, credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ provider: null }),
      })
      if (res.ok) {
        toast.success(t('settings.toast.llmCleared'))
        setLlmProvider('')
        setLlmApiKey('')
        setLlmEndpointUrl('')
        setLlmKeyPreview('')
        setLlmConfigured(false)
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setLlmSaving(false)
  }

  const fetchSMSSettings = async () => {
    setSmsLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/sms`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const data = await res.json()
        setSmsConfigured(data.configured)
        setSmsProvider(data.provider || '')
        setSmsKeyPreview(data.api_key_preview || '')
        setSmsPhoneNumber(data.phone_number || '')
        setSmsApiKey('')
        setSmsApiSecret('')
        setShowSmsKey(false)
      }
    } catch { /* network error */ }
    setSmsLoading(false)
  }

  const saveSMSSettings = async () => {
    if (!token) return
    setSmsSaving(true)
    try {
      const body: Record<string, string | null> = { provider: smsProvider || null }
      if (smsApiKey) body.api_key = smsApiKey
      if (smsApiSecret) body.api_secret = smsApiSecret
      if (smsPhoneNumber) body.phone_number = smsPhoneNumber
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/sms`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() }, credentials: 'include' as RequestCredentials,
        body: JSON.stringify(body),
      })
      if (res.ok) {
        toast.success(smsProvider ? t('settings.toast.smsSaved') : t('settings.toast.smsCleared'))
        fetchSMSSettings()
      } else {
        const err = await res.json().catch(() => ({ error: 'Failed to save' }))
        toast.error(err.error || err.message || t('settings.toast.smsSaveFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setSmsSaving(false)
  }

  const clearSMSSettings = async () => {
    if (!token) return
    setSmsSaving(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/sms`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() }, credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ provider: null }),
      })
      if (res.ok) {
        toast.success(t('settings.toast.smsCleared'))
        setSmsProvider('')
        setSmsApiKey('')
        setSmsApiSecret('')
        setSmsPhoneNumber('')
        setSmsKeyPreview('')
        setSmsConfigured(false)
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setSmsSaving(false)
  }

  const fetchAmlSettings = async () => {
    setAmlLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/aml`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const data = await res.json()
        setAmlEnabled(data.enabled ?? true)
      }
    } catch { /* network error */ }
    setAmlLoading(false)
  }

  const saveAmlSettings = async () => {
    if (!token) return
    setAmlSaving(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/aml`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ enabled: amlEnabled }),
      })
      if (res.ok) {
        toast.success(t('settings.toast.amlSaved'))
      } else {
        const err = await res.json().catch(() => ({}))
        toast.error(err.error || t('settings.toast.saveFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setAmlSaving(false)
  }

  const fetchDedupSettings = async () => {
    setDedupLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/duplicate-detection`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const data = await res.json()
        setDedupEnabled(data.enabled ?? false)
        setDedupAction(data.action ?? 'review')
      }
    } catch { /* network error */ }
    setDedupLoading(false)
  }

  const saveDedupSettings = async () => {
    if (!token) return
    setDedupSaving(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/duplicate-detection`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ enabled: dedupEnabled, action: dedupAction }),
      })
      if (res.ok) {
        toast.success(t('settings.toast.duplicateSaved'))
      } else {
        const err = await res.json().catch(() => ({}))
        toast.error(err.error || t('settings.toast.saveFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setDedupSaving(false)
  }

  const fetchVoiceAuthSettings = async () => {
    setVoiceAuthLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/voice-auth`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const data = await res.json()
        setVoiceAuthEnabled(data.enabled ?? false)
      }
    } catch { /* network error */ }
    setVoiceAuthLoading(false)
  }

  const saveVoiceAuthSettings = async () => {
    if (!token) return
    setVoiceAuthSaving(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/voice-auth`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ enabled: voiceAuthEnabled }),
      })
      if (res.ok) {
        toast.success(t('settings.toast.voiceSaved'))
      } else {
        const err = await res.json().catch(() => ({}))
        toast.error(err.error || t('settings.toast.saveFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setVoiceAuthSaving(false)
  }

  const fetchBrandingSettings = async () => {
    setBrandLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/branding`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const data = await res.json()
        setBrandConfigured(data.configured)
        setBrandLogoUrl(data.logo_url || '')
        setBrandAccentColor(data.accent_color || '')
        setBrandCompanyName(data.company_name || '')
      }
    } catch { /* network error */ }
    setBrandLoading(false)
  }

  const saveBrandingSettings = async () => {
    if (!token) return
    setBrandSaving(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/branding`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() }, credentials: 'include' as RequestCredentials,
        body: JSON.stringify({
          logo_url: brandLogoUrl || null,
          accent_color: brandAccentColor || null,
          company_name: brandCompanyName || null,
        }),
      })
      if (res.ok) {
        toast.success(t('settings.toast.brandingSaved'))
        fetchBrandingSettings()
      } else {
        const err = await res.json().catch(() => ({ error: 'Failed to save' }))
        toast.error(err.error || err.message || t('settings.toast.brandingSaveFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setBrandSaving(false)
  }

  const clearBrandingSettings = async () => {
    if (!token) return
    setBrandSaving(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/settings/branding`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() }, credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ logo_url: null, accent_color: null, company_name: null }),
      })
      if (res.ok) {
        toast.success(t('settings.toast.brandingCleared'))
        setBrandLogoUrl('')
        setBrandAccentColor('')
        setBrandCompanyName('')
        setBrandConfigured(false)
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setBrandSaving(false)
  }

  const uploadBrandingLogo = async (file: File) => {
    if (!token) return
    const formData = new FormData()
    formData.append('file', file)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/branding/logo`, {
        method: 'POST',
        headers: { ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
        body: formData,
      })
      if (res.ok) {
        const { data } = await res.json()
        setBrandLogoUrl(data.logo_url)
        toast.success(t('settings.toast.logoUploaded'))
      } else {
        const err = await res.json().catch(() => ({}))
        toast.error(err.message || t('settings.toast.logoFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
  }

  const fetchReviewers = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/reviewers`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const data = await res.json()
        setReviewers(data.reviewers ?? [])
      }
    } catch { /* network error */ }
  }

  const inviteReviewer = async () => {
    if (!token || !reviewerEmail) return
    setReviewerInviting(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/reviewers/invite`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() }, credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ email: reviewerEmail, name: reviewerName || undefined, role: reviewerRole }),
      })
      const data = await res.json()
      if (res.ok) {
        setReviewers(prev => [data.reviewer, ...prev])
        setReviewerEmail('')
        setReviewerName('')
        setReviewerRole('reviewer')
        toast.success(t('settings.toast.reviewerInvited'))
      } else {
        toast.error(data.message || t('settings.toast.reviewerInviteFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
    setReviewerInviting(false)
  }

  const revokeReviewer = async (id: string) => {
    if (!token) return
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/reviewers/${id}`, {
        method: 'DELETE',
        headers: { ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        setReviewers(prev => prev.map(r => r.id === id ? { ...r, status: 'revoked' } : r))
        toast.success(t('settings.toast.reviewerRevoked'))
      } else {
        const data = await res.json().catch(() => ({}))
        toast.error(data.message || t('settings.toast.reviewerRevokeFailed'))
      }
    } catch { toast.error(t('developer.common.networkError')) }
  }

  const deleteAccount = async () => {
    if (!token) return
    setDeleteAccountLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/account`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ confirm_email: deleteAccountEmail }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.message || 'Failed to delete account')
      fetch(`${API_BASE_URL}/api/auth/logout`, { method: 'POST', credentials: 'include', headers: { ...authHeaders, ...csrfHeader() } }).catch(() => {})
      clearCsrfToken()
      toast.success(t('settings.toast.accountDeleted'))
      onAccountDeleted()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('settings.toast.accountDeleteFailed'))
    } finally {
      setDeleteAccountLoading(false)
      setShowDeleteAccount(false)
      setDeleteAccountEmail('')
    }
  }

  const paneStyle: CSSProperties = {
    height: '100%',
    maxHeight: 'calc(80vh - 1px)',
    overflowY: 'auto',
    padding: '8px 32px 32px',
    boxSizing: 'border-box',
  }

  const tabLabel = (id: SettingsTab) => {
    const tab = tabs.find(item => item.id === id)!
    const Icon = tab.icon
    const danger = id === 'account'
    return (
      <span style={{ display: 'flex', alignItems: 'center', gap: 10, fontFamily: T.mono, fontSize: 12, color: danger ? T.red : undefined }}>
        <Icon style={{ width: 16, height: 16, flexShrink: 0, color: danger ? T.red : undefined }} />
        {t(tab.label)}
      </span>
    )
  }

  const beforeUploadFile = (file: File, upload: (f: File) => void) => {
    if (file.size > 2 * 1024 * 1024) {
      toast.error(t('settings.toast.fileTooLarge'))
      return Upload.LIST_IGNORE
    }
    upload(file)
    return Upload.LIST_IGNORE
  }

  return (
    <>
      <Modal
        open
        onCancel={onClose}
        footer={null}
        width={1200}
        centered
        title={
          <span style={{ display: 'flex', alignItems: 'center', gap: 8, fontFamily: T.mono, fontSize: 14, fontWeight: 600, color: T.text }}>
            <Cog6ToothIcon style={{ width: 18, height: 18 }} />
            {t('settings.title')}
          </span>
        }
        styles={{ body: { height: '80vh', padding: 0, overflow: 'hidden' } }}
      >
        <Tabs
          tabPlacement="start"
          activeKey={activeTab}
          onChange={key => setActiveTab(key as SettingsTab)}
          style={{ height: '100%' }}
          tabBarStyle={{ width: 220, flexShrink: 0, paddingTop: 8, margin: 0, background: T.surface, borderRight: `1px solid ${T.border}` }}
          items={[
            {
              key: 'profile',
              label: tabLabel('profile'),
              children: activeTab === 'profile' && (
                <div style={paneStyle}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 16 }}>
                    <UserCircleIcon style={{ width: 16, height: 16, color: T.accent }} />
                    <div style={{ fontFamily: T.mono, fontWeight: 600, fontSize: 13, color: T.text }}>{t('settings.profile.heading')}</div>
                  </div>

                  {profileLoading ? (
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: T.muted, fontSize: 13 }}>
                      <Spin size="small" />
                      <span>{t('settings.profile.loading')}</span>
                    </div>
                  ) : (
                    <>
                      {/* Avatar */}
                      <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 16 }}>
                        <Upload
                          accept="image/jpeg,image/png"
                          showUploadList={false}
                          beforeUpload={file => beforeUploadFile(file, uploadAvatar)}
                        >
                          <div style={{ position: 'relative', width: 48, height: 48, borderRadius: '50%', overflow: 'hidden', cursor: 'pointer', flexShrink: 0, border: `1px solid ${T.border}` }}>
                            {profileAvatarUrl ? (
                              <img src={profileAvatarUrl} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                            ) : (
                              <div style={{ width: '100%', height: '100%', background: T.surface, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                                <UserCircleIcon style={{ width: 28, height: 28, color: T.dim }} />
                              </div>
                            )}
                          </div>
                        </Upload>
                        <div style={{ flex: 1 }}>
                          <div style={{ fontSize: 12, color: T.muted, marginBottom: 2 }}>{t('settings.profile.avatarHint')}</div>
                          <div style={{ fontSize: 11, color: T.dim }}>{t('settings.profile.avatarFormat')}</div>
                        </div>
                        <Upload
                          accept="image/jpeg,image/png"
                          showUploadList={false}
                          beforeUpload={file => beforeUploadFile(file, uploadAvatar)}
                        >
                          <Button size="small">{t('settings.profile.change')}</Button>
                        </Upload>
                      </div>

                      {/* Email (read-only) */}
                      <label style={labelStyle}>{t('settings.profile.email')}</label>
                      <Input type="email" value={profileEmail} readOnly style={{ marginBottom: 12 }} />

                      {/* Name */}
                      <label style={labelStyle}>{t('settings.profile.name')}</label>
                      <Input
                        value={profileName}
                        onChange={e => setProfileName(e.target.value)}
                        style={{ marginBottom: 12 }}
                        placeholder={t('settings.profile.namePlaceholder')}
                      />

                      {/* Company */}
                      <label style={labelStyle}>{t('settings.profile.company')} <span style={{ color: T.dim, fontWeight: 400 }}>{t('settings.profile.optional')}</span></label>
                      <Input
                        value={profileCompany}
                        onChange={e => setProfileCompany(e.target.value)}
                        style={{ marginBottom: 12 }}
                        placeholder={t('settings.profile.companyPlaceholder')}
                      />

                      {/* Save button */}
                      <Button type="primary" onClick={saveProfile} loading={profileSaving} disabled={!profileName.trim()}>
                        {profileSaving ? t('settings.profile.saving') : t('settings.profile.save')}
                      </Button>
                    </>
                  )}
                </div>
              )
            },
            {
              key: 'team',
              label: tabLabel('team'),
              children: activeTab === 'team' && (
                <div style={paneStyle}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                    <UsersIcon style={{ width: 16, height: 16, color: T.accent }} />
                    <div style={{ fontFamily: T.mono, fontWeight: 600, fontSize: 13, color: T.text }}>{t('settings.team.heading')}</div>
                  </div>
                  <div style={{ color: T.muted, fontSize: 13, marginBottom: 16, lineHeight: 1.6 }}>
                    {t('settings.team.description')}
                    {' '}{t('settings.team.noPassword')}
                  </div>

                  {/* Invite form */}
                  <div style={{ display: 'flex', gap: 8, marginBottom: 16, flexWrap: 'wrap' }}>
                    <Input
                      type="email"
                      value={reviewerEmail}
                      onChange={e => setReviewerEmail(e.target.value)}
                      placeholder={t('settings.team.emailPlaceholder')}
                      style={{ flex: '1 1 160px' }}
                    />
                    <Input
                      value={reviewerName}
                      onChange={e => setReviewerName(e.target.value)}
                      placeholder={t('settings.team.namePlaceholder')}
                      style={{ flex: '0 1 120px' }}
                    />
                    <Select
                      value={reviewerRole}
                      onChange={value => setReviewerRole(value as 'reviewer' | 'admin')}
                      options={[
                        { value: 'reviewer', label: t('settings.team.reviewer') },
                        { value: 'admin', label: t('settings.team.admin') },
                      ]}
                      style={{ flex: '0 0 140px' }}
                    />
                    <Button
                      type="primary"
                      onClick={inviteReviewer}
                      loading={reviewerInviting}
                      disabled={!reviewerEmail.includes('@')}
                    >
                      {reviewerInviting ? t('settings.team.inviting') : t('settings.team.invite')}
                    </Button>
                  </div>

                  {/* Reviewers list */}
                  {reviewers.length > 0 && (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                      {reviewers.map(r => (
                        <div key={r.id} style={{
                          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                          background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0, padding: '8px 12px',
                          opacity: r.status === 'revoked' ? 0.45 : 1,
                        }}>
                          <div style={{ flex: 1, minWidth: 0 }}>
                            <div style={{ fontSize: 13, color: T.text, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                              {r.email}
                              {r.name && <span style={{ color: T.dim, marginLeft: 6 }}>({r.name})</span>}
                            </div>
                            <div style={{ fontSize: 11, color: T.dim, marginTop: 2 }}>
                              {r.status === 'invited' && t('settings.team.invited')}
                              {r.status === 'active' && (r.last_login_at ? t('settings.team.lastLogin', { time: formatDate(r.last_login_at) }) : t('settings.team.active'))}
                              {r.status === 'revoked' && t('settings.team.revoked')}
                            </div>
                          </div>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexShrink: 0, marginLeft: 8 }}>
                            <Tag color={r.role === 'admin' ? 'purple' : 'default'} style={{ marginInlineEnd: 0 }}>
                              {r.role === 'admin' ? t('settings.team.admin') : t('settings.team.reviewer')}
                            </Tag>
                            <Tag
                              color={r.status === 'active' ? 'green' : r.status === 'invited' ? 'cyan' : 'red'}
                              style={{ marginInlineEnd: 0 }}
                            >
                              {r.status === 'active' ? t('settings.team.active') : r.status === 'invited' ? t('settings.team.invited') : t('settings.team.revoked')}
                            </Tag>
                            {r.status !== 'revoked' && (
                              <Button
                                danger
                                type="text"
                                size="small"
                                onClick={() => revokeReviewer(r.id)}
                                title={t('settings.revokeAccess')}
                                icon={<XMarkIcon style={{ width: 14, height: 14 }} />}
                              />
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  )}

                  {reviewers.length === 0 && (
                    <div style={{ color: T.dim, fontSize: 12, textAlign: 'center', padding: '8px 0' }}>
                      {t('settings.team.noReviewers')}
                    </div>
                  )}

                  {/* Copy login link */}
                  <Button
                    onClick={() => {
                      const url = `${window.location.origin}/admin/login`
                      navigator.clipboard.writeText(url).then(() => toast.success(t('settings.toast.loginLinkCopied')))
                    }}
                    style={{ marginTop: 12 }}
                  >
                    {t('settings.team.copyLoginLink')}
                  </Button>
                </div>
              )
            },
            {
              key: 'integrations',
              label: tabLabel('integrations'),
              children: activeTab === 'integrations' && (
                <div style={paneStyle}>
                  {/* OCR Enhancement (LLM Fallback) */}
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                      <CodeBracketIcon style={{ width: 16, height: 16, color: T.accent }} />
                      <div style={{ fontFamily: T.mono, fontWeight: 600, fontSize: 13, color: T.text }}>{t('settings.sections.ocr')}</div>
                    </div>
                    <div style={{ color: T.muted, fontSize: 13, marginBottom: 16, lineHeight: 1.6 }}>
                      {t('settings.integrations.ocrDesc')}
                    </div>

                    {llmLoading ? (
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: T.muted, fontSize: 13 }}>
                        <Spin size="small" />
                        <span>{t('settings.profile.loading')}</span>
                      </div>
                    ) : (
                      <>
                        {/* Provider select */}
                        <label style={labelStyle}>{t('settings.integrations.provider')}</label>
                        <Select
                          value={llmProvider}
                          onChange={value => { setLlmProvider(value); setLlmApiKey(''); setShowLlmKey(false) }}
                          options={[
                            { value: '', label: t('settings.integrations.none') },
                            { value: 'openai', label: t('settings.integrations.openai') },
                            { value: 'anthropic', label: t('settings.integrations.anthropic') },
                            { value: 'custom', label: t('settings.integrations.custom') },
                          ]}
                          style={{ width: '100%', marginBottom: 12 }}
                        />

                        {llmProvider && (
                          <>
                            {/* API Key */}
                            <label style={labelStyle}>
                              {t('settings.integrations.apiKey')}
                              {llmConfigured && llmKeyPreview && !llmApiKey && (
                                <span style={{ color: T.green, marginLeft: 8, fontWeight: 400 }}>
                                  {t('settings.integrations.configured', { preview: llmKeyPreview })}
                                </span>
                              )}
                            </label>
                            <Input.Password
                              value={llmApiKey}
                              onChange={e => setLlmApiKey(e.target.value)}
                              visibilityToggle={{ visible: showLlmKey, onVisibleChange: setShowLlmKey }}
                              placeholder={llmConfigured ? t('settings.integrations.apiKeyPlaceholder') : llmProvider === 'anthropic' ? 'sk-ant-...' : 'sk-...'}
                              style={{ marginBottom: 12 }}
                            />

                            {/* Custom endpoint URL */}
                            {llmProvider === 'custom' && (
                              <>
                                <label style={labelStyle}>{t('settings.integrations.endpointUrl')}</label>
                                <Input
                                  type="url"
                                  style={{ marginBottom: 12 }}
                                  value={llmEndpointUrl}
                                  onChange={e => setLlmEndpointUrl(e.target.value)}
                                  placeholder="https://your-server.com/v1/chat/completions"
                                />
                              </>
                            )}

                            {/* Model — required for custom endpoints, optional
                                override for the built-in providers */}
                            <label style={labelStyle}>
                              {t('settings.integrations.modelLabel')}{llmProvider === 'custom' ? '' : t('settings.integrations.modelOptional')}
                            </label>
                            <Input
                              style={{ marginBottom: 12 }}
                              value={llmModel}
                              onChange={e => setLlmModel(e.target.value)}
                              placeholder={llmProvider === 'anthropic' ? 'claude-sonnet-4-20250514' : llmProvider === 'openai' ? 'gpt-4o' : 'gemini-2.5-flash'}
                            />

                            {/* Save / Clear buttons */}
                            <div style={{ display: 'flex', gap: 10, marginTop: 4 }}>
                              <Button
                                type="primary"
                                onClick={saveLLMSettings}
                                loading={llmSaving}
                                disabled={(!llmApiKey && !llmConfigured) || (llmProvider === 'custom' && !llmModel)}
                              >
                                {llmSaving ? t('settings.saving') : t('settings.save')}
                              </Button>
                              {llmConfigured && (
                                <Button danger onClick={clearLLMSettings} disabled={llmSaving}>
                                  Remove
                                </Button>
                              )}
                            </div>
                          </>
                        )}
                      </>
                    )}
                  </div>

                  {/* Divider between OCR and SMS */}
                  <div style={{ borderTop: `1px solid ${T.border}`, marginTop: 24, paddingTop: 20 }} />

                  {/* SMS Provider (Phone OTP) */}
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                      <DevicePhoneMobileIcon style={{ width: 16, height: 16, color: T.accent }} />
                      <div style={{ fontFamily: T.mono, fontWeight: 600, fontSize: 13, color: T.text }}>{t('settings.sections.phoneOtp')}</div>
                    </div>
                    <div style={{ color: T.muted, fontSize: 13, marginBottom: 16, lineHeight: 1.6 }}>
                      {t('settings.integrations.phoneDesc')}
                    </div>

                    {smsLoading ? (
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: T.muted, fontSize: 13 }}>
                        <Spin size="small" />
                        <span>{t('settings.profile.loading')}</span>
                      </div>
                    ) : (
                      <>
                        {/* Provider select */}
                        <label style={labelStyle}>{t('settings.integrations.provider')}</label>
                        <Select
                          value={smsProvider}
                          onChange={value => { setSmsProvider(value); setSmsApiKey(''); setSmsApiSecret(''); setShowSmsKey(false) }}
                          options={[
                            { value: '', label: t('settings.integrations.none') },
                            { value: 'twilio', label: 'Twilio' },
                            { value: 'vonage', label: 'Vonage' },
                          ]}
                          style={{ width: '100%', marginBottom: 12 }}
                        />

                        {smsProvider && (
                          <>
                            {/* API Key / Account SID */}
                            <label style={labelStyle}>
                              {smsProvider === 'twilio' ? 'Account SID' : 'API Key'}
                              {smsConfigured && smsKeyPreview && !smsApiKey && (
                                <span style={{ color: T.green, marginLeft: 8, fontWeight: 400 }}>
                                  configured: {smsKeyPreview}
                                </span>
                              )}
                            </label>
                            <Input.Password
                              value={smsApiKey}
                              onChange={e => setSmsApiKey(e.target.value)}
                              visibilityToggle={{ visible: showSmsKey, onVisibleChange: setShowSmsKey }}
                              placeholder={smsConfigured ? t('settings.integrations.apiKeyPlaceholder') : smsProvider === 'twilio' ? 'AC...' : 'API key'}
                              style={{ marginBottom: 12 }}
                            />

                            {/* API Secret / Auth Token */}
                            <label style={labelStyle}>{smsProvider === 'twilio' ? t('settings.integrations.authToken') : t('settings.integrations.apiSecret')}</label>
                            <Input.Password
                              style={{ marginBottom: 12 }}
                              value={smsApiSecret}
                              onChange={e => setSmsApiSecret(e.target.value)}
                              placeholder={smsConfigured ? 'Enter new secret to replace' : smsProvider === 'twilio' ? 'Auth token' : 'API secret'}
                            />

                            {/* Phone Number */}
                            <label style={labelStyle}>{t('settings.integrations.senderPhone')}</label>
                            <Input
                              type="tel"
                              style={{ marginBottom: 12 }}
                              value={smsPhoneNumber}
                              onChange={e => setSmsPhoneNumber(e.target.value)}
                              placeholder="+15551234567"
                            />

                            {/* Save / Clear buttons */}
                            <div style={{ display: 'flex', gap: 10, marginTop: 4 }}>
                              <Button
                                type="primary"
                                onClick={saveSMSSettings}
                                loading={smsSaving}
                                disabled={!smsConfigured && (!smsApiKey || !smsApiSecret || !smsPhoneNumber)}
                              >
                                {smsSaving ? t('settings.saving') : t('settings.save')}
                              </Button>
                              {smsConfigured && (
                                <Button danger onClick={clearSMSSettings} disabled={smsSaving}>
                                  Remove
                                </Button>
                              )}
                            </div>
                          </>
                        )}
                      </>
                    )}
                  </div>

                  {/* Divider between SMS and AML */}
                  <div style={{ borderTop: `1px solid ${T.border}`, marginTop: 24, paddingTop: 20 }} />

                  {/* AML / Sanctions Screening */}
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                      <ShieldExclamationIcon style={{ width: 16, height: 16, color: T.accent }} />
                      <div style={{ fontFamily: T.mono, fontWeight: 600, fontSize: 13, color: T.text }}>{t('settings.sections.aml')}</div>
                    </div>
                    <div style={{ color: T.muted, fontSize: 13, marginBottom: 16, lineHeight: 1.6 }}>
                      {t('settings.integrations.amlDesc')}
                    </div>

                    {amlLoading ? (
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: T.muted, fontSize: 13 }}>
                        <Spin size="small" />
                        <span>{t('settings.profile.loading')}</span>
                      </div>
                    ) : (
                      <>
                        {/* Enable toggle */}
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
                          <Switch checked={amlEnabled} onChange={setAmlEnabled} />
                          <span style={{ color: T.text, fontSize: 13 }}>
                            {amlEnabled ? t('settings.enabled') : t('settings.disabled')}
                          </span>
                        </div>

                        {!amlEnabled && (
                          <Alert type="warning" showIcon title={t('settings.amlDisabled')} style={{ marginBottom: 12 }} />
                        )}

                        <Button type="primary" onClick={saveAmlSettings} loading={amlSaving}>
                          {amlSaving ? t('settings.saving') : t('settings.save')}
                        </Button>
                      </>
                    )}
                  </div>

                  {/* Divider between AML and Duplicate Detection */}
                  <div style={{ borderTop: `1px solid ${T.border}`, marginTop: 24, paddingTop: 20 }} />

                  {/* Duplicate Detection */}
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                      <FingerPrintIcon style={{ width: 16, height: 16, color: T.accent }} />
                      <div style={{ fontFamily: T.mono, fontWeight: 600, fontSize: 13, color: T.text }}>{t('settings.sections.duplicate')}</div>
                    </div>
                    <div style={{ color: T.muted, fontSize: 13, marginBottom: 16, lineHeight: 1.6 }}>
                      {t('settings.integrations.dupDesc')}
                    </div>

                    {dedupLoading ? (
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: T.muted, fontSize: 13 }}>
                        <Spin size="small" />
                        <span>{t('settings.profile.loading')}</span>
                      </div>
                    ) : (
                      <>
                        {/* Enable toggle */}
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
                          <Switch checked={dedupEnabled} onChange={setDedupEnabled} />
                          <span style={{ color: T.text, fontSize: 13 }}>
                            {dedupEnabled ? t('settings.enabled') : t('settings.disabled')}
                          </span>
                        </div>

                        {dedupEnabled && (
                          <>
                            <label style={labelStyle}>{t('settings.integrations.action')}</label>
                            <Select
                              value={dedupAction}
                              onChange={value => setDedupAction(value)}
                              options={[
                                { value: 'review', label: t('settings.integrations.review') },
                                { value: 'block', label: t('settings.integrations.block') },
                                { value: 'allow', label: t('settings.integrations.allow') },
                              ]}
                              style={{ width: '100%', marginBottom: 16 }}
                            />
                          </>
                        )}

                        <Button type="primary" onClick={saveDedupSettings} loading={dedupSaving}>
                          {dedupSaving ? t('settings.saving') : t('settings.save')}
                        </Button>
                      </>
                    )}
                  </div>

                  {/* Divider between Duplicate Detection and Voice Auth */}
                  <div style={{ borderTop: `1px solid ${T.border}`, marginTop: 24, paddingTop: 20 }} />

                  {/* Voice Authentication */}
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                      <MicrophoneIcon style={{ width: 16, height: 16, color: T.accent }} />
                      <div style={{ fontFamily: T.mono, fontWeight: 600, fontSize: 13, color: T.text }}>{t('settings.sections.voice')}</div>
                    </div>
                    <div style={{ color: T.muted, fontSize: 13, marginBottom: 16, lineHeight: 1.6 }}>
                      {t('settings.integrations.voiceDesc')}
                    </div>

                    {voiceAuthLoading ? (
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: T.muted, fontSize: 13 }}>
                        <Spin size="small" />
                        <span>{t('settings.profile.loading')}</span>
                      </div>
                    ) : (
                      <>
                        {/* Enable toggle */}
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
                          <Switch checked={voiceAuthEnabled} onChange={setVoiceAuthEnabled} />
                          <span style={{ color: T.text, fontSize: 13 }}>
                            {voiceAuthEnabled ? t('settings.enabled') : t('settings.disabled')}
                          </span>
                        </div>

                        {!voiceAuthEnabled && (
                          <Alert type="warning" showIcon title={t('settings.voiceDisabled')} style={{ marginBottom: 12 }} />
                        )}

                        <Button type="primary" onClick={saveVoiceAuthSettings} loading={voiceAuthSaving}>
                          {voiceAuthSaving ? t('settings.saving') : t('settings.save')}
                        </Button>
                      </>
                    )}
                  </div>
                </div>
              )
            },
            {
              key: 'branding',
              label: tabLabel('branding'),
              children: activeTab === 'branding' && (
                <div style={paneStyle}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                    <PaintBrushIcon style={{ width: 16, height: 16, color: T.accent }} />
                    <div style={{ fontFamily: T.mono, fontWeight: 600, fontSize: 13, color: T.text }}>{t('settings.sections.page')}</div>
                  </div>
                  <div style={{ color: T.muted, fontSize: 13, marginBottom: 16, lineHeight: 1.6 }}>
                    {t('settings.branding.description')}
                  </div>

                  {brandLoading ? (
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: T.muted, fontSize: 13 }}>
                      <Spin size="small" />
                      <span>{t('settings.profile.loading')}</span>
                    </div>
                  ) : (
                    <>
                      {/* Company Name */}
                      <label style={labelStyle}>{t('settings.branding.companyName')}</label>
                      <Input
                        value={brandCompanyName}
                        onChange={e => setBrandCompanyName(e.target.value)}
                        style={{ marginBottom: 12 }}
                        placeholder={t('settings.branding.companyPlaceholder')}
                        maxLength={100}
                      />

                      {/* Logo URL + upload */}
                      <label style={labelStyle}>{t('settings.branding.logoUrl')}</label>
                      <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
                        <Input
                          type="url"
                          value={brandLogoUrl}
                          onChange={e => setBrandLogoUrl(e.target.value)}
                          style={{ flex: 1 }}
                          placeholder={t('settings.branding.logoPlaceholder')}
                        />
                        <Upload
                          accept="image/jpeg,image/png"
                          showUploadList={false}
                          beforeUpload={file => beforeUploadFile(file, uploadBrandingLogo)}
                        >
                          <Button>Upload</Button>
                        </Upload>
                      </div>

                      {/* Accent Color */}
                      <label style={labelStyle}>{t('settings.branding.accent')}</label>
                      <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 12 }}>
                        <Input
                          value={brandAccentColor}
                          onChange={e => setBrandAccentColor(e.target.value)}
                          style={{ flex: 1 }}
                          placeholder="#22d3ee"
                          maxLength={7}
                        />
                        <ColorPicker
                          value={/^#[0-9a-fA-F]{6}$/.test(brandAccentColor) ? brandAccentColor : '#22d3ee'}
                          onChange={color => setBrandAccentColor(color.toHexString())}
                          showText={false}
                        />
                      </div>

                      {/* Live Preview */}
                      <div style={{
                        background: T.bg, border: `1px solid ${T.border}`, borderRadius: 0,
                        padding: 16, marginBottom: 14,
                      }}>
                        <div style={{ fontSize: 10, color: T.dim, textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 10 }}>{t('settings.branding.preview')}</div>
                        <div style={{ textAlign: 'center' }}>
                          {brandLogoUrl ? (
                            <img
                              src={brandLogoUrl}
                              alt="Logo preview"
                              style={{ height: 28, margin: '0 auto 10px', display: 'block', objectFit: 'contain' }}
                              onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
                            />
                          ) : (
                            <img src="/identity-logo.svg" alt="Identity Verification" style={{ height: 28, margin: '0 auto 10px', display: 'block' }} />
                          )}
                          <div style={{ fontSize: 14, fontWeight: 600, color: '#dde2ec', marginBottom: 4 }}>
                            {brandCompanyName ? `Verify with ${brandCompanyName}` : 'Verify Your Identity'}
                          </div>
                          <div style={{ fontSize: 11, color: '#8896aa', marginBottom: 12 }}>{t('settings.branding.chooseCompletion')}</div>
                          <Button
                            style={{
                              background: /^#[0-9a-fA-F]{6}$/.test(brandAccentColor) ? brandAccentColor : T.accent,
                              color: T.bg, border: 'none', borderRadius: 0,
                              padding: '8px 24px', fontSize: 12, fontWeight: 600,
                              fontFamily: T.mono, cursor: 'default',
                            }}
                          >
                            {t('settings.branding.scanQr')}
                          </Button>
                          {(brandLogoUrl || brandCompanyName || brandAccentColor) && (
                            <div style={{ fontSize: 9, color: T.dim, marginTop: 10 }}>
                              {t('settings.branding.poweredBy')}
                            </div>
                          )}
                        </div>
                      </div>

                      {/* Save / Clear buttons */}
                      <div style={{ display: 'flex', gap: 10, marginTop: 4 }}>
                        <Button type="primary" onClick={saveBrandingSettings} loading={brandSaving}>
                          {brandSaving ? t('settings.saving') : t('settings.save')}
                        </Button>
                        {brandConfigured && (
                          <Button danger onClick={clearBrandingSettings} disabled={brandSaving}>
                            {t('settings.integrations.clear')}
                          </Button>
                        )}
                      </div>

                      <Alert
                        type="info"
                        showIcon
                        style={{ marginTop: 16 }}
                        title={
                          <Link to="/developer/page-builder" onClick={onClose} style={{ color: T.accent, fontSize: 13, fontFamily: T.mono, fontWeight: 500 }}>
                            {t('settings.branding.moreControl')}
                          </Link>
                        }
                      />
                    </>
                  )}
                </div>
              )
            },
            {
              key: 'system',
              label: tabLabel('system'),
              children: activeTab === 'system' && (
                <div style={paneStyle}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                    <ServerIcon style={{ width: 16, height: 16, color: T.accent }} />
                    <div style={{ fontFamily: T.mono, fontWeight: 600, fontSize: 13, color: T.text }}>{t('settings.sections.system')}</div>
                  </div>
                  <div style={{ color: T.muted, fontSize: 13, marginBottom: 20, lineHeight: 1.6 }}>
                    {t('settings.system.description')}
                  </div>

                  {versionLoading ? (
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: T.muted, fontSize: 13 }}>
                      <Spin size="small" />
                      <span>{t('settings.system.loading')}</span>
                    </div>
                  ) : (
                    <>
                      {/* ── Version ── */}
                      <div style={{ background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0, padding: 20, marginBottom: 16 }}>
                        <div style={{ fontSize: 11, color: T.dim, textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 8 }}>{t('settings.system.current')}</div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
                          <span style={{ fontSize: 22, fontWeight: 700, color: T.text, fontFamily: T.mono }}>
                            v{versionInfo?.current_version || '...'}
                          </span>
                          {versionInfo?.update_available && (
                            <Tag color="green" style={{ marginInlineEnd: 0, textTransform: 'uppercase', fontFamily: T.mono }}>
                              {t('settings.system.updateAvailable')}
                            </Tag>
                          )}
                        </div>
                        {versionInfo?.latest_version && versionInfo.update_available && (
                          <div style={{ fontSize: 13, color: T.muted }}>
                            {t('settings.system.latestLabel')} <span style={{ color: T.green, fontFamily: T.mono }}>v{versionInfo.latest_version}</span>
                            {versionInfo.release_url && (
                              <>
                                {' \u00b7 '}
                                <a
                                  href={versionInfo.release_url}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  style={{ color: T.accent, textDecoration: 'none' }}
                                >
                                  {t('settings.system.releaseNotes')} &rarr;
                                </a>
                              </>
                            )}
                          </div>
                        )}
                        {versionInfo && !versionInfo.update_available && versionInfo.latest_version && (
                          <div style={{ fontSize: 13, color: T.dim }}>{t('settings.system.latest')}</div>
                        )}
                      </div>

                      {/* ── Update ── */}
                      <div style={{ background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0, padding: 20, marginBottom: 16 }}>
                        <div style={{ fontSize: 11, color: T.dim, textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 8 }}>{t('settings.system.update')}</div>
                        <div style={{ color: T.muted, fontSize: 13, marginBottom: 12, lineHeight: 1.6 }}>
                          {t('settings.system.updateScriptDesc')}
                        </div>
                        <Button
                          block
                          onClick={() => copyCommand('bash update.sh', 'update')}
                          style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', height: 'auto', padding: '10px 14px', marginBottom: 8 }}
                        >
                          <code style={{ fontFamily: T.mono, fontSize: 13, color: T.accent }}>bash update.sh</code>
                          <span style={{ fontSize: 11, color: copiedCommand === 'update' ? T.green : T.dim }}>
                            {copiedCommand === 'update' ? t('settings.system.copied') : t('settings.system.clickToCopy')}
                          </span>
                        </Button>
                        <div style={{ fontSize: 11, color: T.dim }}>
                          {t('settings.orManually')} <code style={{ fontFamily: T.mono, color: T.muted }}>docker compose pull && docker compose up -d</code>
                        </div>
                      </div>

                      {/* ── Auto-Update ── */}
                      <div style={{ background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0, padding: 20, marginBottom: 16 }}>
                        <div style={{ fontSize: 11, color: T.dim, textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 8 }}>{t('settings.system.autoUpdate')}</div>

                        {versionInfo?.watchtower?.configured ? (
                          versionInfo.watchtower.running ? (
                            <>
                              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
                                <div style={{ width: 8, height: 8, borderRadius: '50%', background: T.green, flexShrink: 0 }} />
                                <span style={{ fontSize: 13, color: T.text, fontWeight: 500 }}>{t('settings.system.watchtowerRunning')}</span>
                              </div>
                              <div style={{ color: T.muted, fontSize: 13, marginBottom: 8, lineHeight: 1.6 }}>
                                {t('settings.system.autoCheckDesc')}
                              </div>
                              {(versionInfo.watchtower.containers_scanned != null || versionInfo.watchtower.containers_updated != null) && (
                                <div style={{ display: 'flex', gap: 16, marginBottom: 8, flexWrap: 'wrap' }}>
                                  {versionInfo.watchtower.containers_scanned != null && (
                                    <div style={{ fontSize: 12, color: T.dim }}>
                                      {t('settings.system.scanned')} <span style={{ color: T.text, fontFamily: T.mono }}>{versionInfo.watchtower.containers_scanned}</span>
                                    </div>
                                  )}
                                  {versionInfo.watchtower.containers_updated != null && (
                                    <div style={{ fontSize: 12, color: T.dim }}>
                                      {t('settings.system.updated')} <span style={{ color: T.green, fontFamily: T.mono }}>{versionInfo.watchtower.containers_updated}</span>
                                    </div>
                                  )}
                                  {(versionInfo.watchtower.containers_failed ?? 0) > 0 && (
                                    <div style={{ fontSize: 12, color: T.dim }}>
                                      {t('settings.system.failed')} <span style={{ color: T.red, fontFamily: T.mono }}>{versionInfo.watchtower.containers_failed}</span>
                                    </div>
                                  )}
                                </div>
                              )}
                            </>
                          ) : (
                            <>
                              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
                                <div style={{ width: 8, height: 8, borderRadius: '50%', background: T.dim, flexShrink: 0 }} />
                                <span style={{ fontSize: 13, color: T.muted, fontWeight: 500 }}>{t('settings.system.watchtowerConfigured')}</span>
                              </div>
                              <div style={{ color: T.muted, fontSize: 13, marginBottom: 12, lineHeight: 1.6 }}>
                                {t('settings.system.sidecarStartDesc')}
                              </div>
                              <Button
                                block
                                onClick={() => copyCommand('docker compose --profile autoupdate up -d watchtower', 'watchtower-start')}
                                style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', height: 'auto', padding: '10px 14px' }}
                              >
                                <code style={{ fontFamily: T.mono, fontSize: 13, color: T.accent }}>docker compose --profile autoupdate up -d watchtower</code>
                                <span style={{ fontSize: 11, color: copiedCommand === 'watchtower-start' ? T.green : T.dim }}>
                                  {copiedCommand === 'watchtower-start' ? t('settings.system.copied') : t('settings.system.clickToCopy')}
                                </span>
                              </Button>
                            </>
                          )
                        ) : (
                          <>
                            <div style={{ color: T.muted, fontSize: 13, marginBottom: 12, lineHeight: 1.6 }}>
                              {t('settings.system.autoUpdateDesc')}
                            </div>
                            <Button
                              block
                              onClick={() => copyCommand('bash install.sh', 'autoupdate-enable')}
                              style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', height: 'auto', padding: '10px 14px' }}
                            >
                              <code style={{ fontFamily: T.mono, fontSize: 13, color: T.accent }}>bash install.sh</code>
                              <span style={{ fontSize: 11, color: copiedCommand === 'autoupdate-enable' ? T.green : T.dim }}>
                                {copiedCommand === 'autoupdate-enable' ? t('settings.system.copied') : t('settings.system.clickToCopy')}
                              </span>
                            </Button>
                            <div style={{ fontSize: 11, color: T.dim, marginTop: 6 }}>
                              {t('settings.system.rerunInstaller')}
                            </div>
                          </>
                        )}
                      </div>

                      {/* ── Uninstall ── */}
                      <Alert
                        type="error"
                        showIcon
                        title={t('settings.system.uninstall')}
                        description={
                          <>
                            <div style={{ marginBottom: 12 }}>{t('settings.system.uninstallDesc')}</div>
                            <Button
                              block
                              danger
                              onClick={() => copyCommand('bash uninstall.sh', 'uninstall')}
                              style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', height: 'auto', padding: '10px 14px' }}
                            >
                              <code style={{ fontFamily: T.mono, fontSize: 13, color: T.red }}>bash uninstall.sh</code>
                              <span style={{ fontSize: 11, color: copiedCommand === 'uninstall' ? T.green : T.dim }}>
                                {copiedCommand === 'uninstall' ? t('settings.system.copied') : t('settings.system.clickToCopy')}
                              </span>
                            </Button>
                          </>
                        }
                      />
                    </>
                  )}
                </div>
              )
            },
            {
              key: 'account',
              label: tabLabel('account'),
              children: activeTab === 'account' && (
                <div style={paneStyle}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                    <ExclamationTriangleIcon style={{ width: 16, height: 16, color: T.red }} />
                    <div style={{ fontWeight: 600, fontSize: 14, color: T.red }}>{t('settings.sections.danger')}</div>
                  </div>
                  <div style={{ color: T.muted, fontSize: 13, marginBottom: 16, lineHeight: 1.6 }}>
                    {t('settings.account.deleteDescription')}
                  </div>
                  <Button danger onClick={() => setShowDeleteAccount(true)}>
                    {t('settings.account.delete')}
                  </Button>
                </div>
              ),
            },
          ]}
        />
      </Modal>

      <Modal
        open={showDeleteAccount}
        onCancel={() => { setShowDeleteAccount(false); setDeleteAccountEmail('') }}
        footer={null}
        width={440}
        centered
        title={
          <span style={{ display: 'flex', alignItems: 'center', gap: 8, fontFamily: T.mono, fontSize: 14, fontWeight: 600, color: T.red }}>
            <ExclamationTriangleIcon style={{ width: 20, height: 20 }} />
            {t('settings.account.delete')}
          </span>
        }
      >
        <div style={{ color: T.muted, fontSize: 13, marginBottom: 16, lineHeight: 1.6 }}>
          {t('settings.account.deleteConfirm')}
        </div>
        <Input
          type="email"
          value={deleteAccountEmail}
          onChange={e => setDeleteAccountEmail(e.target.value)}
          placeholder={t('settings.account.emailPlaceholder')}
          autoFocus
        />
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, marginTop: 16 }}>
          <Button onClick={() => { setShowDeleteAccount(false); setDeleteAccountEmail('') }}>
            Cancel
          </Button>
          <Button
            danger
            type="primary"
            loading={deleteAccountLoading}
            onClick={deleteAccount}
            disabled={!deleteAccountEmail}
          >
            Delete My Account
          </Button>
        </div>
      </Modal>
    </>
  )
}
