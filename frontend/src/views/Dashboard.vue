<template>
  <div class="min-h-screen bg-gray-50 flex flex-col">
    <!-- Topbar -->
    <header class="bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between sticky top-0 z-10 shadow-sm">
      <div class="flex items-center gap-3">
        <div class="bg-blue-600 p-2 rounded-lg text-white">
          <ShieldCheck :size="24" />
        </div>
        <div>
          <h1 class="text-xl font-bold text-gray-900 tracking-tight">Let's Encrypt Manager</h1>
          <p class="text-xs text-gray-500">Automated Wildcard SSL Certificates</p>
        </div>
      </div>
      
      <div class="flex items-center gap-4">
        <button 
          @click="showAddDomainModal = true" 
          class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg flex items-center gap-2 transition-all font-medium text-sm shadow-sm hover:shadow-md"
        >
          <Plus :size="18" />
          <span>Add New Domain</span>
        </button>
        <div class="h-8 w-px bg-gray-200"></div>
        <button 
          @click="handleLogout" 
          class="p-2 text-gray-500 hover:bg-red-50 hover:text-red-600 rounded-lg transition-colors group"
          title="Logout"
        >
          <LogOut :size="20" class="group-hover:translate-x-1 transition-transform" />
        </button>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 p-8 max-w-7xl mx-auto w-full">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div class="p-card bg-white p-6 border border-gray-100 flex items-center gap-5">
            <div class="p-4 bg-blue-50 text-blue-600 rounded-2xl"><Globe :size="28" /></div>
            <div><p class="text-sm text-gray-500 font-medium">Total Domains</p><p class="text-2xl font-bold">{{ domains.length }}</p></div>
        </div>
        <div class="p-card bg-white p-6 border border-gray-100 flex items-center gap-5">
            <div class="p-4 bg-green-50 text-green-600 rounded-2xl"><CheckCircle :size="28" /></div>
            <div><p class="text-sm text-gray-500 font-medium">Issued Certificates</p><p class="text-2xl font-bold">{{ issuedCount }}</p></div>
        </div>
        <div class="p-card bg-white p-6 border border-gray-100 flex items-center gap-5">
            <div class="p-4 bg-amber-50 text-amber-600 rounded-2xl"><Clock :size="28" /></div>
            <div><p class="text-sm text-gray-500 font-medium">Expiring Soon</p><p class="text-2xl font-bold">{{ expiringSoonCount }}</p></div>
        </div>
      </div>

      <div class="p-card bg-white border border-gray-200 overflow-hidden">
        <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between bg-gray-50/50">
          <h2 class="font-bold text-gray-800 flex items-center gap-2">
            <LayoutDashboard :size="18" />
            Active Managed Domains
          </h2>
          <button @click="fetchDomains" class="p-2 text-gray-400 hover:text-blue-600 transition-colors" title="Refresh">
            <RefreshCw :size="16" :class="{'animate-spin': loading}" />
          </button>
        </div>
        
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="text-xs uppercase font-bold text-gray-400 bg-gray-50/50">
              <th class="px-6 py-3 border-b border-gray-100">Domain Name</th>
              <th class="px-6 py-3 border-b border-gray-100">Status</th>
              <th class="px-6 py-3 border-b border-gray-100">Expiry Date</th>
              <th class="px-6 py-3 border-b border-gray-100">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="d in domains" :key="d.domain" class="hover:bg-blue-50/30 transition-colors group">
              <td class="px-6 py-4">
                <div class="flex flex-col">
                  <span class="font-bold text-gray-800 tracking-tight">{{ d.domain }}</span>
                  <span class="text-xs text-gray-400">*.{{ d.domain }}</span>
                </div>
              </td>
              <td class="px-6 py-4">
                <span :class="getStatusClass(d.status)" class="px-2.5 py-1 rounded-full text-xs font-bold inline-flex items-center gap-1.5 uppercase tracking-wider">
                  <span class="w-1.5 h-1.5 rounded-full" :class="getStatusDotClass(d.status)"></span>
                  {{ d.status }}
                </span>
                <p v-if="d.error" class="text-xs text-red-500 mt-1 truncate max-w-xs" :title="d.error">{{ d.error }}</p>
              </td>
              <td class="px-6 py-4 text-sm text-gray-600">
                <div v-if="d.cert_expiry" class="flex flex-col">
                  <span class="font-medium">{{ formatDate(d.cert_expiry) }}</span>
                  <span class="text-xs" :class="getExpiryClass(d.cert_expiry)">{{ getExpiryDays(d.cert_expiry) }} days left</span>
                </div>
                <span v-else class="text-gray-300 italic font-mono">-</span>
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-2">
                  <button 
                    v-if="d.status === 'issued'"
                    @click="viewCert(d)" 
                    class="p-2 text-blue-600 hover:bg-blue-100 rounded-lg transition-all"
                    title="View Certificate"
                  >
                    <FileText :size="18" />
                  </button>
                  <button 
                    @click="manageDomain(d)" 
                    class="px-3 py-1.5 bg-gray-100 hover:bg-blue-600 hover:text-white text-gray-700 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5"
                  >
                    <Settings2 :size="14" />
                    MANAGE
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="domains.length === 0 && !loading">
              <td colspan="4" class="px-6 py-20 text-center text-gray-400">
                <div class="flex flex-col items-center gap-3 opacity-50">
                    <Inbox :size="48" stroke-width="1" />
                    <p class="text-lg">No domains found. Start by adding one!</p>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </main>

    <!-- Modal: Add Domain -->
    <div v-if="showAddDomainModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div class="bg-white rounded-2xl w-full max-w-md overflow-hidden shadow-2xl animate-in fade-in zoom-in duration-200">
        <div class="bg-blue-600 p-6 text-white flex items-center justify-between">
            <h3 class="text-xl font-bold flex items-center gap-2"><PlusCircle :size="20" /> Add Domain</h3>
            <button @click="showAddDomainModal = false" class="hover:bg-blue-500 p-1.5 rounded-lg transition-colors"><X :size="20" /></button>
        </div>
        <div class="p-8">
            <p class="text-sm text-gray-500 mb-6 italic">Enter the base domain (e.g., example.com). We will automatically generate certificates for both example.com and *.example.com.</p>
            <div class="space-y-4">
                <div class="flex flex-col gap-1.5">
                    <label class="text-sm font-bold text-gray-700">Domain Name</label>
                    <input 
                        v-model="newDomain" 
                        type="text" 
                        placeholder="example.com"
                        class="p-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none transition-all font-mono"
                        @keyup.enter="handleAddDomain"
                    />
                </div>
                <button 
                    @click="handleAddDomain" 
                    :disabled="!newDomain || loading"
                    class="w-full py-3 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-xl transition-all disabled:opacity-50 shadow-lg shadow-blue-200 flex items-center justify-center gap-2"
                >
                    <Loader2 v-if="loading" class="animate-spin" :size="18" />
                    Confirm Add Domain
                </button>
            </div>
        </div>
      </div>
    </div>

    <!-- Modal: Manage Domain (Workflow) -->
    <div v-if="selectedDomain" class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div class="bg-white rounded-2xl w-full max-w-2xl overflow-hidden shadow-2xl animate-in fade-in zoom-in duration-200 flex flex-col max-h-[90vh]">
        <div class="bg-gray-900 p-6 text-white flex items-center justify-between">
            <div>
                <h3 class="text-xl font-bold">{{ selectedDomain.domain }}</h3>
                <p class="text-xs text-gray-400">Current Status: <span class="uppercase text-blue-400 font-bold">{{ selectedDomain.status }}</span></p>
            </div>
            <button @click="selectedDomain = null" class="hover:bg-gray-800 p-1.5 rounded-lg transition-colors"><X :size="20" /></button>
        </div>
        
        <div class="flex-1 overflow-y-auto p-8 space-y-8">
            <!-- Step 1: Start/Retry Challenge -->
            <div class="relative pl-10">
                <div class="absolute left-0 top-0 w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm" :class="stepClass(1, selectedDomain.status)">1</div>
                <div>
                    <h4 class="font-bold text-gray-800 mb-2">DNS Challenge Authorization</h4>
                    <p class="text-sm text-gray-500 mb-4">Request a DNS challenge record from Let's Encrypt to prove domain ownership.</p>
                    <button 
                        v-if="['pending', 'failed', 'issued'].includes(selectedDomain.status)"
                        @click="handleGetChallenge" 
                        class="px-4 py-2 bg-blue-600 text-white text-sm font-bold rounded-lg hover:bg-blue-700 transition-all flex items-center gap-2"
                        :disabled="loading"
                    >
                        <Zap :size="16" /> Start New Order
                    </button>
                    
                    <div v-if="selectedDomain.challenge" class="mt-4 p-5 bg-gray-50 border border-gray-100 rounded-xl space-y-4 shadow-inner">
                        <div class="grid grid-cols-1 gap-3 text-sm">
                            <div>
                                <p class="text-xs font-bold text-gray-400 uppercase mb-1">Record Type</p>
                                <p class="font-mono bg-white px-2 py-1 rounded border border-gray-100 inline-block font-bold">TXT</p>
                            </div>
                            <div>
                                <p class="text-xs font-bold text-gray-400 uppercase mb-1">Host Name (Record Name)</p>
                                <div class="flex items-center gap-2">
                                    <code class="bg-white px-3 py-1.5 rounded border border-gray-200 text-blue-600 flex-1 truncate">{{ selectedDomain.challenge.record_name }}</code>
                                    <button @click="copy(selectedDomain.challenge.record_name)" class="p-1.5 hover:bg-gray-200 rounded transition-colors" title="Copy"><Copy :size="14" /></button>
                                </div>
                            </div>
                            <div>
                                <p class="text-xs font-bold text-gray-400 uppercase mb-1">Record Value (Key Auth)</p>
                                <div class="flex items-center gap-2">
                                    <code class="bg-white px-3 py-1.5 rounded border border-gray-200 text-amber-600 flex-1 truncate">{{ selectedDomain.challenge.txt_value }}</code>
                                    <button @click="copy(selectedDomain.challenge.txt_value)" class="p-1.5 hover:bg-gray-200 rounded transition-colors" title="Copy"><Copy :size="14" /></button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Step 2: Verify DNS -->
            <div class="relative pl-10" :class="{'opacity-50 pointer-events-none': !selectedDomain.challenge}">
                <div class="absolute left-0 top-0 w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm" :class="stepClass(2, selectedDomain.status)">2</div>
                <div>
                    <h4 class="font-bold text-gray-800 mb-2">DNS Verification</h4>
                    <p class="text-sm text-gray-500 mb-4">Check if the DNS record has propagated and is accessible by the server.</p>
                    <div class="flex items-center gap-3">
                        <button 
                            @click="handleVerifyDNS" 
                            class="px-4 py-2 bg-gray-900 text-white text-sm font-bold rounded-lg hover:bg-black transition-all flex items-center gap-2"
                            :disabled="loading"
                        >
                            <Search :size="16" /> Verify Record
                        </button>
                        <span v-if="verifyStatus" class="text-xs font-bold flex items-center gap-1.5" :class="verifyStatus.success ? 'text-green-600' : 'text-amber-600'">
                            <Check v-if="verifyStatus.success" :size="14" />
                            <AlertTriangle v-else :size="14" />
                            {{ verifyStatus.message }}
                        </span>
                    </div>
                </div>
            </div>

            <!-- Step 3: Issue Cert -->
            <div class="relative pl-10" :class="{'opacity-50 pointer-events-none': selectedDomain.status !== 'verified'}">
                <div class="absolute left-0 top-0 w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm" :class="stepClass(3, selectedDomain.status)">3</div>
                <div>
                    <h4 class="font-bold text-gray-800 mb-2">Issue SSL Certificate</h4>
                    <p class="text-sm text-gray-500 mb-4">Complete the ACME order and generate the certificate files.</p>
                    <button 
                        @click="handleIssueCert" 
                        class="px-4 py-2 bg-green-600 text-white text-sm font-bold rounded-lg hover:bg-green-700 transition-all flex items-center gap-2 shadow-lg shadow-green-100"
                        :disabled="loading || selectedDomain.status !== 'verified'"
                    >
                        <Award :size="16" /> Generate Certificate
                    </button>
                </div>
            </div>
        </div>
      </div>
    </div>

    <!-- Modal: View Certificate -->
    <div v-if="certData" class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div class="bg-white rounded-2xl w-full max-w-4xl overflow-hidden shadow-2xl animate-in fade-in zoom-in duration-200 flex flex-col max-h-[90vh]">
        <div class="bg-blue-600 p-6 text-white flex items-center justify-between">
            <div>
                <h3 class="text-xl font-bold flex items-center gap-2"><Award :size="20" /> {{ certData.domain }}</h3>
                <p class="text-xs text-blue-100">Expires: {{ formatDate(certData.cert_expiry) }}</p>
            </div>
            <button @click="certData = null" class="hover:bg-blue-500 p-1.5 rounded-lg transition-colors"><X :size="20" /></button>
        </div>
        
        <div class="flex-1 overflow-y-auto p-8 bg-gray-900 text-gray-300 font-mono text-xs">
            <div class="space-y-6">
                <div>
                    <div class="flex items-center justify-between mb-2">
                        <span class="text-white font-bold uppercase text-xs tracking-widest bg-gray-800 px-2 py-1 rounded">Fullchain Certificate (PEM)</span>
                        <button @click="copy(certData.fullchain_cer)" class="text-blue-400 hover:text-white flex items-center gap-1"><Copy :size="14" /> Copy</button>
                    </div>
                    <pre class="bg-black/30 p-4 rounded-xl border border-white/5 overflow-x-auto">{{ certData.fullchain_cer }}</pre>
                </div>
                <div>
                    <div class="flex items-center justify-between mb-2">
                        <span class="text-white font-bold uppercase text-xs tracking-widest bg-gray-800 px-2 py-1 rounded">Private Key (Unencrypted)</span>
                        <button @click="copy(certData.private_key)" class="text-amber-400 hover:text-white flex items-center gap-1"><Copy :size="14" /> Copy</button>
                    </div>
                    <pre class="bg-black/30 p-4 rounded-xl border border-white/5 overflow-x-auto">{{ certData.private_key }}</pre>
                </div>
            </div>
        </div>
      </div>
    </div>

    <!-- Toast-like feedback (simplified) -->
    <div v-if="toast" class="fixed bottom-8 left-1/2 -translate-x-1/2 bg-gray-900 text-white px-6 py-3 rounded-2xl shadow-2xl z-[100] flex items-center gap-3 animate-in fade-in slide-in-from-bottom-4">
        <CheckCircle class="text-green-400" :size="20" />
        <span class="font-medium">{{ toast }}</span>
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
  Copy, Search, Award, LayoutDashboard, Inbox, Loader2,
  Check, AlertTriangle
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
        showToast('Domain added successfully')
    } catch (err: any) {
        alert(err.response?.data?.error || 'Failed to add domain')
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
        // Refresh local ref
        selectedDomain.value = domains.value.find(d => d.domain === selectedDomain.value?.domain) || null
        showToast('DNS challenge generated')
    } catch (err: any) {
        alert(err.response?.data?.error || 'Failed to get challenge')
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
            ? { success: true, message: 'DNS record verified!' }
            : { success: false, message: 'Not propagated yet. Try again later.' }
        
        // Refresh local ref
        selectedDomain.value = domains.value.find(d => d.domain === selectedDomain.value?.domain) || null
    } catch (err: any) {
        verifyStatus.value = { success: false, message: err.response?.data?.error || 'Verification error' }
    } finally {
        loading.value = false
    }
}

const handleIssueCert = async () => {
    if (!selectedDomain.value) return
    loading.value = true
    try {
        await domainStore.issueCert(selectedDomain.value.domain)
        showToast('Certificate issued successfully!')
        selectedDomain.value = null
    } catch (err: any) {
        alert(err.response?.data?.error || 'Issuance failed')
    } finally {
        loading.value = false
    }
}

const viewCert = async (domain: DomainInfo) => {
    try {
        const data = await domainStore.getCert(domain.domain)
        certData.value = data
    } catch (err) {
        alert('Failed to fetch certificate details')
    }
}

const copy = (text: string) => {
    navigator.clipboard.writeText(text)
    showToast('Copied to clipboard')
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

const getStatusClass = (status: string) => {
  switch (status) {
    case 'issued': return 'bg-green-50 text-green-700 border border-green-200'
    case 'verifying': return 'bg-amber-50 text-amber-700 border border-amber-200'
    case 'verified': return 'bg-blue-50 text-blue-700 border border-blue-200'
    case 'failed': return 'bg-red-50 text-red-700 border border-red-200'
    default: return 'bg-gray-50 text-gray-700 border border-gray-200'
  }
}

const getStatusDotClass = (status: string) => {
    switch (status) {
        case 'issued': return 'bg-green-500'
        case 'verifying': return 'bg-amber-500'
        case 'verified': return 'bg-blue-500'
        case 'failed': return 'bg-red-500'
        default: return 'bg-gray-400'
    }
}

const getExpiryClass = (dateStr: string) => {
    const days = getExpiryDays(dateStr)
    if (days < 7) return 'text-red-500 font-bold'
    if (days < 30) return 'text-amber-500'
    return 'text-green-500'
}

const stepClass = (step: number, status: string) => {
    if (step === 1) {
        if (['verifying', 'verified', 'issued'].includes(status)) return 'bg-green-500 text-white'
        return 'bg-gray-200 text-gray-500'
    }
    if (step === 2) {
        if (['verified', 'issued'].includes(status)) return 'bg-green-500 text-white'
        if (status === 'verifying') return 'bg-blue-500 text-white shadow-lg shadow-blue-200'
        return 'bg-gray-200 text-gray-500'
    }
    if (step === 3) {
        if (status === 'issued') return 'bg-green-500 text-white'
        if (status === 'verified') return 'bg-blue-500 text-white shadow-lg shadow-blue-200'
        return 'bg-gray-200 text-gray-500'
    }
}

onMounted(() => {
  authStore.init()
  fetchDomains()
})
</script>

<style scoped>
.animate-in {
    animation: fadeIn 0.2s ease-out forwards;
}

@keyframes fadeIn {
    from { opacity: 0; transform: translateY(4px); }
    to { opacity: 1; transform: translateY(0); }
}

pre::-webkit-scrollbar {
  height: 4px;
}
pre::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,0.1);
  border-radius: 4px;
}
</style>
