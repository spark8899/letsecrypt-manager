<template>
  <div class="login-wrapper min-h-screen flex flex-col items-center justify-center bg-slate-50/80 p-4">
    <!-- Subtle Background Decoration -->
    <div class="fixed inset-0 z-0 pointer-events-none opacity-30">
        <div class="absolute top-[20%] left-[30%] w-48 h-48 rounded-full bg-blue-200 blur-[80px]"></div>
        <div class="absolute bottom-[20%] right-[30%] w-48 h-48 rounded-full bg-indigo-200 blur-[80px]"></div>
    </div>

    <!-- Ultra-Compact Container -->
    <div class="w-[280px] relative z-10 animate-in">
      <!-- Minimal Header -->
      <div class="flex items-center justify-center gap-2.5 mb-6">
        <div class="w-8 h-8 rounded-lg bg-blue-600 text-white flex items-center justify-center shadow-sm">
          <ShieldCheck :size="18" />
        </div>
        <h1 class="text-lg font-extrabold text-slate-800 tracking-tight">ACME MGMT</h1>
      </div>

      <!-- Mini Card -->
      <div class="bg-white rounded-2xl shadow-xl shadow-slate-200/40 border border-slate-100 p-5">
        <form @submit.prevent="handleLogin" class="space-y-4">
          
          <!-- Identity -->
          <div class="space-y-1">
            <label for="username" class="text-[10px] font-bold text-slate-400 uppercase tracking-widest ml-0.5">Identity</label>
            <div class="relative flex items-center group">
              <User :size="14" class="absolute left-3 text-slate-300 group-focus-within:text-blue-500 transition-colors" />
              <input 
                v-model="username" 
                id="username" 
                type="text" 
                placeholder="User"
                class="w-full h-9 bg-slate-50 border border-slate-200 text-slate-900 pl-9 pr-3 rounded-lg focus:bg-white focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 outline-none transition-all placeholder:text-slate-300 text-xs"
                required
              />
            </div>
          </div>

          <!-- Credentials -->
          <div class="space-y-1">
            <label for="password" class="text-[10px] font-bold text-slate-400 uppercase tracking-widest ml-0.5">Password</label>
            <div class="relative flex items-center group">
              <Lock :size="14" class="absolute left-3 text-slate-300 group-focus-within:text-blue-500 transition-colors" />
              <input 
                v-model="password" 
                id="password" 
                type="password" 
                placeholder="••••"
                class="w-full h-9 bg-slate-50 border border-slate-200 text-slate-900 pl-9 pr-3 rounded-lg focus:bg-white focus:ring-2 focus:ring-blue-500/10 focus:border-blue-500 outline-none transition-all placeholder:text-slate-300 text-xs"
                required
              />
            </div>
          </div>

          <!-- Error Alert (Mini) -->
          <div v-if="error" class="p-2 bg-red-50 border border-red-100 text-red-500 rounded-lg text-[10px] flex items-center gap-2 leading-tight">
            <AlertCircle :size="12" class="shrink-0" />
            <span class="font-medium truncate" :title="error">{{ error }}</span>
          </div>

          <!-- Action -->
          <button 
            type="submit" 
            :disabled="loading"
            class="w-full h-9 bg-blue-600 hover:bg-blue-700 active:scale-[0.97] text-white font-bold text-xs rounded-lg transition-all shadow-sm shadow-blue-100 flex items-center justify-center gap-2 mt-1 disabled:opacity-50"
          >
            <Loader2 v-if="loading" class="animate-spin" :size="14" />
            <span v-else>Authorize</span>
          </button>
        </form>
      </div>

      <!-- Micro Footer -->
      <div class="mt-6 text-center opacity-40">
        <p class="text-[9px] font-mono text-slate-500 tracking-tighter uppercase">
          SECURE_GATEWAY_V1
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'
import { ShieldCheck, User, Lock, AlertCircle, Loader2 } from 'lucide-vue-next'

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const router = useRouter()
const authStore = useAuthStore()

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  try {
    await authStore.login(username.value, password.value)
    router.push({ name: 'Dashboard' })
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Denied'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrapper {
  font-family: ui-sans-serif, system-ui, sans-serif;
}

.animate-in {
  animation: enter 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes enter {
  from { opacity: 0; transform: scale(0.98) translateY(2px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
</style>
