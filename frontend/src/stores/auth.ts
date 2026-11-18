import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', () => {
  const developerToken = ref(localStorage.getItem('iv_developer_token') || '')
  const adminToken = ref(localStorage.getItem('iv_admin_token') || '')
  const apiKey = ref(localStorage.getItem('iv_api_key') || '')
  const sessionToken = ref(localStorage.getItem('iv_session_token') || '')
  const email = ref(localStorage.getItem('iv_email') || '')

  const isAdmin = computed(() => !!adminToken.value)
  const isDeveloper = computed(() => !!developerToken.value || !!apiKey.value)

  function setDeveloperToken(token: string, accountEmail = '') {
    developerToken.value = token
    email.value = accountEmail
    localStorage.setItem('iv_developer_token', token)
    if (accountEmail) localStorage.setItem('iv_email', accountEmail)
  }
  function setAdminToken(token: string, accountEmail = '') {
    adminToken.value = token
    email.value = accountEmail
    localStorage.setItem('iv_admin_token', token)
    if (accountEmail) localStorage.setItem('iv_email', accountEmail)
  }
  function setAPIKey(value: string) {
    apiKey.value = value
    localStorage.setItem('iv_api_key', value)
  }
  function setSessionToken(value: string) {
    sessionToken.value = value
    localStorage.setItem('iv_session_token', value)
  }
  function logout() {
    developerToken.value = ''
    adminToken.value = ''
    apiKey.value = ''
    sessionToken.value = ''
    email.value = ''
    ;['iv_developer_token', 'iv_admin_token', 'iv_api_key', 'iv_session_token', 'iv_email'].forEach((key) => localStorage.removeItem(key))
  }
  return { developerToken, adminToken, apiKey, sessionToken, email, isAdmin, isDeveloper, setDeveloperToken, setAdminToken, setAPIKey, setSessionToken, logout }
})
