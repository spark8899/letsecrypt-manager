<template>
  <div class="dashboard">
    <!-- Header -->
    <header class="header">
      <div class="header-left">
        <div class="logo">
          <ShieldCheck :size="24" />
        </div>
        <div>
          <h1>ACME Manager</h1>
          <p>SSL Certificate Management</p>
        </div>
      </div>
      <div class="header-right">
        <button class="btn-primary" @click="showAddDomainModal = true">
          <Plus :size="20" />
          Add Domain
        </button>
        <button class="btn-icon" @click="handleLogout" title="Logout">
          <LogOut :size="20" />
        </button>
      </div>
    </header>

    <!-- Main -->
    <main class="main">
      <!-- Stats -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon blue">
            <Globe :size="24" />
          </div>
          <div class="stat-info">
            <span class="stat-label">Total Domains</span>
            <span class="stat-value">{{ domains.length }}</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon green">
            <CheckCircle :size="24" />
          </div>
          <div class="stat-info">
            <span class="stat-label">Active Certificates</span>
            <span class="stat-value">{{ issuedCount }}</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon orange">
            <Clock :size="24" />
          </div>
          <div class="stat-info">
            <span class="stat-label">Expiring Soon</span>
            <span class="stat-value">{{ expiringSoonCount }}</span>
          </div>
        </div>
      </div>

      <!-- Domains Table -->
      <div class="card">
        <div class="card-header">
          <h2>
            <LayoutDashboard :size="20" />
            Managed Domains
          </h2>
          <button class="btn-icon" @click="fetchDomains" :class="{ spinning: loading }">
            <RefreshCw :size="18" />
          </button>
        </div>
        
        <div class="table-wrapper">
          <table>
            <thead>
              <tr>
                <th>Domain</th>
                <th>Status</th>
                <th>Expires</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="d in domains" :key="d.domain">
                <td>
                  <div class="domain-cell">
                    <span class="domain-name">{{ d.domain }}</span>
                    <span class="domain-wildcard">*.{{ d.domain }}</span>
                  </div>
                </td>
                <td>
                  <span :class="['status-badge', d.status]">
                    {{ d.status }}
                  </span>
                  <p v-if="d.error" class="error-text">{{ d.error }}</p>
                </td>
                <td>
                  <span v-if="d.cert_expiry" :class="getExpiryClass(d.cert_expiry)">
                    {{ formatDate(d.cert_expiry) }}
                    <br><span class="days-left">({{ getExpiryDays(d.cert_expiry) }} days)</span>
                  </span>
                  <span v-else class="no-date">—</span>
                </td>
                <td>
                  <div class="action-buttons">
                    <button v-if="d.status === 'issued'" @click="viewCert(d)" class="btn-icon-sm" title="View Certificate">
                      <FileText :size="16" />
                    </button>
                    <button @click="manageDomain(d)" class="btn-secondary-sm">
                      Manage
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="domains.length === 0 && !loading">
                <td colspan="4" class="empty-state">
                  <Inbox :size="48" />
                  <p>No domains added yet</p>
                  <p class="hint">Click "Add Domain" to get started</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </main>

    <!-- Add Domain Modal -->
    <div v-if="showAddDomainModal" class="modal-overlay" @click.self="showAddDomainModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3>Add New Domain</h3>
          <button class="btn-close" @click="showAddDomainModal = false">
            <X :size="20" />
          </button>
        </div>
        <div class="modal-body">
          <label>Domain Name</label>
          <input 
            v-model="newDomain" 
            type="text" 
            placeholder="example.com"
            @keyup.enter="handleAddDomain"
          />
          <p class="hint">Will provision wildcard certificate (*.domain + domain)</p>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showAddDomainModal = false">Cancel</button>
          <button class="btn-primary" @click="handleAddDomain" :disabled="!newDomain || loading">Add Domain</button>
        </div>
      </div>
    </div>

    <!-- Domain Management Modal -->
    <div v-if="selectedDomain" class="modal-overlay" @click.self="selectedDomain = null">
      <div class="modal modal-lg">
        <div class="modal-header">
          <div>
            <h3>{{ selectedDomain.domain }}</h3>
            <span :class="['status-badge', selectedDomain.status]">{{ selectedDomain.status }}</span>
          </div>
          <button class="btn-close" @click="selectedDomain = null">
            <X :size="20" />
          </button>
        </div>
        
        <div class="modal-body">
          <!-- Step 1 -->
          <div class="step" :class="{ done: selectedDomain.challenge }">
            <div class="step-num">1</div>
            <div class="step-content">
              <h4>DNS Challenge</h4>
              <p>Get DNS records to add for domain verification</p>
              <button 
                v-if="['pending', 'failed', 'issued'].includes(selectedDomain.status)"
                @click="handleGetChallenge" 
                :disabled="loading"
                class="btn-primary"
              >
                Get DNS Records
              </button>
              
              <div v-if="selectedDomain.challenge" class="dns-records">
                <div class="record">
                  <span class="record-label">Record Name</span>
                  <div class="record-value">
                    <code>{{ selectedDomain.challenge.record_name }}</code>
                    <button @click="copy(selectedDomain.challenge.record_name)" class="btn-icon-xs">
                      <Copy :size="14" />
                    </button>
                  </div>
                </div>
                <div class="record">
                  <span class="record-label">TXT Value</span>
                  <div class="record-value">
                    <code>{{ selectedDomain.challenge.txt_value.slice(0, 50) }}...</code>
                    <button @click="copy(selectedDomain.challenge.txt_value)" class="btn-icon-xs">
                      <Copy :size="14" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 2 -->
          <div class="step" :class="{ disabled: !selectedDomain.challenge }">
            <div class="step-num">2</div>
            <div class="step-content">
              <h4>Verify DNS</h4>
              <p>Check if DNS records have propagated</p>
              <div class="step-actions">
                <button @click="handleVerifyDNS" :disabled="loading || !selectedDomain.challenge" class="btn-dark">
                  Verify Now
                </button>
                <span v-if="verifyStatus" :class="verifyStatus.success ? 'text-success' : 'text-warning'">
                  {{ verifyStatus.message }}
                </span>
              </div>
            </div>
          </div>

          <!-- Step 3 -->
          <div class="step" :class="{ disabled: selectedDomain.status !== 'verified' }">
            <div class="step-num">3</div>
            <div class="step-content">
              <h4>Issue Certificate</h4>
              <p>Complete certificate issuance</p>
              <button 
                @click="handleIssueCert" 
                :disabled="loading || selectedDomain.status !== 'verified'"
                class="btn-success"
              >
                Issue Certificate
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- View Cert Modal -->
    <div v-if="certData" class="modal-overlay" @click.self="certData = null">
      <div class="modal modal-xl">
        <div class="modal-header">
          <div>
            <h3>
              <Award :size="20" />
              {{ certData.domain }}
            </h3>
            <p class="cert-expiry">Valid until: {{ formatDate(certData.cert_expiry) }}</p>
          </div>
          <button class="btn-close" @click="certData = null">
            <X :size="20" />
          </button>
        </div>
        
        <div class="modal-body cert-view">
          <div class="cert-section">
            <div class="cert-header">
              <span>Certificate (Full Chain)</span>
              <button @click="copy(certData.fullchain_cer)" class="btn-link">Copy</button>
            </div>
            <pre class="cert-content">{{ certData.fullchain_cer }}</pre>
          </div>
          <div class="cert-section">
            <div class="cert-header">
              <span>Private Key</span>
              <button @click="copy(certData.private_key)" class="btn-link warning">Copy</button>
            </div>
            <pre class="cert-content key">{{ certData.private_key }}</pre>
          </div>
        </div>
      </div>
    </div>

    <!-- Toast -->
    <div v-if="toast" class="toast">
      <CheckCircle :size="18" />
      {{ toast }}
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
  RefreshCw, FileText, LayoutDashboard, Inbox, Copy, 
  X, Award
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
    selectedDomain.value = domains.value.find(d => d.domain === selectedDomain.value?.domain) || null
    showToast('DNS records retrieved')
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
      ? { success: true, message: 'Verified!' }
      : { success: false, message: 'Not yet propagated' }
    selectedDomain.value = domains.value.find(d => d.domain === selectedDomain.value?.domain) || null
  } catch (err: any) {
    verifyStatus.value = { success: false, message: 'Error' }
  } finally {
    loading.value = false
  }
}

const handleIssueCert = async () => {
  if (!selectedDomain.value) return
  loading.value = true
  try {
    await domainStore.issueCert(selectedDomain.value.domain)
    showToast('Certificate issued!')
    selectedDomain.value = null
  } catch (err: any) {
    alert(err.response?.data?.error || 'Issue failed')
  } finally {
    loading.value = false
  }
}

const viewCert = async (domain: DomainInfo) => {
  try {
    const data = await domainStore.getCert(domain.domain)
    certData.value = data
  } catch (err) {
    alert('Failed to fetch certificate')
  }
}

const copy = (text: string) => {
  navigator.clipboard.writeText(text)
  showToast('Copied to clipboard')
}

const showToast = (msg: string) => {
  toast.value = msg
  setTimeout(() => toast.value = '', 2500)
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

const getExpiryClass = (dateStr: string) => {
  const days = getExpiryDays(dateStr)
  if (days < 7) return 'date-danger'
  if (days < 30) return 'date-warning'
  return 'date-normal'
}

onMounted(() => {
  authStore.init()
  fetchDomains()
})
</script>

<style scoped>
/* Reset & Base */
* {
  box-sizing: border-box;
}

.dashboard {
  min-height: 100vh;
  background: #f1f5f9;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* Header */
.header {
  background: white;
  border-bottom: 1px solid #e2e8f0;
  padding: 16px 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.header-left h1 {
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.header-left p {
  font-size: 12px;
  color: #64748b;
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* Buttons */
.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
  padding: 10px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #e2e8f0;
}

.btn-secondary-sm {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
  padding: 6px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary-sm:hover {
  background: #e2e8f0;
}

.btn-dark {
  background: #1e293b;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-dark:hover:not(:disabled) {
  background: #0f172a;
}

.btn-dark:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-success {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(34, 197, 94, 0.3);
}

.btn-success:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(34, 197, 94, 0.4);
}

.btn-success:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  background: transparent;
  border: none;
  padding: 10px;
  border-radius: 10px;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: #f1f5f9;
  color: #ef4444;
}

.btn-icon-sm {
  background: #eff6ff;
  border: none;
  padding: 8px;
  border-radius: 8px;
  color: #3b82f6;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon-sm:hover {
  background: #dbeafe;
}

.btn-icon-xs {
  background: transparent;
  border: none;
  padding: 4px;
  color: #94a3b8;
  cursor: pointer;
}

.btn-icon-xs:hover {
  color: #3b82f6;
}

.btn-close {
  background: transparent;
  border: none;
  padding: 8px;
  border-radius: 8px;
  color: #64748b;
  cursor: pointer;
}

.btn-close:hover {
  background: #f1f5f9;
}

.btn-link {
  background: none;
  border: none;
  color: #3b82f6;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.btn-link:hover {
  text-decoration: underline;
}

.btn-link.warning {
  color: #f59e0b;
}

/* Main */
.main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 32px;
}

/* Stats */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon.blue {
  background: #eff6ff;
  color: #3b82f6;
}

.stat-icon.green {
  background: #f0fdf4;
  color: #22c55e;
}

.stat-icon.orange {
  background: #fffbeb;
  color: #f59e0b;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-label {
  font-size: 14px;
  color: #64748b;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #1e293b;
}

/* Card */
.card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header h2 {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Table */
.table-wrapper {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  text-align: left;
  padding: 14px 24px;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  background: #f8fafc;
}

td {
  padding: 16px 24px;
  border-top: 1px solid #f1f5f9;
}

.domain-cell {
  display: flex;
  flex-direction: column;
}

.domain-name {
  font-weight: 600;
  color: #1e293b;
}

.domain-wildcard {
  font-size: 12px;
  color: #94a3b8;
}

.status-badge {
  display: inline-flex;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  text-transform: capitalize;
}

.status-badge.issued {
  background: #dcfce7;
  color: #16a34a;
}

.status-badge.verifying {
  background: #fef3c7;
  color: #d97706;
}

.status-badge.verified {
  background: #dbeafe;
  color: #2563eb;
}

.status-badge.failed {
  background: #fee2e2;
  color: #dc2626;
}

.status-badge.pending {
  background: #f1f5f9;
  color: #64748b;
}

.error-text {
  font-size: 12px;
  color: #dc2626;
  margin: 4px 0 0;
}

.date-danger {
  color: #dc2626;
  font-weight: 500;
}

.date-warning {
  color: #d97706;
  font-weight: 500;
}

.date-normal {
  color: #1e293b;
}

.days-left {
  font-size: 12px;
  color: #94a3b8;
}

.no-date {
  color: #cbd5e1;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: flex-end;
}

.empty-state {
  text-align: center;
  padding: 60px 24px !important;
  color: #94a3b8;
}

.empty-state p {
  margin: 8px 0 0;
}

.empty-state .hint {
  font-size: 13px;
  color: #cbd5e1;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: 20px;
}

.modal {
  background: white;
  border-radius: 16px;
  width: 100%;
  max-width: 440px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.modal-lg {
  max-width: 560px;
}

.modal-xl {
  max-width: 800px;
}

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.modal-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-header p {
  font-size: 13px;
  color: #64748b;
  margin: 4px 0 0;
}

.modal-body {
  padding: 24px;
}

.modal-body label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
}

.modal-body input {
  width: 100%;
  height: 48px;
  padding: 0 14px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 15px;
  transition: all 0.2s;
}

.modal-body input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.modal-body .hint {
  font-size: 13px;
  color: #94a3b8;
  margin-top: 8px;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid #f1f5f9;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* Steps */
.step {
  display: flex;
  gap: 16px;
  padding: 20px;
  background: #f8fafc;
  border-radius: 12px;
  margin-bottom: 16px;
}

.step.disabled {
  opacity: 0.5;
  pointer-events: none;
}

.step.done {
  background: #f0fdf4;
}

.step-num {
  width: 36px;
  height: 36px;
  background: #e2e8f0;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: #64748b;
  flex-shrink: 0;
}

.step.done .step-num {
  background: #22c55e;
  color: white;
}

.step-content h4 {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 4px;
}

.step-content p {
  font-size: 13px;
  color: #64748b;
  margin: 0 0 12px;
}

.step-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dns-records {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.record {
  margin-bottom: 12px;
}

.record-label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 6px;
}

.record-value {
  display: flex;
  align-items: center;
  gap: 8px;
}

.record-value code {
  flex: 1;
  background: white;
  padding: 10px 12px;
  border-radius: 8px;
  font-size: 13px;
  color: #3b82f6;
  overflow-x: auto;
}

.text-success {
  color: #22c55e;
  font-size: 14px;
  font-weight: 500;
}

.text-warning {
  color: #f59e0b;
  font-size: 14px;
  font-weight: 500;
}

/* Cert View */
.cert-view {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.cert-section {
  background: #f8fafc;
  border-radius: 12px;
  overflow: hidden;
}

.cert-header {
  padding: 14px 16px;
  background: white;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.cert-content {
  padding: 16px;
  margin: 0;
  font-size: 11px;
  font-family: monospace;
  color: #64748b;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 250px;
  overflow-y: auto;
}

.cert-content.key {
  color: #d97706;
}

/* Toast */
.toast {
  position: fixed;
  bottom: 24px;
  right: 24px;
  background: #1e293b;
  color: white;
  padding: 14px 20px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 500;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  z-index: 300;
  animation: slideUp 0.3s ease;
}

.toast svg {
  color: #22c55e;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive */
@media (max-width: 768px) {
  .header {
    padding: 12px 16px;
  }
  
  .main {
    padding: 16px;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }
  
  .cert-view {
    grid-template-columns: 1fr;
  }
}
</style>
