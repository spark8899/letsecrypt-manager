import { defineStore } from 'pinia'
import axios from 'axios'

export interface Challenge {
  type: string
  record_name: string
  txt_value: string
  cname_target: string
  instructions: string
}

export interface DomainInfo {
  domain: string
  status: string
  cert_expiry: string | null
  challenge: Challenge | null
  error: string | null
}

export const useDomainStore = defineStore('domain', {
  state: () => ({
    domains: [] as DomainInfo[],
    loading: false
  }),
  actions: {
    async fetchDomains() {
      this.loading = true
      try {
        const res = await axios.get('/api/domains')
        this.domains = res.data.domains
      } finally {
        this.loading = false
      }
    },
    async addDomain(domain: string) {
      await axios.post('/api/domains', { domain })
      await this.fetchDomains()
    },
    async getChallenge(domain: string) {
      const res = await axios.post(`/api/domains/${domain}/dns-challenge`)
      await this.fetchDomains()
      return res.data.challenge as Challenge
    },
    async verifyDNS(domain: string) {
      const res = await axios.get(`/api/domains/${domain}/dns-verify`)
      await this.fetchDomains()
      return res.data.verified as boolean
    },
    async issueCert(domain: string) {
      await axios.post(`/api/domains/${domain}/issue`)
      await this.fetchDomains()
    },
    async getCert(domain: string) {
        const res = await axios.get(`/api/domains/${domain}/cert`)
        return res.data
    }
  }
})
