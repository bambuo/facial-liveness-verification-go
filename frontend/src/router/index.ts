import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: () => import('../views/DeveloperPage.vue'), meta: { title: 'Developer' } },
    { path: '/developer', name: 'developer', component: () => import('../views/DeveloperPage.vue'), meta: { title: 'Developer' } },
    { path: '/developer/page-builder', name: 'page-builder', component: () => import('../views/PageBuilderPage.vue'), meta: { title: 'Page Builder' } },
    {
      path: '/',
      component: () => import('../layouts/PublicLayout.vue'),
      children: [
        { path: 'demo', name: 'demo', component: () => import('../views/DemoPage.vue'), meta: { title: 'Demo' } },
        { path: 'verify', redirect: '/demo' },
        { path: 'verify-credential', name: 'verify-credential', component: () => import('../views/VerifyCredentialPage.vue'), meta: { title: 'Verify Credential' } },
        { path: 'docs', name: 'docs', component: () => import('../views/DocsPage.vue'), meta: { title: 'Documentation' } },
        { path: 'docs/guides', name: 'docs-guides', component: () => import('../views/DocsGuides.vue'), meta: { title: 'Guides' } },
        { path: 'docs/sdk', name: 'docs-sdk', component: () => import('../views/DocsSdk.vue'), meta: { title: 'SDK' } },
        { path: 'docs/features', name: 'docs-features', component: () => import('../views/DocsFeatures.vue'), meta: { title: 'Features' } },
        { path: 'docs/reference', name: 'docs-reference', component: () => import('../views/DocsReference.vue'), meta: { title: 'API Reference' } },
        { path: 'docs/markdown', name: 'docs-markdown', component: () => import('../views/MarkdownDocsPage.vue'), meta: { title: 'Markdown Docs' } },
        { path: 'docs/review', name: 'docs-review', component: () => import('../views/ReviewDashboardDocs.vue'), meta: { title: 'Review Dashboard' } },
        { path: 'legal', name: 'legal', component: () => import('../views/LegalPage.vue'), meta: { title: 'Legal' } },
        { path: 'setup', name: 'setup', component: () => import('../views/SetupPage.vue'), meta: { title: 'Setup' } },
        { path: 'operator/login', name: 'operator-login', component: () => import('../views/OperatorLogin.vue'), meta: { title: 'Operator Login' } },
      ],
    },
    { path: '/user-verification', name: 'user-verification', component: () => import('../views/UserVerificationPage.vue'), meta: { title: 'Verify Identity' } },
    { path: '/v/:slug', name: 'slug-verification', component: () => import('../views/UserVerificationPage.vue'), meta: { title: 'Verify Identity' } },
    { path: '/live-capture', name: 'live-capture', component: () => import('../views/LiveCapturePage.vue'), meta: { title: 'Live Capture' } },
    { path: '/verify/mobile', name: 'mobile-verification', component: () => import('../views/MobileVerificationPage.vue'), meta: { title: 'Mobile Verification' } },
    { path: '/admin/login', name: 'admin-login', component: () => import('../views/AdminLogin.vue'), meta: { title: 'Admin Login' } },
    {
      path: '/admin',
      component: () => import('../layouts/AdminLayout.vue'),
      meta: { requiresAdmin: true },
      children: [
        { path: '', redirect: '/admin/verifications' },
        { path: 'verifications', name: 'admin-verifications', component: () => import('../views/VerificationManagement.vue'), meta: { title: 'Verification Management', requiresAdmin: true } },
        { path: 'developers', name: 'admin-developers', component: () => import('../views/DevelopersList.vue'), meta: { title: 'Developers', requiresAdmin: true } },
        { path: 'thresholds', name: 'admin-thresholds', component: () => import('../views/AdminThresholds.vue'), meta: { title: 'Thresholds', requiresAdmin: true } },
      ],
    },
    { path: '/admin/*', redirect: '/admin/verifications' },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('../views/NotFoundPage.vue'), meta: { title: 'Not Found' } },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  if (to.meta.requiresAdmin && !localStorage.getItem('iv_admin_token')) {
    return { name: 'admin-login', query: { redirect: to.fullPath } }
  }
  document.title = `${to.meta.title || 'Identity Verification'} · Identity Verification`
})

export default router
