<template>
  <div class="min-h-screen bg-slate-50 text-slate-900 flex flex-col font-sans selection:bg-blue-100 selection:text-blue-700">
    <!-- Topbar: Modern White Glass -->
    <header class="sticky top-0 z-50 bg-white/80 backdrop-blur-md border-b border-slate-200/60">
      <div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="p-2 bg-blue-600 rounded-xl text-white shadow-lg shadow-blue-200">
            <ShieldCheck :size="24" stroke-width="2.2" />
          </div>
          <div>
            <h1 class="text-xl font-bold text-slate-900 tracking-tight flex items-center gap-2">
              ACME <span class="text-blue-600">Node</span>
              <span class="text-[10px] bg-slate-100 text-slate-500 px-2 py-0.5 rounded-full border border-slate-200 font-bold tracking-widest uppercase">Console</span>
            </h1>
            <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest">Autonomous SSL Gateway</p>
          </div>
        </div>
        
        <div class="flex items-center gap-6">
          <button 
            @click="showAddDomainModal = true" 
            class="bg-blue-600 hover:bg-blue-700 text-white px-5 py-2.5 rounded-xl flex items-center gap-2.5 transition-all font-bold text-xs uppercase tracking-widest shadow-lg shadow-blue-100 active:scale-95"
          >
            <Plus :size="16" stroke-width="3" />
            <span>Deploy Domain</span>
          </button>
          <div class="h-6 w-[1px] bg-slate-200"></div>
          <button 
            @click="handleLogout" 
            class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-xl transition-all group"
            title="Terminate Session"
          >
            <LogOut :size="20" class="group-hover:translate-x-1 transition-transform" />
          </button>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 p-8 max-w-7xl mx-auto w-full space-y-8">
      <!-- Stats: Clean White Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="bg-white border border-slate-200/60 p-6 rounded-[2rem] shadow-sm hover:shadow-md transition-shadow group">
            <div class="flex items-center gap-5">
                <div class="p-4 bg-blue-50 text-blue-600 rounded-2xl group-hover:scale-110 transition-transform"><Globe :size="28" /></div>
                <div><p class="text-[10px] text-slate-400 font-black uppercase tracking-[0.2em] mb-1">Managed Nodes</p><p class="text-3xl font-black text-slate-900">{{ domains.length }}</p></div>
            </div>
        </div>
        <div class="bg-white border border-slate-200/60 p-6 rounded-[2rem] shadow-sm hover:shadow-md transition-shadow group">
            <div class="flex items-center gap-5">
                <div class="p-4 bg-emerald-50 text-emerald-600 rounded-2xl group-hover:scale-110 transition-transform"><CheckCircle :size="28" /></div>
                <div><p class="text-[10px] text-slate-400 font-black uppercase tracking-[0.2em] mb-1">Active Certs</p><p class="text-3xl font-black text-slate-900">{{ issuedCount }}</p></div>
            </div>
        </div>
        <div class="bg-white border border-slate-200/60 p-6 rounded-[2rem] shadow-sm hover:shadow-md transition-shadow group">
            <div class="flex items-center gap-5">
                <div class="p-4 bg-amber-50 text-amber-600 rounded-2xl group-hover:scale-110 transition-transform"><Clock :size="28" /></div>
                <div><p class="text-[10px] text-slate-400 font-black uppercase tracking-[0.2em] mb-1">Expiring Soon</p><p class="text-3xl font-black text-slate-900">{{ expiringSoonCount }}</p></div>
            </div>
        </div>
      </div>

      <!-- Main Table Card -->
      <div class="bg-white border border-slate-200/60 rounded-[2rem] overflow-hidden shadow-sm">
        <div class="px-8 py-6 border-b border-slate-100 flex items-center justify-between bg-slate-50/30">
          <h2 class="font-bold text-slate-800 uppercase tracking-widest flex items-center gap-3 text-sm">
            <LayoutDashboard :size="18" class="text-blue-600" />
            Operational Nodes
          </h2>
          <button @click="fetchDomains" class="p-2 text-slate-400 hover:text-blue-600 transition-colors" title="Sync Data">
            <RefreshCw :size="16" :class="{'animate-spin': loading}" />
          </button>
        </div>
        
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="text-[10px] uppercase font-black text-slate-400 tracking-[0.2em] bg-slate-50/50">
                <th class="px-8 py-4 border-b border-slate-100">Network Domain</th>
                <th class="px-8 py-4 border-b border-slate-100">Status Core</th>
                <th class="px-8 py-4 border-b border-slate-100">TTL / Expiry</th>
                <th class="px-8 py-4 border-b border-slate-100 text-right">Operations</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50">
              <tr v-for="d in domains" :key="d.domain" class="hover:bg-blue-50/50 transition-all group">
                <td class="px-8 py-5">
                  <div class="flex flex-col">
                    <span class="font-bold text-slate-800 tracking-tight text-base">{{ d.domain }}</span>
                    <span class="text-[10px] font-mono text-slate-400 uppercase">SAN: *.{{ d.domain }}</span>
                  </div>
                </td>
                <td class="px-8 py-5">
                  <div class="flex flex-col gap-1">
                    <span :class="getStatusClassLight(d.status)" class="px-3 py-1 rounded-lg text-[10px] font-bold inline-flex items-center gap-2 uppercase tracking-tighter w-fit border">
                      <span class="w-1.5 h-1.5 rounded-full" :class="getStatusDotClassLight(d.status)"></span>
                      {{ d.status }}
                    </span>
                    <p v-if="d.error" class="text-[10px] text-red-500 font-medium mt-1 truncate max-w-xs opacity-80" :title="d.error">ERR: {{ d.error }}</p>
                  </div>
                </td>
                <td class="px-8 py-5">
                  <div v-if="d.cert_expiry" class="flex flex-col">
                    <span class="font-bold text-slate-700 text-sm">{{ formatDate(d.cert_expiry) }}</span>
                    <span class="text-[10px] font-bold uppercase tracking-widest" :class="getExpiryClassLight(d.cert_expiry)">
                      {{ getExpiryDays(d.cert_expiry) }} Days Remaining
                    </span>
                  </div>
                  <span v-else class="text-slate-300 font-mono text-xs italic tracking-widest">--- NO_DATA ---</span>
                </td>
                <td class="px-8 py-5">
                  <div class="flex items-center justify-end gap-3">
                    <button 
                      v-if="d.status === 'issued'"
                      @click="viewCert(d)" 
                      class="p-2.5 text-blue-600 bg-blue-50 hover:bg-blue-100 border border-blue-100 rounded-xl transition-all"
                      title="Dump Certificate"
                    >
                      <FileText :size="18" />
                    </button>
                    <button 
                      @click="manageDomain(d)" 
                      class="px-4 py-2 bg-slate-100 hover:bg-blue-600 text-slate-600 hover:text-white rounded-xl text-[10px] font-bold transition-all flex items-center gap-2 uppercase tracking-widest border border-slate-200 hover:border-blue-600"
                    >
                      <Settings2 :size="14" />
                      Protocol
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="domains.length === 0 && !loading">
                <td colspan="4" class="px-8 py-24 text-center">
                  <div class="flex flex-col items-center gap-4 opacity-20">
                      <Inbox :size="64" stroke-width="1.2" />
                      <p class="text-xl font-bold uppercase tracking-[0.3em] text-slate-400">No Active Nodes</p>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </main>

    <!-- Modal: Add Domain -->
    <Teleport to="body">
        <div v-if="showAddDomainModal" class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
            <div class="bg-white border border-slate-200 rounded-[2.5rem] w-full max-w-md overflow-hidden shadow-2xl animate-in">
                <div class="bg-blue-600 p-8 text-white">
                    <h3 class="text-xl font-bold flex items-center gap-3 uppercase tracking-tight">
                        <PlusCircle :size="24" /> Node Deployment
                    </h3>
                </div>
                <div class="p-10 space-y-8">
                    <div class="space-y-3">
                        <label class="text-[10px] font-bold text-slate-400 uppercase tracking-[0.2em] ml-1">Target FQDN</label>
                        <input 
                            v-model="newDomain" 
                            type="text" 
                            placeholder="internal.network.com"
                            class="w-full h-14 bg-slate-50 border border-slate-200 rounded-2xl px-5 text-slate-900 focus:bg-white focus:ring-4 focus:ring-blue-500/10 focus:border-blue-500 outline-none transition-all font-medium text-sm placeholder:text-slate-300"
                            @keyup.enter="handleAddDomain"
                        />
                        <p class="text-[10px] text-slate-400 italic mt-2">Will auto-provision Wildcard and Root certificates.</p>
                    </div>
                    <div class="flex gap-3">
                        <button @click="showAddDomainModal = false" class="flex-1 py-4 bg-slate-100 hover:bg-slate-200 text-slate-600 font-bold uppercase text-[10px] tracking-widest rounded-2xl transition-all">Abort</button>
                        <button 
                            @click="handleAddDomain" 
                            :disabled="!newDomain || loading"
                            class="flex-[2] py-4 bg-blue-600 hover:bg-blue-700 text-white font-bold uppercase text-[10px] tracking-widest rounded-2xl transition-all shadow-lg shadow-blue-100 flex items-center justify-center gap-2"
                        >
                            <Loader2 v-if="loading" class="animate-spin" :size="16" />
                            Initiate Deploy
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </Teleport>

    <!-- Modal: Manage Domain -->
    <div v-if="selectedDomain" class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
      <div class="bg-white rounded-[2.5rem] w-full max-w-2xl overflow-hidden shadow-2xl border border-slate-200 flex flex-col max-h-[90vh] animate-in">
        <div class="bg-slate-50 p-8 border-b border-slate-200 flex items-center justify-between">
            <div>
                <h3 class="text-2xl font-bold text-slate-900 tracking-tight uppercase">{{ selectedDomain.domain }}</h3>
                <div class="flex items-center gap-3 mt-2">
                    <span class="text-[10px] font-bold uppercase text-blue-600 tracking-widest bg-blue-50 px-2.5 py-1 rounded-lg border border-blue-100">{{ selectedDomain.status }}</span>
                    <span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">System_Auth_Required</span>
                </div>
            </div>
            <button @click="selectedDomain = null" class="p-3 bg-white hover:bg-red-50 hover:text-red-500 text-slate-400 rounded-2xl transition-all border border-slate-100"><X :size="20" /></button>
        </div>
        
        <div class="flex-1 overflow-y-auto p-10 space-y-12">
            <div class="relative pl-14">
                <div class="absolute left-0 top-0 w-10 h-10 rounded-2xl flex items-center justify-center font-bold text-sm border-2" :class="stepClassLight(1, selectedDomain.status)">01</div>
                <div>
                    <h4 class="font-bold text-slate-800 uppercase tracking-widest text-xs mb-2">Phase Alpha: DNS Challenge</h4>
                    <p class="text-[11px] text-slate-500 leading-relaxed mb-6 font-medium">Request ACME DNS-01 authorization tokens for ownership verification.</p>
                    <button 
                        v-if="['pending', 'failed', 'issued'].includes(selectedDomain.status)"
                        @click="handleGetChallenge" 
                        class="px-6 py-3 bg-blue-600 text-white text-[10px] font-bold rounded-xl hover:bg-blue-700 transition-all flex items-center gap-3 uppercase tracking-widest shadow-md shadow-blue-100"
                        :disabled="loading"
                    >
                        <Zap :size="16" /> Execute Token Request
                    </button>
                    
                    <div v-if="selectedDomain.challenge" class="mt-6 p-6 bg-slate-50 border border-slate-200 rounded-2xl space-y-6">
                        <div class="grid grid-cols-1 gap-6">
                            <div>
                                <p class="text-[9px] font-bold text-slate-400 uppercase tracking-[0.2em] mb-2">Registry Type</p>
                                <span class="font-mono bg-white text-blue-600 px-3 py-1 rounded-lg border border-slate-200 text-xs font-bold shadow-sm">TXT_RECORD</span>
                            </div>
                            <div>
                                <p class="text-[9px] font-bold text-slate-400 uppercase tracking-[0.2em] mb-2">Canonical Host</p>
                                <div class="flex items-center gap-3 bg-white p-3 rounded-xl border border-slate-200 shadow-sm">
                                    <code class="text-slate-600 flex-1 truncate font-mono text-xs">{{ selectedDomain.challenge.record_name }}</code>
                                    <button @click="copy(selectedDomain.challenge.record_name)" class="p-2 hover:bg-slate-50 rounded-lg text-slate-400 transition-colors"><Copy :size="14" /></button>
                                </div>
                            </div>
                            <div>
                                <p class="text-[9px] font-bold text-slate-400 uppercase tracking-[0.2em] mb-2">Cryptographic Key</p>
                                <div class="flex items-center gap-3 bg-white p-3 rounded-xl border border-slate-200 shadow-sm">
                                    <code class="text-amber-600 flex-1 truncate font-mono text-xs">{{ selectedDomain.challenge.txt_value }}</code>
                                    <button @click="copy(selectedDomain.challenge.txt_value)" class="p-2 hover:bg-slate-50 rounded-lg text-slate-400 transition-colors"><Copy :size="14" /></button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="relative pl-14" :class="{'opacity-30 grayscale pointer-events-none': !selectedDomain.challenge}">
                <div class="absolute left-0 top-0 w-10 h-10 rounded-2xl flex items-center justify-center font-bold text-sm border-2" :class="stepClassLight(2, selectedDomain.status)">02</div>
                <div>
                    <h4 class="font-bold text-slate-800 uppercase tracking-widest text-xs mb-2">Phase Beta: Propagation Check</h4>
                    <p class="text-[11px] text-slate-500 leading-relaxed mb-6 font-medium">Verify global recursive DNS synchronization for the provided challenge tokens.</p>
                    <div class="flex items-center gap-4">
                        <button 
                            @click="handleVerifyDNS" 
                            class="px-6 py-3 bg-slate-900 text-white text-[10px] font-bold rounded-xl hover:bg-black transition-all flex items-center gap-3 uppercase tracking-widest"
                            :disabled="loading"
                        >
                            <Search :size="16" /> Scan Global DNS
                        </button>
                        <span v-if="verifyStatus" class="text-[10px] font-bold flex items-center gap-2 uppercase tracking-widest" :class="verifyStatus.success ? 'text-emerald-600' : 'text-amber-600'">
                            <span class="w-2 h-2 rounded-full" :class="verifyStatus.success ? 'bg-emerald-500 shadow-sm' : 'bg-amber-500'"></span>
                            {{ verifyStatus.message }}
                        </span>
                    </div>
                </div>
            </div>

            <div class="relative pl-14" :class="{'opacity-30 grayscale pointer-events-none': selectedDomain.status !== 'verified'}">
                <div class="absolute left-0 top-0 w-10 h-10 rounded-2xl flex items-center justify-center font-bold text-sm border-2" :class="stepClassLight(3, selectedDomain.status)">03</div>
                <div>
                    <h4 class="font-bold text-slate-800 uppercase tracking-widest text-xs mb-2">Phase Gamma: Provisioning</h4>
                    <p class="text-[11px] text-slate-500 leading-relaxed mb-6 font-medium">Finalize ACME order and generate high-entropy RSA-4096 / ECC certificate pair.</p>
                    <button 
                        @click="handleIssueCert" 
                        class="px-8 py-4 bg-emerald-600 text-white text-[11px] font-bold rounded-2xl hover:bg-emerald-700 transition-all flex items-center gap-3 uppercase tracking-[0.2em] shadow-lg shadow-emerald-100"
                        :disabled="loading || selectedDomain.status !== 'verified'"
                    >
                        <Award :size="18" /> Commit Certificate
                    </button>
                </div>
            </div>
        </div>
      </div>
    </div>

    <!-- Modal: View Cert -->
    <div v-if="certData" class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[150] flex items-center justify-center p-4">
      <div class="bg-white rounded-[3rem] w-full max-w-5xl overflow-hidden shadow-2xl border border-slate-200 flex flex-col max-h-[90vh] animate-in">
        <div class="bg-blue-600 p-10 text-white flex items-center justify-between">
            <div>
                <h3 class="text-2xl font-bold flex items-center gap-4 uppercase tracking-tight">
                    <Award :size="28" stroke-width="2.5" /> {{ certData.domain }}
                </h3>
                <p class="text-[10px] font-bold uppercase tracking-[0.2em] mt-2 opacity-80">VALID_UNTIL: {{ formatDate(certData.cert_expiry) }}</p>
            </div>
            <button @click="certData = null" class="p-3 bg-white/10 hover:bg-white/20 rounded-2xl transition-all"><X :size="24" /></button>
        </div>
        
        <div class="flex-1 overflow-y-auto p-10 bg-slate-50 text-slate-600 font-mono text-[11px]">
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-10">
                <div class="space-y-4">
                    <div class="flex items-center justify-between bg-white p-4 rounded-2xl border border-slate-200 shadow-sm">
                        <span class="text-slate-400 font-bold uppercase tracking-widest text-[9px]">FULLCHAIN_PUB_KEY</span>
                        <button @click="copy(certData.fullchain_cer)" class="text-blue-600 hover:text-blue-700 transition-colors flex items-center gap-2 font-bold uppercase text-[9px] tracking-widest">
                            <Copy :size="14" /> Copy_Buffer
                        </button>
                    </div>
                    <pre class="bg-white p-6 rounded-3xl border border-slate-200 overflow-x-auto text-slate-500 leading-relaxed shadow-sm">{{ certData.fullchain_cer }}</pre>
                </div>
                <div class="space-y-4">
                    <div class="flex items-center justify-between bg-white p-4 rounded-2xl border border-slate-200 shadow-sm">
                        <span class="text-amber-600 font-bold uppercase tracking-widest text-[9px]">PRIVATE_RSA_KEY</span>
                        <button @click="copy(certData.private_key)" class="text-amber-600 hover:text-amber-700 transition-colors flex items-center gap-2 font-bold uppercase text-[9px] tracking-widest">
                            <Copy :size="14" /> Copy_Buffer
                        </button>
                    </div>
                    <pre class="bg-white p-6 rounded-3xl border border-slate-200 overflow-x-auto text-amber-600/80 leading-relaxed shadow-sm">{{ certData.private_key }}</pre>
                </div>
            </div>
        </div>
      </div>
    </div>

    <!-- Toast Notification -->
    <div v-if="toast" class="fixed bottom-10 right-10 bg-white text-slate-900 px-8 py-4 rounded-2xl shadow-2xl z-[200] flex items-center gap-4 animate-in border border-slate-100">
        <div class="w-8 h-8 bg-blue-50 text-blue-600 rounded-full flex items-center justify-center">
            <CheckCircle :size="18" stroke-width="3" />
        </div>
        <span class="font-bold uppercase tracking-widest text-[10px]">{{ toast }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'
import { useDomainStore, DomainInfo } from '../store/domain'
import { 
  ShieldCheck, LogOut, Plus, Globe, CheckCircle, Clock, 
  RefreshCw, FileText, Settings2, PlusCircle, X, Zap, 
  Copy, Search, Award, LayoutDashboard, Inbox, Loader2
} from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const domainStore = useDomainStore()

const loading = ref(false)
const showAddDomainModal = ref(false)
const newDomain = ref('')
const selectedDomain = ref<DomainInfo | null>(null)
const certData = ref<any>(null)
const verifyStatus = ref<{success: boolean, message: string} | null>(null)
const toast = ref('')

const domains = computed(() => domainStore.domains)
const issuedCount = computed(() => domains.value.filter(d => d.status === 'issued').length)
const expiringSoonCount = computed(() => {
    const now = new Date()
    const thirtyDaysLater = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000)
    return domains.value.filter(d => {
        if (!d.cert_expiry) return false
        const expiry = new Date(d.cert_expiry)
        return expiry < thirtyDaysLater && expiry > now
    }).length
})

const fetchDomains = async () => {
  loading.value = true
  try {
    await domainStore.fetchDomains()
  } finally {
    loading.value = false
  }
}

const handleLogout = () => {
  authStore.logout()
  router.push({ name: 'Login' })
}

const handleAddDomain = async () => {
    if (!newDomain.value) return
    loading.value = true
    try {
        await domainStore.addDomain(newDomain.value)
        newDomain.value = ''
        showAddDomainModal.value = false
        showToast('Node deployed')
    } catch (err: any) {
        alert(err.response?.data?.error || 'Deployment failed')
    } finally {
        loading.value = false
    }
}

const manageDomain = (domain: DomainInfo) => {
    selectedDomain.value = domain
    verifyStatus.value = null
}

const handleGetChallenge = async () => {
    if (!selectedDomain.value) return
    loading.value = true
    try {
        await domainStore.getChallenge(selectedDomain.value.domain)
        selectedDomain.value = domains.value.find(d => d.domain === selectedDomain.value?.domain) || null
        showToast('Tokens requested')
    } catch (err: any) {
        alert(err.response?.data?.error || 'Request failed')
    } finally {
        loading.value = false
    }
}

const handleVerifyDNS = async () => {
    if (!selectedDomain.value) return
    loading.value = true
    try {
        const verified = await domainStore.verifyDNS(selectedDomain.value.domain)
        verifyStatus.value = verified 
            ? { success: true, message: 'Tokens verified' }
            : { success: false, message: 'Sync pending' }
        selectedDomain.value = domains.value.find(d => d.domain === selectedDomain.value?.domain) || null
    } catch (err: any) {
        verifyStatus.value = { success: false, message: 'Scan error' }
    } finally {
        loading.value = false
    }
}

const handleIssueCert = async () => {
    if (!selectedDomain.value) return
    loading.value = true
    try {
        await domainStore.issueCert(selectedDomain.value.domain)
        showToast('Cert committed')
        selectedDomain.value = null
    } catch (err: any) {
        alert(err.response?.data?.error || 'Commit failed')
    } finally {
        loading.value = false
    }
}

const viewCert = async (domain: DomainInfo) => {
    try {
        const data = await domainStore.getCert(domain.domain)
        certData.value = data
    } catch (err) {
        alert('Data fetch failed')
    }
}

const copy = (text: string) => {
    navigator.clipboard.writeText(text)
    showToast('Buffered to clipboard')
}

const showToast = (msg: string) => {
    toast.value = msg
    setTimeout(() => toast.value = '', 3000)
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric'
  })
}

const getExpiryDays = (dateStr: string) => {
    const diff = new Date(dateStr).getTime() - new Date().getTime()
    return Math.floor(diff / (1000 * 60 * 60 * 24))
}

const getStatusClassLight = (status: string) => {
  switch (status) {
    case 'issued': return 'bg-emerald-50 text-emerald-600 border-emerald-100'
    case 'verifying': return 'bg-amber-50 text-amber-600 border-amber-100'
    case 'verified': return 'bg-blue-50 text-blue-600 border-blue-100'
    case 'failed': return 'bg-red-50 text-red-600 border-red-100'
    default: return 'bg-slate-50 text-slate-400 border-slate-100'
  }
}

const getStatusDotClassLight = (status: string) => {
    switch (status) {
        case 'issued': return 'bg-emerald-500'
        case 'verifying': return 'bg-amber-500'
        case 'verified': return 'bg-blue-500'
        case 'failed': return 'bg-red-500'
        default: return 'bg-slate-300'
    }
}

const getExpiryClassLight = (dateStr: string) => {
    const days = getExpiryDays(dateStr)
    if (days < 7) return 'text-red-600'
    if (days < 30) return 'text-amber-600'
    return 'text-blue-600'
}

const stepClassLight = (step: number, status: string) => {
    if (step === 1) {
        if (['verifying', 'verified', 'issued'].includes(status)) return 'bg-blue-50 border-blue-600 text-blue-600'
        return 'border-slate-200 text-slate-300'
    }
    if (step === 2) {
        if (['verified', 'issued'].includes(status)) return 'bg-blue-50 border-blue-600 text-blue-600'
        if (status === 'verifying') return 'bg-blue-600 border-blue-600 text-white shadow-lg shadow-blue-100 animate-pulse'
        return 'border-slate-200 text-slate-300'
    }
    if (step === 3) {
        if (status === 'issued') return 'bg-emerald-50 border-emerald-600 text-emerald-600'
        if (status === 'verified') return 'bg-blue-600 border-blue-600 text-white shadow-lg shadow-blue-100 animate-pulse'
        return 'border-slate-200 text-slate-300'
    }
}

onMounted(() => {
  authStore.init()
  fetchDomains()
})
</script>

<style scoped>
.animate-in {
    animation: techSlideUp 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes techSlideUp {
    from { opacity: 0; transform: translateY(12px); }
    to { opacity: 1; transform: translateY(0); }
}

pre::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
pre::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 10px;
}
pre::-webkit-scrollbar-track {
  background: transparent;
}
</style>
