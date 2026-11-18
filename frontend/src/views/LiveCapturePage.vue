<template>
  <div class="capture-shell">
    <header class="capture-header"><router-link to="/" class="brand">Identity Verification</router-link><a-tag color="arcoblue">Live capture</a-tag></header>
    <main class="capture-main">
      <div class="capture-card">
        <div class="eyebrow">Biometric capture</div><h1>Live selfie</h1>
        <p class="muted">Center your face, look at the camera and keep your eyes open.</p>
        <div class="camera-frame">
          <video ref="video" autoplay playsinline muted></video>
          <canvas ref="canvas" style="display:none"></canvas>
          <div v-if="!streaming" class="camera-placeholder"><icon-camera style="font-size:36px" /><span>Camera preview</span></div>
        </div>
        <a-space style="margin-top:16px" wrap>
          <a-button v-if="!streaming" type="primary" @click="startCamera">Start camera</a-button>
          <a-button v-if="streaming" type="primary" @click="capture" :loading="uploading">Capture & analyze</a-button>
          <a-button v-if="streaming" @click="stopCamera">Stop</a-button>
          <label class="upload-label">Upload instead<input type="file" accept="image/jpeg,image/png" @change="uploadFile" /></label>
        </a-space>
        <a-alert v-if="message" style="margin-top:16px" :type="error ? 'error' : 'success'">{{ message }}</a-alert>
        <div v-if="result" class="result-grid">
          <div class="panel"><div class="stat-label">Face confidence</div><div class="stat-value">{{ format(result.face_confidence) }}</div></div>
          <div class="panel"><div class="stat-label">Liveness</div><div class="stat-value">{{ result.liveness_passed ? 'Passed' : 'Failed' }}</div></div>
          <div class="panel"><div class="stat-label">Liveness score</div><div class="stat-value">{{ format(result.liveness_score) }}</div></div>
          <div class="panel"><div class="stat-label">Deepfake</div><div class="stat-value">{{ result.deepfake_check ? (result.deepfake_check.isReal ? 'Real' : 'Fake') : '—' }}</div></div>
        </div>
      </div>
      <aside class="capture-aside"><h3>Tips for a good capture</h3><ul><li>Use even, natural light.</li><li>Remove sunglasses and hats.</li><li>Keep your whole face in frame.</li><li>Do not use a photo of a photo.</li></ul><a-alert type="info">Camera access requires HTTPS or localhost.</a-alert></aside>
    </main>
  </div>
</template>
<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
const video = ref<HTMLVideoElement | null>(null); const canvas = ref<HTMLCanvasElement | null>(null)
const streaming = ref(false); const uploading = ref(false); const message = ref(''); const error = ref(false); const result = ref<any>(null)
let mediaStream: MediaStream | null = null
function format(value: any) { return typeof value === 'number' ? value.toFixed(3) : '—' }
async function startCamera() {
  try {
    mediaStream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'user', width: { ideal: 1280 }, height: { ideal: 960 } }, audio: false })
    if (video.value) { video.value.srcObject = mediaStream; await video.value.play() }
    streaming.value = true; message.value = ''; error.value = false
  } catch (e) { error.value = true; message.value = 'Camera unavailable: ' + (e as Error).message }
}
function stopCamera() { mediaStream?.getTracks().forEach((track) => track.stop()); mediaStream = null; streaming.value = false }
function capture() {
  if (!video.value || !canvas.value) return
  const width = video.value.videoWidth || 640; const height = video.value.videoHeight || 480
  canvas.value.width = width; canvas.value.height = height
  canvas.value.getContext('2d')?.drawImage(video.value, 0, 0, width, height)
  canvas.value.toBlob((blob) => { if (blob) analyze(new File([blob], 'selfie.jpg', { type: 'image/jpeg' })) }, 'image/jpeg', 0.9)
}
function uploadFile(event: Event) { const file = (event.target as HTMLInputElement).files?.[0]; if (file) analyze(file) }
async function analyze(file: File) {
  uploading.value = true; error.value = false
  try {
    const form = new FormData(); form.append('file', file)
    const { data } = await api.post('/extract/live', form)
    result.value = data.result; message.value = 'Capture analyzed'
  } catch (e) { error.value = true; message.value = apiMessage(e) } finally { uploading.value = false }
}
onBeforeUnmount(stopCamera)
</script>
<style scoped>
.capture-shell { min-height: 100vh; background: #0f172a; color: #e2e8f0; }
.capture-header { height: 64px; padding: 0 24px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid #1e293b; }
.brand { color: #fff; text-decoration: none; font-weight: 700; }
.capture-main { max-width: 1080px; margin: 0 auto; padding: 32px 20px 60px; display: grid; grid-template-columns: minmax(0,1fr) 280px; gap: 22px; }
.capture-card { background: #111827; border: 1px solid #1f2937; border-radius: 18px; padding: 28px; }
.capture-aside { background: #111827; border: 1px solid #1f2937; border-radius: 14px; padding: 20px; }
.capture-aside li { margin-bottom: 8px; color: #94a3b8; }
.camera-frame { position: relative; aspect-ratio: 4/3; border-radius: 14px; overflow: hidden; background: #020617; margin-top: 18px; }
.camera-frame video { width: 100%; height: 100%; object-fit: cover; }
.camera-placeholder { position: absolute; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; color: #475569; }
.upload-label { display: inline-flex; align-items: center; padding: 0 15px; height: 32px; border: 1px solid #334155; border-radius: 4px; cursor: pointer; font-size: 14px; }
.upload-label input { display: none; }
.result-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-top: 18px; }
.result-grid .panel { background: #0b1220; border-color: #1f2937; }
.result-grid .stat-label { color: #94a3b8; }
@media (max-width: 900px) { .capture-main { grid-template-columns: 1fr; } .result-grid { grid-template-columns: repeat(2,1fr); } }
</style>
