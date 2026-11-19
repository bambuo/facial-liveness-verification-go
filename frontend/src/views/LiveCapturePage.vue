<template>
  <div class="lc-page">
    <!-- Session expired -->
    <div v-if="sessionExpired" class="lc-center">
      <a-card class="lc-card lc-center-card">
        <icon-exclamation-circle class="lc-icon lc-icon-error" />
        <h2>Session expired</h2>
        <p class="lc-hint">Your live capture session has expired. Please start a new verification.</p>
        <a-button type="primary" long @click="startOver">Start new verification</a-button>
      </a-card>
    </div>

    <!-- Capture result + verification status -->
    <div v-else-if="captureResult" class="lc-center">
      <a-card class="lc-card lc-center-card">
        <a-spin v-if="isProcessing" :size="36" />
        <icon-check-circle v-else class="lc-icon" :class="isFailed ? 'lc-icon-error' : 'lc-icon-ok'" />
        <h2>{{ isProcessing ? 'Processing…' : 'Capture complete' }}</h2>
        <p class="lc-hint">{{ resultBody }}</p>

        <a-descriptions :data="resultRows" :column="1" bordered size="small" class="lc-gap" />

        <a-button
          v-if="!isProcessing"
          type="primary"
          long
          class="lc-gap"
          @click="goToResults"
        >
          View full results
        </a-button>
        <div v-else class="lc-processing">
          <a-spin :size="16" />
          <span class="lc-mono">{{ isPolling ? 'Checking verification status…' : 'Processing verification…' }}</span>
        </div>

        <a-alert v-if="error" type="error" class="lc-gap">
          <template #title>{{ error }}</template>
          <a-link @click="goToResults">Check results manually</a-link>
        </a-alert>
      </a-card>
    </div>

    <!-- Live capture -->
    <div v-else class="lc-main">
      <header class="lc-head">
        <div class="eyebrow">identity / live-capture</div>
        <h1>Live identity verification</h1>
        <p class="lc-hint">Complete your verification with live face capture.</p>
      </header>

      <a-alert v-if="sessionError" type="error" class="lc-gap">{{ sessionError }}</a-alert>

      <a-card v-if="!sessionError" class="lc-card" :body-style="{ padding: '0' }">
        <div class="lc-camera" :class="{ 'lc-camera-hidden': cameraState !== 'ready' }">
          <video ref="videoEl" playsinline muted autoplay />
          <canvas ref="overlayEl" width="640" height="480" />
        </div>

        <!-- Camera permission prompt -->
        <div v-if="cameraState === 'prompt'" class="lc-state">
          <icon-camera class="lc-icon lc-icon-accent" />
          <h2>Camera access required</h2>
          <p class="lc-hint">We need access to your camera for live identity verification.</p>
          <a-button type="primary" :loading="cameraBusy" @click="startCamera">Enable camera</a-button>
        </div>

        <!-- Initializing -->
        <div v-else-if="cameraState === 'initializing'" class="lc-state">
          <a-spin :size="32" />
          <p class="lc-hint">Starting camera…</p>
        </div>

        <!-- Camera error -->
        <div v-else-if="cameraState === 'error'" class="lc-state">
          <icon-exclamation-circle class="lc-icon lc-icon-error" />
          <h2>Camera access failed</h2>
          <p class="lc-hint">{{ cameraError }}</p>
          <a-button @click="retryCamera">Try again</a-button>
        </div>

        <!-- Live camera feed panel -->
        <div v-if="cameraState === 'ready'" class="lc-panel">
          <div class="lc-challenge">
            <div class="lc-challenge-icon"><icon-eye /></div>
            <div>
              <h3>Liveness challenge</h3>
              <p class="lc-hint">{{ challengeInstruction }}</p>
            </div>
          </div>

          <div class="lc-face" :class="faceDetected ? 'lc-face-ok' : 'lc-face-bad'">
            <a-badge :status="faceDetected ? 'success' : 'danger'" />
            <span class="lc-mono">
              {{ faceDetected ? 'FACE DETECTED — READY' : 'NO FACE DETECTED' }}
            </span>
            <span v-if="!faceDetected" class="lc-hint lc-mono">Position your face in frame</span>
            <span v-else class="lc-mono">
              Liveness: {{ Math.round(livenessScore * 100) }}% · Stability: {{ Math.round(faceStability * 100) }}%
            </span>
          </div>

          <div v-if="countdown !== null" class="lc-countdown">
            <strong>{{ countdown }}</strong>
            <span class="lc-mono">Get ready…</span>
          </div>

          <div class="lc-meta">
            <span class="lc-mono">Attempts: {{ captureAttempts }}/3</span>
            <span class="lc-mono">Expires: {{ expiresLabel }}</span>
          </div>

          <a-button
            v-if="countdown === null"
            type="primary"
            long
            :disabled="!faceDetected || livenessScore < 0.4 || faceStability < 0.5 || captureBusy"
            :loading="captureBusy"
            @click="startChallenge"
          >
            {{ captureBusy ? 'Processing…'
              : !faceDetected ? 'Position your face'
              : livenessScore < 0.4 ? 'Improve lighting'
              : faceStability < 0.5 ? 'Hold steady'
              : 'Start capture' }}
          </a-button>

          <a-alert v-if="error" type="error" class="lc-gap">{{ error }}</a-alert>
        </div>
      </a-card>

      <a-card class="lc-card lc-gap" title="Live capture instructions">
        <ul class="lc-tips">
          <li>Live face capture is processed by the verification engine</li>
          <li>Ensure good lighting on your face</li>
          <li>Position your face in the centre of the frame</li>
          <li>Wait for the green indicator showing face detection</li>
        </ul>
        <div class="lc-meta">
          <span class="lc-mono">Status: {{ cameraState }}</span>
          <a-button v-if="cameraState === 'ready'" size="small" @click="retryCamera">Restart camera</a-button>
        </div>
      </a-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const query = (key: string) => {
  const value = route.query[key]
  return typeof value === 'string' ? value : ''
}

const token = query('token')
const urlVerificationId = query('verification_id')
const urlApiKey = query('api_key')
const apiKey = urlApiKey || auth.apiKey

const verificationId = ref(urlVerificationId)
const tokenHeader = ref<'X-Session-Token' | 'X-Handoff-Token'>('X-Session-Token')
const sessionError = ref('')
const sessionExpired = ref(false)
const sessionDeadline = ref(Date.now() + 30 * 60 * 1000)

const cameraState = ref<'prompt' | 'initializing' | 'ready' | 'error'>('prompt')
const cameraBusy = ref(false)
const captureBusy = ref(false)
const cameraError = ref('')
const faceDetected = ref(false)
const livenessScore = ref(0)
const faceStability = ref(0)
const countdown = ref<number | null>(null)
const captureAttempts = ref(0)
const captureResult = ref<Record<string, any> | null>(null)
const verification = ref<Record<string, any> | null>(null)
const isPolling = ref(false)
const error = ref('')
const now = ref(Date.now())

const challengeInstruction = ref('Look directly at the camera and hold still while the frame is captured.')

const videoEl = ref<HTMLVideoElement | null>(null)
const overlayEl = ref<HTMLCanvasElement | null>(null)

let mediaStream: MediaStream | null = null
let rafId: number | null = null
let analysisCanvas: HTMLCanvasElement | null = null
let clockTimer: number | null = null
let expiryTimer: number | null = null
const faceHistory: boolean[] = []
const stabHistory: number[] = []

function headers(): Record<string, string> {
  if (token) return { [tokenHeader.value]: token }
  if (apiKey) return { 'X-API-Key': apiKey }
  return {}
}

const expiresLabel = computed(() => {
  const left = Math.max(0, Math.floor((sessionDeadline.value - now.value) / 1000))
  return `${Math.floor(left / 60)}:${String(left % 60).padStart(2, '0')}`
})

const finalStatus = computed(() => {
  const value = verification.value?.final_result ?? captureResult.value?.status ?? ''
  return typeof value === 'string' ? value : ''
})
const isProcessing = computed(() => !finalStatus.value || finalStatus.value === 'processing' || finalStatus.value === 'pending')
const isFailed = computed(() => finalStatus.value === 'failed')

const resultBody = computed(() => {
  if (isProcessing.value) return 'Please wait while we verify your identity.'
  if (finalStatus.value === 'verified') return 'Your identity has been successfully verified.'
  if (finalStatus.value === 'failed') return 'Verification failed. Please try again.'
  if (finalStatus.value === 'manual_review') return 'Your verification is under manual review.'
  return 'Your live capture has been successfully processed.'
})

const resultRows = computed(() => {
  const rows: { label: string; value: string }[] = []
  rows.push({ label: 'Status', value: finalStatus.value || 'processing' })
  if (captureResult.value) {
    rows.push({ label: 'Liveness check', value: 'Enabled' })
    rows.push({ label: 'Face matching', value: 'Enabled' })
  }
  const liveness = verification.value?.liveness_score ?? captureResult.value?.liveness_results?.liveness_score
  if (typeof liveness === 'number') rows.push({ label: 'Liveness score', value: `${Math.round(liveness * 100)}%` })
  const face = verification.value?.face_match_score ?? captureResult.value?.face_match_results?.confidence
  if (typeof face === 'number') rows.push({ label: 'Face match', value: `${Math.round(face * 100)}%` })
  const confidence = verification.value?.confidence_score
  if (typeof confidence === 'number') rows.push({ label: 'Confidence', value: `${Math.round(confidence * 100)}%` })
  return rows
})

async function resolveVerification() {
  if (verificationId.value) return
  if (!token) {
    sessionError.value = 'Invalid or missing live capture token.'
    return
  }
  try {
    const { data } = await api.get('/v2/verify/session-info', { headers: { 'X-Session-Token': token } })
    verificationId.value = data.verification_id || ''
    tokenHeader.value = 'X-Session-Token'
  } catch {
    try {
      const { data } = await api.get('/v2/verify/session-info', { headers: { 'X-Handoff-Token': token } })
      verificationId.value = data.verification_id || ''
      tokenHeader.value = 'X-Handoff-Token'
    } catch (err) {
      sessionError.value = apiMessage(err)
    }
  }
}

async function startCamera() {
  cameraState.value = 'initializing'
  cameraError.value = ''
  cameraBusy.value = true
  try {
    if (!navigator.mediaDevices?.getUserMedia) throw new Error('Camera not supported in this browser')
    mediaStream = await navigator.mediaDevices.getUserMedia({
      video: { width: { ideal: 640 }, height: { ideal: 480 }, facingMode: 'user' },
      audio: false,
    })
    const video = videoEl.value
    if (video) {
      video.srcObject = mediaStream
      await video.play()
    }
    cameraState.value = 'ready'
    rafId = window.requestAnimationFrame(cameraLoop)
  } catch (err) {
    cameraState.value = 'error'
    const domError = err as DOMException
    cameraError.value =
      domError.name === 'NotAllowedError' ? 'Camera permission denied. Please enable camera access.'
        : domError.name === 'NotFoundError' ? 'No camera found. Please connect a camera.'
          : domError.name === 'NotReadableError' ? 'Camera is already in use by another application.'
            : `Camera error: ${domError.message || 'Unknown error'}`
  } finally {
    cameraBusy.value = false
  }
}

function stopCamera() {
  if (rafId !== null) {
    window.cancelAnimationFrame(rafId)
    rafId = null
  }
  mediaStream?.getTracks().forEach((track) => track.stop())
  mediaStream = null
  const video = videoEl.value
  if (video) video.srcObject = null
  countdown.value = null
}

function retryCamera() {
  stopCamera()
  faceHistory.length = 0
  stabHistory.length = 0
  captureAttempts.value = 0
  error.value = ''
  cameraState.value = 'prompt'
  void startCamera()
}

function cameraLoop() {
  const video = videoEl.value
  const overlay = overlayEl.value
  if (cameraState.value !== 'ready') return
  if (!video || !overlay) {
    rafId = window.requestAnimationFrame(cameraLoop)
    return
  }
  if (video.readyState >= 2 && video.videoWidth > 0) {
    if (!analysisCanvas) analysisCanvas = document.createElement('canvas')
    analysisCanvas.width = 160
    analysisCanvas.height = 120
    const ctx = analysisCanvas.getContext('2d', { willReadFrequently: true })
    if (ctx) {
      ctx.drawImage(video, 0, 0, 160, 120)
      analyzeFrame(ctx.getImageData(0, 0, 160, 120))
      drawOverlay(overlay)
    }
  }
  rafId = window.requestAnimationFrame(cameraLoop)
}

function analyzeFrame(image: ImageData) {
  const { data, width, height } = image
  const cx = width / 2
  const cy = height / 2
  const size = Math.min(width, height) * 0.5
  let skin = 0
  let warm = 0
  let total = 0
  let brightness = 0
  let variance = 0
  for (let y = cy - size / 2; y < cy + size / 2; y++) {
    for (let x = cx - size / 2; x < cx + size / 2; x++) {
      if (x < 0 || x >= width || y < 0 || y >= height) continue
      const i = (Math.floor(y) * width + Math.floor(x)) * 4
      const r = data[i]
      const g = data[i + 1]
      const b = data[i + 2]
      const br = (r + g + b) / 3
      brightness += br
      variance += Math.abs(r - g) + Math.abs(g - b) + Math.abs(r - b)
      if ((r > 50 && g > 25 && b > 15 && r > b) || (br > 40 && br < 250 && r >= g) || (r > 30 && r + g > b * 1.5)) skin++
      if (r > g && r > b && br > 50) warm++
      total++
    }
  }
  const skinRatio = total ? skin / total : 0
  const warmRatio = total ? warm / total : 0
  const avgBrightness = total ? brightness / total : 0
  const avgVariance = total ? variance / total : 0
  livenessScore.value = Math.max(
    0.5,
    Math.min(1, skinRatio * 3) * 0.6
      + (avgBrightness > 30 && avgBrightness < 250 ? 1 : 0.7) * 0.2
      + Math.min(1, avgVariance / 20) * 0.2,
  )
  stabHistory.push(skinRatio > 0.02 || warmRatio > 0.05 ? 1 : 0)
  if (stabHistory.length > 5) stabHistory.shift()
  faceStability.value = Math.max(0.5, stabHistory.reduce((a, b) => a + b, 0) / Math.max(stabHistory.length, 1))

  const detected =
    (avgBrightness > 30 && avgBrightness < 250 && (skinRatio > 0.02 || warmRatio > 0.05))
    || (livenessScore.value > 0.3 && avgBrightness > 40 && avgVariance > 5)
  faceHistory.push(detected)
  if (faceHistory.length > 6) faceHistory.shift()
  const positive = faceHistory.filter(Boolean).length
  faceDetected.value = faceHistory.length >= 3 && positive >= Math.ceil(faceHistory.length * 0.5)
}

function drawOverlay(canvas: HTMLCanvasElement) {
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const { width, height } = canvas
  const cx = width / 2
  const cy = height / 2
  const radiusX = Math.min(width, height) * 0.22
  const radiusY = radiusX * 1.35
  ctx.clearRect(0, 0, width, height)
  ctx.strokeStyle = faceDetected.value ? '#10b981' : '#ef4444'
  ctx.lineWidth = 2
  ctx.setLineDash([8, 6])
  ctx.beginPath()
  ctx.ellipse(cx, cy, radiusX, radiusY, 0, 0, 2 * Math.PI)
  ctx.stroke()
  ctx.setLineDash([])
  ctx.fillStyle = ctx.strokeStyle
  ctx.font = '14px monospace'
  ctx.textAlign = 'center'
  ctx.fillText(faceDetected.value ? 'FACE DETECTED' : 'POSITION FACE IN FRAME', cx, cy + radiusY + 24)
  if (faceDetected.value) {
    ctx.font = '12px monospace'
    ctx.fillText(
      `LIVENESS ${Math.round(livenessScore.value * 100)}%  |  STABILITY ${Math.round(faceStability.value * 100)}%`,
      cx,
      cy + radiusY + 44,
    )
  }
}

function startChallenge() {
  if (countdown.value !== null) return
  if (!faceDetected.value || livenessScore.value < 0.4 || faceStability.value < 0.5) {
    error.value = !faceDetected.value
      ? 'No face detected. Please position your face clearly in the centre of the frame.'
      : livenessScore.value < 0.4
        ? 'Please ensure good lighting and your face is clearly visible.'
        : 'Please hold your face steady in the centre of the frame.'
    return
  }
  error.value = ''
  countdown.value = 3
  const timer = window.setInterval(() => {
    if (countdown.value === null) {
      window.clearInterval(timer)
      return
    }
    if (!faceDetected.value || livenessScore.value < 0.4 || faceStability.value < 0.5) {
      window.clearInterval(timer)
      countdown.value = null
      error.value = 'Face detection lost during countdown. Please try again.'
      return
    }
    if (countdown.value <= 1) {
      window.clearInterval(timer)
      countdown.value = null
      void performCapture()
    } else {
      countdown.value -= 1
    }
  }, 1000)
}

async function performCapture() {
  const video = videoEl.value
  if (!video || !verificationId.value) {
    error.value = 'Missing required data for capture.'
    return
  }
  captureBusy.value = true
  captureAttempts.value += 1
  try {
    const canvas = document.createElement('canvas')
    canvas.width = video.videoWidth || 640
    canvas.height = video.videoHeight || 480
    canvas.getContext('2d')?.drawImage(video, 0, 0, canvas.width, canvas.height)
    const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/jpeg', 0.8))
    if (!blob) throw new Error('Failed to capture image.')

    const form = new FormData()
    form.append('selfie', blob, 'selfie.jpg')
    const { data } = await api.post(`/v2/verify/${verificationId.value}/live-capture`, form, { headers: headers() })
    captureResult.value = data
    stopCamera()
    pollVerificationStatus()
  } catch (err) {
    error.value = apiMessage(err)
    if (captureAttempts.value >= 3) {
      error.value = 'Maximum capture attempts exceeded. Please refresh and try again.'
      stopCamera()
    }
  } finally {
    captureBusy.value = false
  }
}

async function pollVerificationStatus() {
  if (!verificationId.value || isPolling.value) return
  isPolling.value = true
  for (let attempt = 0; attempt < 60; attempt++) {
    try {
      const { data } = await api.get(`/v2/verify/${verificationId.value}/status`, { headers: headers() })
      verification.value = data
      if (data.final_result) {
        isPolling.value = false
        return
      }
    } catch {
      /* retry on the next tick */
    }
    await new Promise((resolve) => window.setTimeout(resolve, 5000))
  }
  isPolling.value = false
  error.value = 'Verification is taking longer than expected. Please check results manually.'
}

function goToResults() {
  if (token) {
    router.push(`/user-verification?session=${encodeURIComponent(token)}`)
    return
  }
  router.push('/user-verification')
}

function startOver() {
  router.push('/user-verification')
}

onMounted(async () => {
  await resolveVerification()
  clockTimer = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)
  expiryTimer = window.setTimeout(() => {
    sessionExpired.value = true
    stopCamera()
  }, Math.max(0, sessionDeadline.value - Date.now()))
})

onBeforeUnmount(() => {
  stopCamera()
  if (clockTimer !== null) window.clearInterval(clockTimer)
  if (expiryTimer !== null) window.clearTimeout(expiryTimer)
})
</script>

<style scoped>
.lc-page { min-height: 100vh; background: #f7f6f3; color: #0b0b0d; }
.lc-main { max-width: 720px; margin: 0 auto; padding: 32px 16px 56px; }
.lc-head { text-align: center; margin-bottom: 22px; }
.lc-head h1 { font-size: 24px; font-weight: 600; margin: 6px 0; font-family: ui-monospace, Menlo, monospace; }
.lc-hint { color: #6b6b70; font-size: 13px; line-height: 1.55; margin: 0; }
.lc-mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; color: #6b6b70; }
.lc-gap { margin-top: 14px; }
.lc-center { min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 24px 16px; }
.lc-card { border-radius: 4px; }
.lc-center-card { width: min(440px, 100%); text-align: center; }
.lc-center-card h2 { margin: 12px 0 6px; }
.lc-icon { font-size: 40px; }
.lc-icon-accent { color: #2563eb; }
.lc-icon-ok { color: #10b981; }
.lc-icon-error { color: #ef4444; }
.lc-camera { position: relative; background: #000; }
.lc-camera-hidden { display: none; }
.lc-camera video { width: 100%; max-height: 384px; object-fit: cover; display: block; }
.lc-camera canvas { position: absolute; inset: 0; width: 100%; height: 100%; }
.lc-state { display: flex; flex-direction: column; align-items: center; gap: 10px; padding: 32px; text-align: center; }
.lc-state h2 { margin: 0; }
.lc-panel { display: flex; flex-direction: column; gap: 14px; padding: 20px; border-top: 1px solid #e4e2dc; }
.lc-challenge { display: flex; align-items: center; gap: 12px; }
.lc-challenge-icon { width: 36px; height: 36px; border: 1px solid #e4e2dc; display: flex; align-items: center; justify-content: center; flex-shrink: 0; color: #2563eb; }
.lc-challenge h3 { margin: 0 0 2px; font-size: 14px; font-family: ui-monospace, Menlo, monospace; }
.lc-face { display: flex; flex-direction: column; align-items: center; gap: 6px; padding: 12px; border: 1px solid; }
.lc-face-ok { border-color: #10b981; background: #ecfdf5; }
.lc-face-bad { border-color: #ef4444; background: #fef2f2; }
.lc-countdown { text-align: center; font-family: ui-monospace, Menlo, monospace; }
.lc-countdown strong { display: block; font-size: 36px; line-height: 1; }
.lc-meta { display: flex; align-items: center; justify-content: space-between; gap: 10px; flex-wrap: wrap; }
.lc-processing { display: flex; align-items: center; justify-content: center; gap: 10px; }
.lc-tips { margin: 0; padding-left: 18px; color: #0b0b0d; font-size: 13px; line-height: 1.8; }
</style>
