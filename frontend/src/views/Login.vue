<template>
  <div class="login-wrapper min-h-screen flex flex-col items-center justify-center bg-slate-100 p-4 relative overflow-hidden">
    <!-- Sophisticated Abstract Background -->
    <div class="fixed inset-0 z-0 pointer-events-none">
        <div class="absolute -top-[10%] -right-[10%] w-[40%] h-[40%] rounded-full bg-blue-200/40 blur-[120px]"></div>
        <div class="absolute -bottom-[10%] -left-[10%] w-[40%] h-[40%] rounded-full bg-indigo-200/40 blur-[120px]"></div>
        <div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/cubes.png')] opacity-5"></div>
    </div>

    <!-- Login Container -->
    <div class="w-full max-w-[400px] relative z-10 animate-in">
      <!-- Logo Header -->
      <div class="flex flex-col items-center gap-4 mb-10">
        <div class="w-16 h-16 rounded-2xl bg-white text-blue-600 flex items-center justify-center shadow-xl border border-slate-200">
          <ShieldCheck :size="32" stroke-width="2.2" />
        </div>
        <div class="text-center">
          <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight">
            ACME <span class="text-blue-600 font-black">Node</span>
          </h1>
          <p class="text-sm text-slate-600 font-semibold tracking-wide mt-1 italic">Enterprise SSL Management</p>
        </div>
      </div>

      <!-- Modern Tech Card -->
      <div class="bg-white rounded-[2rem] shadow-[0_20px_50px_rgba(0,0,0,0.1)] border border-slate-200 p-8 sm:p-10 relative overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-blue-600 via-indigo-600 to-blue-600"></div>
        
        <form @submit.prevent="handleLogin" class="space-y-6">
          
          <div class="space-y-2">
            <label for="username" class="text-xs font-bold text-slate-600 uppercase tracking-widest ml-1">Username</label>
            <div class="relative flex items-center group">
              <User :size="18" class="absolute left-4 text-slate-500 group-focus-within:text-blue-600 transition-colors" />
              <input 
                v-model="username" 
                id="username" 
                type="text" 
                placeholder="Username"
                class="w-full h-13 bg-slate-50 border border-slate-300 text-slate-900 pl-12 pr-4 rounded-xl focus:bg-white focus:ring-4 focus:ring-blue-500/10 focus:border-blue-600 outline-none transition-all placeholder:text-slate-500 font-medium text-sm"
                required
              />
            </div>
          </div>

          <div class="space-y-2">
            <label for="password" class="text-xs font-bold text-slate-600 uppercase tracking-widest ml-1">Password</label>
            <div class="relative flex items-center group">
              <Lock :size="18" class="absolute left-4 text-slate-500 group-focus-within:text-blue-600 transition-colors" />
              <input 
                v-model="password" 
                id="password" 
                type="password" 
                placeholder="Password"
                class="w-full h-13 bg-slate-50 border border-slate-300 text-slate-900 pl-12 pr-4 rounded-xl focus:bg-white focus:ring-4 focus:ring-blue-500/10 focus:border-blue-600 outline-none transition-all placeholder:text-slate-500 font-medium text-sm"
                required
              />
            </div>
          </div>

          <div v-if="error" class="p-3 bg-red-50 border border-red-200 text-red-700 rounded-xl text-xs flex items-center gap-3 font-semibold">
            <AlertCircle :size="16" class="shrink-0" />
            <span>{{ error }}</span>
          </div>

          <button 
            type="submit" 
            :disabled="loading"
            class="w-full h-13 bg-blue-700 hover:bg-blue-800 active:scale-[0.98] text-white font-bold text-sm uppercase tracking-widest rounded-xl transition-all shadow-[0_8px_20px_rgba(29,78,216,0.3)] hover:shadow-[0_12px_24px_rgba(29,78,216,0.4)] flex items-center justify-center gap-3 disabled:opacity-50"
          >
            <Loader2 v-if="loading" class="animate-spin" :size="18" />
            <span v-else>Login</span>
          </button>
        </form>
      </div>

      <!-- Compact Footer -->
      <div class="mt-8 text-center">
        <p class="text-[11px] font-bold text-slate-500 uppercase tracking-[0.2em] opacity-90">
          Node_Security_Protocol // v1.0.4
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
